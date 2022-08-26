import json
import hashlib
import inspect
import logging
import os
from typing import Callable

from testinfra.host import Host

from .models import *


class CLIv1:
    def __init__(self, host: Host, log: logging.Logger = logging, cache_enabled: bool = False):
        self.log = log
        self.host = host
        self.cache_enabled = cache_enabled

    def telemetry(self, component_ids, alias: str = None):
        if len(component_ids) == 0:
            return []
        fullquery = self._telemetry_script(component_ids)
        if alias is None:
            alias = self._query_digest(fullquery)

        result = self.script(fullquery, alias)
        series = {}

    def topology(self, query: str, alias: str = None) -> TopologyResult:
        fullquery = self._topology_script(query)
        if alias is None:
            alias = self._query_digest(fullquery)

        result = self.script(fullquery, alias)

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

    @staticmethod
    def _query_digest(q: str) -> str: return hashlib.sha1(q.encode('utf-8')).hexdigest()

    def script(self, fullquery: str, alias) -> dict:
        log = self.log
        log.info(f"Querying StackState Script API with CLIv1: {fullquery}, cache: {self.cache_enabled}")

        def query():
            # Write topology query into the expected file
            home = os.path.expanduser("~")
            with open(f"{home}/sts-query.stsl", 'w') as script_query:
                # do stuff with temp file
                script_query.write(fullquery)
            # Execute the query
            executed = self.host.run(f"{home}/sts-query.sh")
            log.info(f"STSL script exit status: {executed.exit_status}")
            return executed.stdout

        return self._cached_json(query, alias)["result"]

    @staticmethod
    def _telemetry_script(ids: list, start="-10m"):
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
    def _topology_script(query):
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

    def topic_api(self, topic, limit=1000) -> dict:
        log = self.log
        log.info(f"Querying StackState Topic API: {topic}")

        def query():
            executed = self.host.run(f"sts-cli topic show {topic} -l {limit}")
            log.info(f"Queried {topic}: {executed.exit_status}")
            return executed.stdout

        return self._cached_json(query, topic)

    def _cached_json(self, api_call: Callable, alias: str):
        log = self.log
        caller = self._find_test_fn_name()
        cachefile = f"{caller}-{alias}.json"
        if self.cache_enabled:
            try:
                with open(cachefile, 'r') as f:
                    log.warning(f"Using cached result from {cachefile}")
                    return json.load(f)
            except IOError:
                pass

        executed = api_call()
        json_data = json.loads(executed)
        with open(cachefile, 'w') as f:
            log.info(f"Query result saved in file {cachefile}")
            f.write(executed)
        return json_data

    @staticmethod
    def _find_test_fn_name():
        frames = inspect.stack()
        for f in frames:
            if f.function.startswith("test_"):
                return f.function
        return "test_NA"
