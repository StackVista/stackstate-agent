# Convert Agent v1 Pickle files to a Agent v2 JSON format

- If you installed Agent v2 make sure it is not running when doing this process by running `service stackstate-agent stop`
- CD into the `agent-v1-to-v2-pickle-conversion` directory
- Before executing the python script edit the following variables in the `config.yaml` file
  - `splunk_instance_url`
    - The instance url should match the url that is passed into the splunk conf files for example: `instances[].url` in `splunk_event_conf.yml`
  - `v1_sts_splunk_cache_folder`
    - Edit this to point to the run directory for the agent, Usually `/opt/stackstate-agent/run`
  - `v2_sts_splunk_cache_folder`
    - Edit this to point to the run directory for the agent, Usually `/opt/stackstate-agent/run`
  - `events_search_name`
    - Change this value to match the `saved_search` name in the `splunk_event_conf.yml` file
  - `metrics_search_name`
    - Change this value to match the `saved_search` name in the `splunk_metric_conf.yml` file
  - `relations_search_name`
    - Change this value to match the `relation_saved_search` name in the `splunk_topology_conf.yml` file
  - `components_search_name`
    - Change this value to match the `component_saved_search` name in the `splunk_topology_conf.yml` file
- Run the following command `sudo ./run.sh` this will execute the python script
- Confirm that the json files has been generated with content, this can be verified in the `v2_sts_splunk_cache_folder` directory
  - Make sure the ownership of these files and permission is correct and the correct user. If stackstate does not have permissions to these files then things will fail
  - The owner of the files is usually be stackstate. This can be done with this command `chown stackstate-agent:stackstate-agent splunk_*`
- Start the Agent v2 back up by running `service stackstate-agent start`
- Verify the logs for no errors


Notes:
- What if something went wrong?
  - If anything went wrong backups file is created alongside this script. These backups can be found in the backups folder.
  - If you want to restore one of these backup files, pick the earliest one this can be verified with the number in the file name and rename the file by remove the `__backup-v1_<NUMBER>__` from the file name for eventCheckData, metricCheckData, topologyCheckData and copy them back into the `v1_sts_splunk_cache_folder` folder

