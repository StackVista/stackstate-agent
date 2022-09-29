import json
import yaml
import time

from tools import PickleConversion

with open(r'config.yaml') as file:
    config = yaml.load(file, Loader=yaml.FullLoader)

    print("Found the following config:")
    print("You have 30 seconds to verify if the below config is correct ... \n")
    print(json.dumps(config, indent=4))

    time.sleep(30)

    service_identifier = "splunk"

    # Get the required data from the config.yaml file
    instance_url = config.get('splunk_instance_url')
    v1_run_path = config.get('v1_sts_splunk_cache_folder')
    v2_run_path = config.get('v2_sts_splunk_cache_folder')
    events_search_name = config.get('events_search_name')
    metrics_search_name = config.get('metrics_search_name')
    relations_search_name = config.get('relations_search_name')
    components_search_name = config.get('components_search_name')

    # Verify that we have no broken data
    if instance_url is None or instance_url == "" or \
        v1_run_path is None or v1_run_path == ""or \
        v2_run_path is None or v2_run_path == ""or \
        events_search_name is None or events_search_name == ""or \
        metrics_search_name is None or metrics_search_name == ""or \
        relations_search_name is None or relations_search_name == ""or \
        components_search_name is None or components_search_name == "":
        raise Exception("The config.yaml contains invalid or missing data")

    pickle_converter = PickleConversion(
        instance_url=instance_url,
        service_identifier=service_identifier
    )

    # Convert Event Pickle File
    pickle_converter.convert_pickle_file(v1_directory=v1_run_path,
                                         v1_filename="{}_eventCheckData".format(service_identifier),
                                         v2_file_prefix="splunk_event",
                                         saved_search_name=events_search_name)

    # Convert Metric Pickle File
    pickle_converter.convert_pickle_file(v1_directory=v1_run_path,
                                         v1_filename="{}_metricCheckData".format(service_identifier),
                                         v2_file_prefix="transactional_check_state",
                                         saved_search_name=metrics_search_name)

    # Convert Relation Pickle File
    pickle_converter.convert_pickle_file(v1_directory=v1_run_path,
                                         v1_filename="{}_topologyCheckData".format(service_identifier),
                                         saved_search_name=relations_search_name)

    # Convert Components Pickle File
    pickle_converter.convert_pickle_file(v1_directory=v1_run_path,
                                         v1_filename="{}_topologyCheckData".format(service_identifier),
                                         saved_search_name=components_search_name)

    # Generate general state for all the checks
    pickle_converter.generate_v2_general_check_state()

    # Generate a state format for events
    pickle_converter.generate_v2_single_check_state(events_search_name)

    # Generate a state format for metrics
    pickle_converter.generate_v2_single_check_state(metrics_search_name)

    # Generate a state format for relations
    pickle_converter.generate_v2_single_check_state(relations_search_name)

    # Generate a state format for components
    pickle_converter.generate_v2_single_check_state(components_search_name)

    # Export the generated information into the v2 directory
    pickle_converter.export_v2(v2_directory=v2_run_path)
