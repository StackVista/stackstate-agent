import json
import pickle
import sys
import re
import os
import time
import shutil


# Create a fake class for the checks class so that we do not have to attempt import on the old v1 project
class checks(object):
    pass


# Create a fake class for the CheckData object so that we do not have to attempt import on the old v1 project
# This is the primary class that the pickle file will be loaded into, meaning this is the core class we fetch data from
class CheckData(object):
    created_at = None
    created_by_pid = None
    data = None


# Map only the CheckData class into the check_status class seeing that the pickle file imports it
class check_status(object):
    # Map this class as it is used when mapping the pickle data file
    CheckData = CheckData


# Create a fake class for the AgentStatus class so that we do not have to attempt import on the old v1 project
class AgentStatus(object):
    pass


# Add the above classes as modules to create the import routes
sys.modules["checks"] = checks
sys.modules["checks.check_status"] = check_status
sys.modules["checks.check_status.CheckData"] = CheckData


class PickleConversion:
    # A general state for the state files
    general_state = {}

    # Stored state for v1 to v2 data
    state = {}

    def __init__(self, instance_url, service_identifier):
        # Create a directory for backups if something goes wrong
        if not os.path.exists("backups"):
            os.mkdir("backups")

        self.instance_url = instance_url
        self.service_identifier = service_identifier

    # Strip symbols from a url to create a format for v2
    @staticmethod
    def remove_symbols_from_string(value):
        return re.sub(r'\W', '', value)

    # Loop through the generated states and create the v2 state
    def export_v2(self, v2_directory):
        for key in self.state.keys():
            # If there is a check_state then let's generate a state file for it
            if "check_state" in self.state[key]:
                # The prev generate v2 file name format for this export
                filename = self.state[key].get("file_identifier_v2")

                # Attempt to backup the original state files when creating new state
                # If this fails then let's still continue
                try:
                    shutil.move("{}/{}".format(v2_directory, filename),
                                "backups/__backup-v2_{}-{}".format(str(time.time()), filename))
                except:
                    pass

                # Dump the check state that was generated
                content = json.dumps(self.state[key].get("check_state"), indent=4)

                # Write the dumped state into the v2 directory
                file = open("{}/{}".format(v2_directory, filename), "w")
                file.write(content)
                file.close()

                print("\nWrote the following content for the cache file: {}"
                      .format("{}/{}".format(v2_directory, filename)))
                print("{}\n\n".format(content))

        # Now we want to dump the generate state for all the checks
        filename = "{}_{}_check_state".format(
            self.service_identifier,
            self.remove_symbols_from_string(self.instance_url)
        )

        # Attempt to backup the original state files when creating new state
        # If this fails then let's still continue
        try:
            shutil.move("{}/{}".format(v2_directory, filename),
                        "backups/__backup-v2_{}__{}".format(str(time.time()), filename))
        except:
            pass


        # Dump the check state that was generated
        content = json.dumps(self.general_state, indent=4)

        # Write the dumped state into the v2 directory
        file = open("{}/{}".format(v2_directory, filename), "w")
        file.write(content)
        file.close()

        print("\nWrote the following content for the cache file: {}. This content is generated from the "
              .format("{}/{}".format(v2_directory, filename)))
        print("{}\n\n".format(content))

    # Loop through the current check state and create a general state
    def generate_v2_general_check_state(self):
        for key in self.state.keys():
            if key in self.state:
                self.general_state["sid_{}".format(key)] = self.state[key].get("data_identifier")
            else:
                print("'{}' does not exist for v2, skipping adding into the general check file for this v1 file"
                      .format(key))

    # Create a state for a single check from v1 information
    def generate_v2_single_check_state(self, saved_search_name):
        if saved_search_name in self.state:
            timestamp = self.state[saved_search_name].get("timestamp")

            if timestamp is not None:
                self.state[saved_search_name]["check_state"] = {
                    saved_search_name: timestamp
                }
        else:
            print("'{}' does not exist for v2, skipping conversion for this v1 file".format(saved_search_name))


    # Load a pickle files
    # Then map the information and predicted information into a object to allow
    # conversion on a later stage
    def convert_pickle_file(self, v1_directory, v1_filename, saved_search_name, v2_file_prefix=None):
        file_target = "{}/{}.pickle".format(v1_directory, v1_filename)

        if os.path.isfile(file_target):
            # Write the dumped state into the v2 directory
            shutil.copyfile(file_target,
                            "backups/__backup-v1_{}__{}.pickle".format(str(time.time()), v1_filename))


            if self.instance_url is None:
                raise Exception("Instance URL needs to be supplied before continuing")

            # The v1 pickle file
            v1_pickle_path = "{}/{}.pickle".format(v1_directory, v1_filename)

            # Load the pickle file content, it will load into the CheckData class
            f = open(v1_pickle_path, 'rb')
            v1_pickle_data = pickle.load(f)

            # If anything fails then we want to kill the script before writing
            if v1_pickle_data is None:
                raise Exception("Unable to retrieve v1 pickle data for `{}`".format(v1_pickle_path))

            elif v1_pickle_data.created_at is None:
                raise Exception("Pickle data is invalid, created_at value not found")

            elif v1_pickle_data.created_by_pid is None:
                raise Exception("Pickle data is invalid, created_by_pid value not found")

            elif v1_pickle_data.data is None:
                raise Exception("Pickle data is invalid, data value not found")

            # Build up instance information from v1
            instance_key = "{}{}".format(self.instance_url, saved_search_name)
            data_identifier = v1_pickle_data.data.get(instance_key)

            if data_identifier is None:
                raise Exception("Could not find the saved search name `{}` inside the pickle file `{}`"
                                .format(saved_search_name, v1_pickle_path))

            # All the data required to generate a v2 file or the original v1 file
            self.state[saved_search_name] = {
                "file_identifier_v1": v1_filename,
                "saved_search_name": saved_search_name,
                "instance_url": self.instance_url,
                "instance_key": instance_key,
                "data_identifier": str(data_identifier),
            }

            # If there is a v2_file_prefix then we want a v2 file otherwise no file will be generated for this v1 file
            if v2_file_prefix is not None:
                self.state[saved_search_name]["file_identifier_v2"] = "{}_{}_{}".format(self.service_identifier,
                                                                                        self.remove_symbols_from_string(
                                                                                            self.instance_url),
                                                                                        v2_file_prefix)

            # Attempt to get a timestamp for a certain instance url
            # If there is a timestamp then usually that will be a separate file of state so we split off in the object
            try:
                timestamp_entry = v1_pickle_data.data.get(self.instance_url)
                timestamp = timestamp_entry.get(saved_search_name)
                self.state[saved_search_name]['timestamp'] = timestamp

            except:
                pass
        else:
            print("'{}' does not exist, skipping conversion for this v1 file".format(file_target))
