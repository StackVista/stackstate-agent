testinfra_hosts = ["local"]


def _get_pod_restarts(kubecontext, host):
    jsonpath = "'{range .items[*]}|{.metadata.name}={range .status.containerStatuses[*]},{.restartCount}'"
    cmd = host.run("kubectl --context={0} get pod -o jsonpath={1}".format(kubecontext, jsonpath))

    assert cmd.rc == 0
    restarts = cmd.stdout.split("|")
    print(restarts)
    return restarts


def test_agents_do_not_restart(host, ansible_var):
    kubecontext = ansible_var("agent_kubecontext")
    for pod in _get_pod_restarts(kubecontext, host):
        if "=," in pod:
            pod_split = pod.split("=,")
            pod_name, restarts = pod_split[0], pod_split[1].split(",")
            for restart in restarts:
                if restart.isnumeric():
                    assert int(restart) < 0, "pod {} has a container with {} restarts".format(pod_name, restart)
