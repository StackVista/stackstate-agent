import json
import hashlib
import inspect
import logging
import os
import time
from pathlib import Path
from typing import Callable

from testinfra.host import Host

from .models import *


class CLIv1:
    def __init__(self, host: Host, log: logging.Logger = logging, cache_enabled: bool = False):
        self.log = log
        self.host = host
        self.cache_enabled = cache_enabled

    def telemetry(self, component_ids, alias: str = None, config_location=None):
        if len(component_ids) == 0:
            return []
        fullquery = self._telemetry_script(component_ids)
        if alias is None:
            alias = self._query_digest(fullquery)

        result = self.script(fullquery, alias, config_location)
        series = {}

    def topology(self, query: str, alias: str = None, config_location=None) -> TopologyResult:
        logging.info(f"Executing the following query: '{query}'")

        fullquery = self._topology_script(query)
        if alias is None:
            alias = self._query_digest(fullquery)

        result = self.script(fullquery, alias, config_location)

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

    def script(self, fullquery: str, alias, config_location=None) -> dict:
        log = self.log
        log.info(f"Querying StackState Script API with CLIv1: {fullquery}, cache: {self.cache_enabled}")

        def find_file_or_alt(file_path, alt_file_path, error_message):
            if Path(file_path).is_file():
                return file_path
            elif Path(alt_file_path).is_file():
                return alt_file_path
            else:
                raise FileNotFoundError(error_message)

        def query():
            # Write topology query into the expected file
            home = os.path.expanduser("~")
            with open(f"{home}/sts-query.stsl", 'w') as script_query:
                # do stuff with temp file
                script_query.write(fullquery)

            # Find the sts-query.sh in the root, if it does not exist then attempt to find it in the bees dir
            # The alt will mainly be for local dev
            # TODO: Find a alternative to get to the './../../sut/bees/k8s-stackstate/files' dir if possible
            # TODO: The path of this file will never change but still feels horrible selecting it with a path like this
            file = find_file_or_alt(f"{home}/sts-query.sh",
                                    f"./../../sut/bees/k8s-stackstate/files/sts-query.sh",
                                    "Unable to find the sts-query.sh script")
            logging.info(f"Found the sts-query.sh at the following location: '{file}'")

            # Execute the query
            if config_location is None:
                executed = self.host.run(f'{file}')
            else:
                executed = self.host.run(f'{file} {config_location}')

            log.info(f"STSL script exit status: {executed.exit_status}")
            return executed.stdout

        return self._cached_json(query, alias)["result"]["value"]

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

    def topic_api(self, topic, limit=1000, config_location=None) -> dict:
        log = self.log
        log.info(f"Querying StackState Topic API: {topic}")

        def query():
            if config_location is None:
                executed = self.host.run(f"sts topic describe --name {topic} --nr {limit} --output json")
            else:
                executed = self.host.run(f"sts --config {config_location} topic describe --name {topic} --nr {limit} --output json")
            log.info(f"Queried {topic}: {executed.exit_status}")
            return executed.stdout

        return self._cached_json(query, topic)

    def promql_script(self, script: str, data_point_name: str = None) -> dict:
        log = self.log
        log.info(f"Querying StackState Script API: {script}")

        def query():
            executed = self.host.run(f'echo {script} | sts-cli script execute')
            log.info(f"Executed {script}: {executed.exit_status}")
            return executed.stdout

        return self._cached_json(query, data_point_name)

    def _cached_json(self, api_call: Callable, alias: str):
        log = self.log
        test_file, test_fn_name = self._find_test_fn_name()
        test_file = Path(test_file).stem

        # Creates debug dir under test group dir, where json files will be saved
        parent_path = Path(f"debug/{test_file}/{test_fn_name}")
        parent_path.mkdir(parents=True, exist_ok=True)

        cachefile = f"{parent_path}/{alias}_{time.time_ns()}.json"
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
    def _find_test_fn_name() -> (str, str):
        frames = inspect.stack()
        for f in frames:
            if f.function.startswith("test_"):
                return f.filename, f.function
        return "test_NA"
