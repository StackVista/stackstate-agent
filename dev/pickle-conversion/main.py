from tools import PickleConversion

instance_url = "https://localhost:8089"

v1_run_path = "run/v1"
v2_run_path = "run/v2"

events_search_name = "events-v1"
metrics_search_name = "metrics-v1"
relations_search_name = "relations-v1"
components_search_name = "components-v1"

service_identifier = "splunk"

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
