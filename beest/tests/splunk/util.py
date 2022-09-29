import time
import logging
import re

from copy import copy
from typing import Callable, Union, Optional
from stscliv1 import CLIv1
from ststest import TopologyMatcher


# Find a specific query result in a StackState topic. You can further drill things down to test either
# a value or dict to make sure the result that is returned is expected.
# The StackState Topic will be queried with the CLIv1 and drilled down.
def find_in_topic(cliv1: CLIv1,
                  topic: str,
                  query: str,
                  first_match: bool = False,
                  equals_value: Union[str, bool, int, list[any]] = None,
                  contains_dict: dict[str, any] = None,
                  on_failure_action: Callable[[], None] = None) -> Optional[Union[list[dict], dict]]:
    # Retrieve data from a certain StackState Topic
    json_data = cliv1.topic_api(topic)

    # Verify that the value and dict contains is not defined
    # They will contradict each other and will always be false if both is tested
    # Either you want a direct value or test that the results contains a dict of definitions
    if equals_value is not None and contains_dict is not None:
        raise AttributeError("You should not specify both equals_value and contains_dict, Only use one.")

    # All the matched results will be stored here.
    matched_results: Optional[Union[list[dict], dict]] = []

    # If you only want the first match instead of returning everything
    # We can change the return from a list to a single
    if first_match is True:
        matched_results = None

    # The StackState Topic always returns a messages object. We can immediately drill into that
    for message in json_data["messages"]:

        # Split the query into parts to drill deeper into the json objects
        # For example a.b.c will become [a, b, c] we can then use this list to drill into the dict
        query_parts = query.split('.')

        # The json object will be drilled one level deeper, and the next part of query_parts will be taken
        # and the process repeated until the query_part does not exist or there is no query_parts left
        # This loop is called for every object in the query_parts variable
        # The cycle is:
        # 1. Pop the first value from query_parts
        # 2. Drill deeper into the dict with the popped value
        # 3 A. If the value exists and the query_parts array is not empty rerun the loop with the new drilled down
        #      dict and the new query_parts list without the popped value
        # 3 B. If the value does not exist then return a None as the dict we are looking at does not contain the
        #      result you are looking for.
        # 3 C. If the query_parts arr is empty, then return the dict as this dict matched all the requirements and is a
        #      match.
        def loop(message_target: dict[any], remaining_query: list[str]) -> Optional[dict[any]]:
            # If the remaining_query still contains values in the arr then we have not satisfied the condition
            if len(remaining_query) > 0:
                # Remove the first item from the remaining_query array and test the dict with this key
                key: str = remaining_query.pop(0)
                loop_through_all_values: bool = False
                after_key_index: Optional[int] = None

                try:
                    # if the key ends with [*] then it means that it can be any of the values in the array
                    # so no need to do regex for the number we just need to rerun the loop on every entry and see
                    # if something matches. This is an expensive operation so do not run this if the array is
                    # always 1 entry
                    if key.endswith("[*]"):
                        loop_through_all_values = True

                        # Cleanup the key value to allow us and still use it in the dict
                        key = key.replace(f"[*]", "")

                    # Do a check if the key contains the possible pattern for a index, before running the more
                    # expensive regex operation to find the int value
                    elif key.find("[") > -1 and key.find("]") > -1:
                        # Find if the key has a array int attached
                        array_index_after_key = re.search(r".*\[([0-9]+)]$", key)
                        if array_index_after_key is not None:
                            index_position = array_index_after_key.group(1)

                            # Cleanup the key value to allow us and still use it in the dict
                            key = key.replace(f"[{index_position}]", "")
                            after_key_index = int(index_position)
                except Exception as e:
                    raise ValueError(f"Unable to determine if there is a array index requirement, "
                                     f"Error received: {e}")

                # if the popped query value is in the dict we can rerun the loop but with the drilled down result.
                # We will continue the loop with the new lower level dict object, and the new remaining_query list
                # that does not contain the popped value.
                if key in message_target:
                    nested_dict = message_target[key]

                    try:
                        # If we need to loop through all the items to possibly find a match
                        if loop_through_all_values is True:
                            # Make sure we are testing a array
                            if type(nested_dict) is list:
                                for nested_dict_item in nested_dict:
                                    result = loop(nested_dict_item, copy(remaining_query))

                                    # If a valid result is found then return it and break the loop
                                    if result is not None:
                                        return result
                            else:
                                return None

                        # If there is a index to be fetched from the nested result
                        elif after_key_index is not None:
                            # Let's make sure the nested result is a dict and contains a element on that index
                            # after_key_index will be a 0 for 1 to we need to match it with the len value
                            if type(nested_dict) is list and len(nested_dict) >= after_key_index + 1:
                                nested_dict = nested_dict[after_key_index]
                            # If the array does not match then it means what we are looking at does not match the query
                            # so we can move on to the next result
                            else:
                                return None

                    except Exception as e:
                        raise KeyError(f"An array index was specified but failed when attempting to be "
                                       f"used in the nested value: {e}")

                    # Rerun the loop with the new information
                    return loop(nested_dict, remaining_query)

                # If we are unable to find the key in the dict then the dict we are testing against is invalid
                # and we can stop testing this dict
                else:
                    return None

            # If we manage to get to 0 queries left then we found what you are looking for and can start testing
            # the value to make sure it matches
            else:
                # If you want to check if the dict values exists in the result
                if contains_dict is not None:
                    # We need to test every key in the dict you defined to make sure all of them matches
                    for key in contains_dict:
                        # If the key exists in the dict and the dict value is equal to what you defined then we can
                        # continue the for loop
                        if key in message_target and message_target[key] == contains_dict[key]:
                            continue

                        # Otherwise it is invalid and we will return this as a invalid result
                        else:
                            return None

                # If you want to check if the query results is equal to a value
                elif equals_value is not None and message_target != equals_value:
                    return None

                # If none of the above caused a failure then the result is valid and we will return what we found.
                return message_target

        # Start the loop with the original message, and the query broken up into a array
        query_results = loop(message, query_parts)

        # If you wish to return only one result then we can instantly break from the loop and return the matched result
        if first_match is True and query_results is not None:
            matched_results = query_results

        # If we found a match then lets push it into the list and keep on searching for more results, We will not break
        # from the loop
        elif query_results is not None:
            matched_results.append(query_results)

    if matched_results is None or \
       matched_results == "" or \
       matched_results is dict and len(matched_results) <= 0:

        if on_failure_action is not None:
            on_failure_action()
            return matched_results
        else:
            raise Exception(f"Value not found in topic, Topic: {topic}, Query: {query}")

    return matched_results


