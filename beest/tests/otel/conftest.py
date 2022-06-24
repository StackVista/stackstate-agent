import hashlib
import json
import os
import pytest
import tempfile as tfile


def telemetry_script(ids: list, start="-10m"):
    return """
Topology.query("id in (__IDS__)")
  .fullComponents()
  .metricStreams()
  .thenStream {
     Telemetry
       .multiQuery(it)
       .aggregation("mean", "1m")
       .start("__START__")
  }
""" \
        .replace('__IDS__', "'" + "','".join(map(str, ids)) + "'") \
        .replace('__START__', start)


def topology_script(query):
    escaped_query = query.replace("'", "\\'")
    return """
Topology.query('__QUERY__')
  .then { result ->
    components = result.queryResults[0].result.components
    relations = result.queryResults[0].result.relations
    compTypeIDs = components.collect { c -> c.type }.unique()
    compTypesPromise = Graph.query{ it.V(compTypeIDs) }
    relTypeIDs = relations.collect { c -> c.type }.unique()
    relTypesPromise = Graph.query{ it.V(relTypeIDs) }
    compTypesPromise.then { compTypes ->
      relTypesPromise.then { relTypes ->
        [
         components: components,
         relations: relations,
         component_types: compTypes,
         relation_types: relTypes
        ]
      }
    }
  }
""".replace('__QUERY__', escaped_query)


class CLIv1:
    def __init__(self, host):
        self.host = host

    def telemetry(self, component_ids):
        if len(component_ids) == 0:
            return []
        result = self.script(telemetry_script(component_ids))
        series = {}

    def topology(self, query: str):
        result = self.script(topology_script(query))
        component_type_map = {}
        for comp_type in result['component_types']:
            component_type_map[comp_type['id']] = comp_type
        relation_type_map = {}
        for rel_type in result['relation_types']:
            relation_type_map[rel_type['id']] = rel_type
        return {
            'components': [
                {**comp, **{"type": component_type_map[comp['type']]['name']}}
                for comp in result['components']
            ],
            'relations': [
                {**rel, **{"type": relation_type_map[rel['type']]['name']}}
                for rel in result['relations']
            ],
        }

    def script(self, fullquery):
        cachefile = hashlib.sha1(fullquery.encode('utf-8')).hexdigest() + '.json'
        try:
            with open(cachefile, 'r') as f:
                return json.load(f)['result']
        except IOError:
            pass

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
            transfer_result = self.host.ansible("kubernetes.core.k8s_cp", f"{ctx} {ns} {pod} {local_path} {remote_path}")
            print(f"[cli] transfer result: {transfer_result}")
        finally:
            os.remove(path)

        # Execute the query
        command = f"command=\"bash query.sh\""
        executed = self.host.ansible("kubernetes.core.k8s_exec", f"{ctx} {ns} {pod} {command}")
        print(f"[cli] executed query: {executed}")
        json_data = json.loads(executed["stdout"])['result']
        with open(cachefile, 'w') as f:
            f.write(executed["stdout"])
        return json_data


@pytest.fixture
def cliv1(host) -> CLIv1:
    return CLIv1(host)
