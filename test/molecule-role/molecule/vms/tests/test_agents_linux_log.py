import os
import re
import util
from testinfra.utils.ansible_runner import AnsibleRunner

testinfra_hosts = AnsibleRunner(os.environ['MOLECULE_INVENTORY_FILE']).get_hosts('agent_linux_vm')


def _get_log(host, log_name, agent_log_path):
    agent_log = host.file(agent_log_path).content_string
    with open("./{}.log".format(log_name), 'wb') as f:
        f.write(agent_log.encode('utf-8'))
    return agent_log


def test_stackstate_agent_log(host, hostname):
    agent_log_path = "/var/log/stackstate-agent/agent.log"

    # Check for presence of success
    def wait_for_check_successes():
        agent_log = _get_log(host, "{}-{}".format(hostname, "agent"), agent_log_path)
        assert re.search("Successfully posted payload to.*stsAgent", agent_log)

    util.wait_until(wait_for_check_successes, 60, 3)

    ignored_errors_regex = [
        # TODO: Collecting processes snap -> Will be addressed with STAC-3531
        "Error code \"400 Bad Request\" received while "
        "sending transaction to \"https://.*/stsAgent/intake/.*"
        "Failed to deserialize JSON on fields: , "
        "with message: Object is missing required member \'internalHostname\'",
        "net/ntp.go.*There was an error querying the ntp host",
        "Service listener factory .* already registered",  # double register of the ECS service listener factory
        "workloadmeta collector .* could not start. error",  # ignoring INFO log containing the word error
        "For verbose messaging see aws.Config.CredentialsChainVerboseErrors",  # ignoring => contains the word error
        "unable to get tags from gce and cache is empty"  # ignoring when tags lookup don't work for gce collector
    ]

    # Check for errors
    agent_log = _get_log(host, "{}-{}".format(hostname, "agent"), agent_log_path)
    for line in agent_log.splitlines():
        ignored = False
        for ignored_error in ignored_errors_regex:
            if len(re.findall(ignored_error, line, re.DOTALL)) > 0:
                ignored = True
        if "0.datadog.pool.ntp.org" in line:
            print("Datadog default host still exist for ntp in line {}".format(line))
            ignored = False
        if ignored:
            continue

        print("Considering: %s" % line)
        assert not re.search("error", line, re.IGNORECASE)


def test_stackstate_process_agent_no_log_errors(host, hostname):
    process_agent_log_path = "/var/log/stackstate-agent/process-agent.log"

    # Check for presence of success
    def wait_for_check_successes():
        process_agent_log = _get_log(host, "{}-{}".format(hostname, "process-agent"), process_agent_log_path)
        assert re.search("Finished check #1", process_agent_log)
        if hostname != "agent-centos":
            assert re.search("starting network tracer locally", process_agent_log)
        if hostname == "agent-ubuntu":
            assert re.search("Setting process api endpoint from config using `sts_url`", process_agent_log)

    util.wait_until(wait_for_check_successes, 30, 3)

    ignored_errors_regex = [
        "failed to create network tracer: failed to init module: error guessing offsets: error initializing tcptracer_status map: unable to update element: .*. Retrying...",
        "failed to create network tracer: error while loading .*",
        "- Caught signal 'terminated'; terminating.",
        "- Caught signal continued; continuing/ignoring",
        "\(FileName\(\) error: error during runtime.Caller:-1\)"
    ]

    offset_guessing = "Offset guessing was completed successfully"
    found_offset_guessing = False
    # Check for errors
    process_agent_log = _get_log(host, "{}-{}".format(hostname, "process-agent"), process_agent_log_path)
    for line in process_agent_log.splitlines():
        # Ignore offset guessing for centos
        if hostname == "agent-centos" or re.search(offset_guessing, line):
            found_offset_guessing = True
        ignored = False
        for ignored_error in ignored_errors_regex:
            if len(re.findall(ignored_error, line, re.DOTALL)) > 0:
                ignored = True
        if ignored:
            continue
        print("Considering: %s" % line)
        assert not re.search("error", line, re.IGNORECASE)

    assert found_offset_guessing, "Process agent could not guess offset"


def test_stackstate_trace_agent_no_log_errors(host, hostname):
    trace_agent_log_path = "/var/log/stackstate-agent/trace-agent.log"

    # Check for presence of success
    def wait_for_check_successes():
        trace_agent_log = _get_log(host, "{}-{}".format(hostname, "trace-agent"), trace_agent_log_path)
        assert re.search("Trace agent running on host", trace_agent_log)
        assert re.search("No data received", trace_agent_log)

    util.wait_until(wait_for_check_successes, 30, 3)

    # Check for errors
    ignored_errors_regex = [
        "workloadmeta collector .* could not start. error",  # ignoring INFO log containing the word error
    ]
    trace_agent_log = _get_log(host, "{}-{}".format(hostname, "trace-agent"), trace_agent_log_path)
    for line in trace_agent_log.splitlines():
        ignored = False
        for ignored_error in ignored_errors_regex:
            if len(re.findall(ignored_error, line, re.DOTALL)) > 0:
                ignored = True
        if "0.datadog.pool.ntp.org" in line:
            print("Datadog default host still exist for ntp in line {}".format(line))
            ignored = False
        if ignored:
            continue

        print("Considering: %s" % line)
        assert not re.search("error", line, re.IGNORECASE)
