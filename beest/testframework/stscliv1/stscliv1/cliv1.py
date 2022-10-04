import json
import hashlib
import inspect
import logging
import os
from typing import Callable
import uuid
from pathlib import Path

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

    def topic_api(self, topic, limit=1000) -> dict:
        log = self.log
        log.info(f"Querying StackState Topic API: {topic}")

        def query():
            executed = self.host.run(f"sts-cli topic show {topic} -l {limit}")
            log.info(f"Queried {topic}: {executed.exit_status}")
            return executed.stdout

        return self._cached_json(query, topic)

    def topology_topic(self, topic: str, limit: int = 250) -> TopicTopologyResult:
        json_data = self.topic_api(topic, limit)

        schema = TopicAPIResponseSchema()
        topic_response: TopicAPIResponse = schema.load(json_data)
        topic_result = TopicTopologyResult()

        for msg in topic_response.messages:
            payload = msg.message.topology_element.payload

            if payload.topology_component:
                topic_result.components.append(payload.topology_component.wrap())

            elif payload.topology_relation:
                topic_result.relations.append(payload.topology_relation.wrap())

            elif payload.topology_delete:
                topic_result.deletes.append(payload.topology_delete.wrap())
            else:
                pass

        return topic_result

    def topology_topic_snapshot(self, topic: str, limit: int = 1000) -> dict[str, TopologySnapshotResult]:
        json_data = self.topic_api(topic, limit)

        schema = TopicAPIResponseSchema()
        topic_response: TopicAPIResponse = schema.load(json_data)

        current_id = None
        snapshot_topology_results: dict[str, TopologySnapshotResult] = {}
        for msg in topic_response.messages:
            payload = msg.message.topology_element.payload

            if payload.topology_start_snapshot is not None:
                # if we reach start snapshot, we've reached the end of the current snapshot
                snapshot_topology_results[current_id].start_snapshot(msg.offset)

                # empty the current_id, until we reach the next stop_snapshot
                current_id = None

            elif current_id and payload.topology_component:
                snapshot_topology_results[current_id].component(payload.topology_component.wrap())

            elif current_id and payload.topology_relation:
                snapshot_topology_results[current_id].relation(payload.topology_relation.wrap())

            elif payload.topology_stop_snapshot is not None:
                # if we reach stop snapshot, we've reached the start of the current snapshot
                current_id = str(uuid.uuid4())
                snapshot_topology_results[current_id] = TopologySnapshotResult()

                snapshot_topology_results[current_id].stop_snapshot(msg.offset)

            else:
                pass

        return snapshot_topology_results

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
            executed = self.host.run(file)
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
