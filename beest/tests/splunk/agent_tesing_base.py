from typing import Callable

import paramiko
import logging
import time
import json

from pathlib import Path
from paramiko import SSHClient


class AgentTestingBase:
    client: SSHClient = None
    open_connection: bool = False

    def __init__(self, ansible_var, hostname: str, username: str, key_file_path: str,  password: str = None):
        self.ansible_var = ansible_var
        self.establish_connection(hostname, username, key_file_path, password)

    def establish_connection(self,
                             hostname: str,
                             username: str,
                             key_file_path: str = None,
                             password: str = None):
        if self.open_connection:
            print("Connection is already established.")
            return

        self.client = paramiko.SSHClient()
        self.client.set_missing_host_key_policy(paramiko.AutoAddPolicy())
        if password is not None:
            self.client.connect(hostname=hostname,
                                username=username,
                                password=password)
        else:
            self.client.connect(hostname=hostname,
                                username=username,
                                key_filename=key_file_path)

        self.open_connection = True

    def close_connection(self):
        self.client.close()
        self.open_connection = False

    # Exposes the capabilities to Start the StackState agent on a external machine
    # This is achieved by SSH into the machine and executing a shell command
    def start_agent_on_host(self):
        logging.info("Starting the StackState-Agent Service")

        # Execute the shell command to start the agent service
        self.client.exec_command('sudo service stackstate-agent start')

        # Give the agent 10 seconds to properly start up
        time.sleep(10)

        # Test if the agent is running
        _, stdout, _ = self.client.exec_command('sudo systemctl is-active --quiet stackstate-agent '
                                                '&& echo OK')

        if stdout.read() != b'OK\n':
            raise ProcessLookupError("The StackState Agent is not running on the host machine ...")
        else:
            logging.info("Success, The StackState Agent is running")

    # Exposes the capabilities to Stop the StackState agent on a external machine
    # This is achieved by SSH into the machine and executing a shell command
    def stop_agent_on_host(self):
        logging.info("Stopping the StackState-Agent Service")

        # Execute the shell command to start the agent service
        self.client.exec_command('sudo service stackstate-agent stop')

        # Give the agent 5 seconds to properly shutdown
        time.sleep(5)

        # Test if the agent is running
        _, stdout, _ = self.client.exec_command('sudo systemctl is-active --quiet stackstate-agent '
                                                '&& echo OK')

        if stdout.read() == b'OK\n':
            raise ProcessLookupError("The StackState Agent is still running after attempting to stop it ...")
        else:
            logging.info("Success, The StackState Agent is stopped")

    # Remove Agent v2 run cache
    def remove_agent_run_cache(self):
        logging.info("Removing StackState-Agent v2 Run Cache")

        # Execute the shell command to start the agent service
        self.client.exec_command('sudo rm -rf /opt/stackstate-agent/run')

    # Convert Agent v1 to v2 cache
    def convert_agent_v1_run_cache_to_v2(self):
        logging.info("Converting StackState Agent v1 Cache to v2 format")

        # Execute the shell command to convert the run cache files
        self.client.exec_command('sudo ./home/ubuntu/agent-v1-to-v2-pickle-conversion/run.sh')

    def cache_delete_stateful_check_state(self):
        logging.info("Clearing the Agent Stateful Cache ...")
        self.client.exec_command('cd /opt/stackstate-agent/run/ && sudo find . -type f -name "*_check_state" '
                                 '! -name "*_transactional_check_state" -delete')

    def cache_delete_transactional_check_state(self):
        logging.info("Clearing the Agent Transactional Cache ...")
        self.client.exec_command('cd /opt/stackstate-agent/run/ && sudo find . '
                                 '-type f -name "*_transactional_check_state" -delete')

    def cache_delete_event_state(self):
        logging.info("Clearing the Agent Event Cache ...")
        self.client.exec_command('cd /opt/stackstate-agent/run/ && sudo find . -type f -name "*_event" -delete')

    def cache_clear(self):
        logging.info("Clearing the Agent cache ...")
        self.cache_delete_stateful_check_state()
        self.cache_delete_transactional_check_state()
        self.cache_delete_event_state()
        logging.info("Agent cache cleared.")

    def block_routing_to_sts_instance(self):
        def block_port(port: int, type: str):
            logging.info(f"Block StackState Agent traffic on the following port: "
                         f"sudo iptables -A {type} -p tcp --dport {port} -j DROP")
            self.client.exec_command(f'sudo iptables -A {type} -p tcp --dport {port} -j DROP')

        def block_host(host: str, type: str):
            logging.info(f"Block StackState Agent traffic on the following host: "
                         f"sudo iptables -A {type} -j DROP -d {host}")
            self.client.exec_command(f'sudo iptables -A {type} -j DROP -d {host}')

        # Retrieve the url the agent sends data to
        sts_instance_url = self.ansible_var("sts_url")

        # Cleanup the URL to use it in a iptable
        sts_instance_url = sts_instance_url.replace('http://', '').replace('https://', '')

        # Block Port
        block_port(7078, "OUTPUT")
        block_port(7078, "FORWARD")
        block_port(7078, "INPUT")

        # Block Port
        block_port(7077, "OUTPUT")
        block_port(7077, "FORWARD")
        block_port(7077, "INPUT")

        # Block Host
        block_host(sts_instance_url, "OUTPUT")
        block_host(sts_instance_url, "FORWARD")
        block_host(sts_instance_url, "INPUT")

        # Apply the block rules
        self.client.exec_command(f'sudo iptables-save')

    def allow_routing_to_sts_instance(self):
        logging.info(f"Open up all iptables routes ...")
        self.client.exec_command(f'sudo iptables -F')
        self.client.exec_command(f'sudo iptables -P INPUT ACCEPT')
        self.client.exec_command(f'sudo iptables -P OUTPUT ACCEPT')
        self.client.exec_command(f'sudo iptables -P FORWARD ACCEPT')
        self.client.exec_command(f'sudo iptables-save')

    def transactional_run_cycle_test(self,
                                     func_before_blocking_routes: Callable[[], any],
                                     func_after_blocking_routes: Callable[[], any],
                                     rerun_func_unblocking_blocked_routes: Callable[[], any],
                                     wait_after_blocking_routes: int = 60):
        # Open up the routing from the agent to sts instance
        self.allow_routing_to_sts_instance()
        # Make sure the agent is running on the host
        self.start_agent_on_host()
        # Run a function before we block the routes
        func_before_blocking_routes()
        # Now let's block the communication to the sts instance
        self.block_routing_to_sts_instance()
        # Sleep for a bit to make sure the packets is frozen
        time.sleep(30)
        # Now lets run a function while the routing is blocked
        func_after_blocking_routes()
        # Let's wait two minutes to allow a gap between testing what the agent pulled
        time.sleep(wait_after_blocking_routes)
        # Now let's allow the communication to the sts instance again
        self.allow_routing_to_sts_instance()
        # Now lets run a function after the routing was unblocked
        rerun_func_unblocking_blocked_routes()

    def stateful_state_run_cycle_test(self,
                                      func_before_agent_stop: Callable[[], any],
                                      func_after_agent_stop: Callable[[], any],
                                      func_after_agent_startup: Callable[[], any],
                                      time_between_stop_and_start: int = 30):
        # Lets first stop the Agent
        self.stop_agent_on_host()
        # Now let's remove old cache data before starting the agent up again
        self.cache_clear()
        # Make sure the agent is first running before attempting anything
        self.start_agent_on_host()
        # Run a function before stopping the agent
        func_before_agent_stop()
        # Stop the agent
        self.stop_agent_on_host()
        # Run a function  while the agent is stopped
        func_after_agent_stop()
        # Wait some time after the agent has stopped to make the gap bigger between data and next start
        time.sleep(time_between_stop_and_start)
        # Start the agent up again
        self.start_agent_on_host()
        # Run a function after the agent has started up again
        func_after_agent_startup()

    def get_current_time_on_agent_machine(self) -> str:
        _, stdout, _ = self.client.exec_command(f'date +"%Y-%m-%d %H:%M:%S"')
        host_response = stdout.read().decode().rstrip()
        return host_response

    def dump_logs(self, request, from_date_time):
        print(f"Dumping all StackState Agent logs from the time: {from_date_time}")
        print(f"Executing the following command for the agent logs: "
              f"journalctl -u stackstate-agent.service --since '{from_date_time}'")

        _, stdout, _ = self.client.exec_command(f'journalctl -u stackstate-agent.service '
                                                f'--since "{from_date_time}"')

        agent_docker_logs = stdout.read().decode()

        data_dump_filename = "{}-{}-agent_dump.log".format(
            Path(str(request.node.fspath)).stem,
            request.node.originalname,
        )

        # Dump the data before starting to play around with it like trimming it
        with open(data_dump_filename, "w") as outfile:
            outfile.write(agent_docker_logs)




