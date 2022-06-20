import json
import os
import pytest
import tempfile as tfile


@pytest.fixture
def cliv1(host):
    class CLIv1:
        def components(self, query):
            return self.script(f"Topology.query(\"{query}\").components()\n")

        def script(self, fullquery):
            ctx = "context={{ kubecontext }}"
            ns = "namespace={{ namespace }}"
            pod = "pod=stackstate-cli"

            # Transfer query to a file inside the cli pod
            fd, path = tfile.mkstemp()
            try:
                # Write topology query to a temporary file first
                with os.fdopen(fd, 'w') as tmp_topo_query:
                    # do stuff with temp file
                    tmp_topo_query.write(fullquery)

                local_path = f"local_path=\"{path}\""
                remote_path = "remote_path=\"/query.stql\""
                # then transfer it
                transferred = host.ansible("kubernetes.core.k8s_cp", f"{ctx} {ns} {pod} {local_path} {remote_path}")["result"]
                print(f"[cli] transferred query: {transferred}")
            finally:
                os.remove(path)

            # Execute the query
            command = f"command=\"bash query.sh\""
            executed = host.ansible("kubernetes.core.k8s_exec", f"{ctx} {ns} {pod} {command}")["stdout"]
            print(f"[cli] executed query: {executed}")
            return json.loads(executed)

    return CLIv1()
