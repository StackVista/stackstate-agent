import json
import hashlib
import logging
import os

from testinfra.host import Host

from .models import *


class CLIv1:
    def __init__(self, host: Host, log: logging.Logger = logging, cache_enabled: bool = False):
        self.log = log
        self.host = host
        self.cache_enabled = cache_enabled

    def topic_api(self, topic, limit=1000) -> dict:
        executed = self.host.run(f"sts-cli topic show {topic} -l {limit}")
        self.log.info(f"executed sts-cli topic show for topic {topic}: {executed}")
        json_data = json.loads(executed.stdout)

        return json_data

    def telemetry(self, component_ids):
        if len(component_ids) == 0:
            return []
        result = self.script(CLIv1.telemetry_script(component_ids))
        series = {}

    def topology(self, query: str) -> TopologyResult:
        result = self.script(CLIv1.topology_script(query))
        component_type_map = {}
        for comp_type in result['component_types']:
            component_type_map[comp_type['id']] = comp_type
        relation_type_map = {}
        for rel_type in result['relation_types']:
            relation_type_map[rel_type['id']] = rel_type
        return TopologyResult(
            components=[
                ComponentWrapper({**comp, **{"type": component_type_map[comp['type']]['name']}})
                for comp in result['components']
            ],
            relations=[
                RelationWrapper({**rel, **{"type": relation_type_map[rel['type']]['name']}})
                for rel in result['relations']
            ],
        )

    def script(self, fullquery) -> dict:
        log = self.log
        log.info(f"querying StackState with CLIv1: %s, cache: %s", fullquery, self.cache_enabled)
        cachefile = hashlib.sha1(fullquery.encode('utf-8')).hexdigest() + '.json'
        if self.cache_enabled:
            try:
                with open(cachefile, 'r') as f:
                    log.warning(f"using cached result from %s", cachefile)
                    return json.load(f)['result']
            except IOError:
                pass

        # Write topology query into the expected file
        home = os.path.expanduser("~")
        with open(f"{home}/sts-query.stsl", 'w') as script_query:
            # do stuff with temp file
            script_query.write(fullquery)

        # Execute the query
        executed = self.host.run(f"{home}/sts-query.sh")
        log.info(f"executed STSL script: {executed}")
        json_data = json.loads(executed.stdout)['result']
        if self.cache_enabled:
            with open(cachefile, 'w') as f:
                f.write(executed["stdout"])
        return json_data

    @staticmethod
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

    @staticmethod
    def topology_script(query):
        escaped_query = query.replace("\\", "\\\\").replace('"', '\\"').replace("'", "\\'")
        return """
def emptyPromise() {
  Async.sequence([])
}

Topology.query('__QUERY__')
  .then { result ->
    components = result.queryResults[0].result.components
    relations = result.queryResults[0].result.relations
    compTypeIDs = components.collect { c -> c.type }.unique()
    compTypesPromise = (compTypeIDs.size() > 0) ? Graph.query{ it.V(compTypeIDs) } : emptyPromise()
    relTypeIDs = relations.collect { c -> c.type }.unique()
    relTypesPromise = (relTypeIDs.size() > 0) ? Graph.query{ it.V(relTypeIDs) } : emptyPromise()
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
