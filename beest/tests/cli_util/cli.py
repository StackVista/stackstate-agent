import json
import subprocess

# Topology script documentation: https://docs.stackstate.com/develop/reference/scripting/script-apis/topology

class CLIUtil(object):

    def query_by_name(self, name):
        script = f"""
        Topology.query('name = "{name}"')
            .components()
        """
        return self.run_script(script)

    def run_script(self, script):
        stdout = subprocess.run(["sts", "script", "run", "--json", "--script", script], capture_output=True).stdout
        return json.loads(stdout)["result"]["value"]


client = CLIUtil()
print(json.dumps(client.query_by_name("cluster-agent")[0], indent=4))