def wait_until_topic_match(cliv1: CLIv1,
                           timeout: int,
                           topic: str,
                           query: str,
                           first_match: bool = False,
                           contains_dict: dict[str, any] = None,
                           equals_value: Union[str, bool, int, list[any]] = None,
                           period: int = 0.25,
                           on_failure_action: Callable[[], None] = None) -> Optional[Union[list[dict], dict]]:
    def loop() -> Optional[Union[list[dict], dict]]:
        def raise_not_found():
            raise Exception(f"Value not found in topic, Topic: {topic}, Query: {query}")

        return find_in_topic(cliv1,
                             topic=topic,
                             query=query,
                             first_match=first_match,
                             equals_value=equals_value,
                             contains_dict=contains_dict,
                             on_failure_action=raise_not_found)

    return wait_until(loop, timeout, period, on_failure_action)


def wait_until_topology_match(cliv1: CLIv1,
                              topology_matcher: Callable[[], TopologyMatcher],
                              topology_query: Callable[[], str],
                              timeout: int,
                              period: int = 0.25,
                              on_failure_action: Callable[[], None] = None) -> None:
    def loop():
        # Call the Lambda function to get the Topology builder and the compiled query
        expected = topology_matcher()
        actual = topology_query()

        # Run the CLI and the TopologyBuilder and attempt to find a match
        actual_topology = cliv1.topology(actual)
        expected_matches = expected.find(actual_topology)
        expected_matches.assert_exact_match()

    wait_until(loop, timeout, period, on_failure_action)


def wait_until(someaction: Callable[[any, any], any],
               timeout: int,
               period: int = 0.25,
               on_failure_action: Callable[[], None] = None,
               *args: any,
               **kwargs: any) -> any:
    mustend = time.time() + timeout
    while True:
        try:
            return someaction(*args, **kwargs)
        except:
            if time.time() >= mustend:
                logging.error("Waiting timed out after %d" % timeout)
                if on_failure_action is not None:
                    logging.error("Running on_failure_action action")
                    on_failure_action()
                raise
            time.sleep(period)
