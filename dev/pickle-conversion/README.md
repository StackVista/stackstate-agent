# Convert Agent v1 Pickle files to a Agent v2 JSON format

- CD into the `pickle-conversion` directory


- Run the command `tox` to enable the python 2 environment or make sure the env your are running this script in is python version `2.7.18`


- Before executing the test python script edit the following variables in the `main.py` file
  - `instance_url`
    - The instance url should match the url that is passed into the splunk conf files for example: `instances[].url` in `splunk_event_conf.yml`
  - `v1_run_path`
    - Edit this to point to the run directory for the agent, Usually `/opt/stackstate-agent/run`
  - `v2_run_path`
    - Edit this to point to the run directory for the agent, Usually `/opt/stackstate-agent/run`
  - `events_search_name`
    - Change this value to match the `saved_search` name in the `splunk_event_conf.yml` file
  - `metrics_search_name`
    - Change this value to match the `saved_search` name in the `splunk_metric_conf.yml` file
  - `relations_search_name`
    - Change this value to match the `relation_saved_search` name in the `splunk_topology_conf.yml` file
  - `components_search_name`
    - Change this value to match the `component_saved_search` name in the `splunk_topology_conf.yml` file


- Execute the pickle conversion script using the .tox python2 executable
  - `./.tox/py27/bin/python main.py`

- Confirm that the json files has been generated with content, this can be verified in the `v2_run_path` directory


Note: If anything went wrong a __bac__ file is created alongside this script, these are the original splunk run files is state for v1 needs to be restored
