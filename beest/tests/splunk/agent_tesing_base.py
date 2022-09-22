import paramiko


class AgentTestingBase:
    client = None

    def establish_connection(self, hostname: str, username: str, key_file_path: str = None, password: str = None):
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

    @staticmethod
    def start_agent_on_host():
        print("Hello World")

