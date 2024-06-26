"""
Cluster Agent tasks
"""

import glob
import os
import shutil

from invoke import task
from invoke.exceptions import Exit

from .build_tags import get_build_tags, get_default_build_tags
from .cluster_agent_helpers import build_common, clean_common, refresh_assets_common, version_common
from .go import deps
from .utils import (
    load_release_versions,
    do_go_rename,  # sts
    do_sed_rename,  # sts
)

# constants
BIN_PATH = os.path.join(".", "bin", "stackstate-cluster-agent")
AGENT_TAG = "stackstate/cluster_agent:master"
POLICIES_REPO = "https://github.com/DataDog/security-agent-policies.git"


# sts begin
@task
def apply_branding(ctx):
    """
    Apply stackstate branding
    """
    sts_lower_replace = 's/datadog/stackstate/g'
    datadog_metrics_replace = 's/"datadog./"stackstate./g'

    # Config
    do_go_rename(ctx, '"\\"dd_url\\" -> \\"sts_url\\""', "./pkg/config")
    do_go_rename(ctx, '"\\"https://app.datadoghq.com\\" -> \\"http://localhost:7077\\""', "./pkg/config")
    do_go_rename(ctx, '"\\"DD_PROXY_HTTP\\" -> \\"STS_PROXY_HTTP\\""', "./pkg/config")
    do_go_rename(ctx, '"\\"DD_PROXY_HTTPS\\" -> \\"STS_PROXY_HTTPS\\""', "./pkg/config")
    do_go_rename(ctx, '"\\"DD_PROXY_NO_PROXY\\" -> \\"STS_PROXY_NO_PROXY\\""', "./pkg/config")
    do_go_rename(ctx, '"\\"DOCKER_DD_AGENT\\" -> \\"DOCKER_STS_AGENT\\""', "./pkg/config")
    do_go_rename(ctx, '"\\"DD\\" -> \\"STS\\""', "./pkg/config")
    do_go_rename(ctx, '"\\"datadog\\" -> \\"stackstate\\""', "./pkg/config")
    do_go_rename(ctx, '"\\"/etc/datadog-agent/conf.d\\" -> \\"/etc/stackstate-agent/conf.d\\""', "./pkg/config")
    do_go_rename(ctx, '"\\"/etc/datadog-agent/checks.d\\" -> \\"/etc/stackstate-agent/checks.d\\""', "./pkg/config")
    do_go_rename(ctx, '"\\"/opt/datadog-agent/run\\" -> \\"/opt/stackstate-agent/run\\""', "./pkg/config")

    # [sts] turn of the metadata collection, the receiver does not recognize these payloads
    do_sed_rename(ctx, 's/"enable_metadata_collection"\\, true/"enable_metadata_collection"\\, false/g',
                  "./pkg/config/config.go")
    do_sed_rename(ctx, 's/"enable_gohai"\\, true/"enable_gohai"\\, false/g', "./pkg/config/config.go")
    do_sed_rename(ctx, 's/"inventories_enabled"\\, true/"inventories_enabled"\\, false/g', "./pkg/config/config.go")

    # Trace Agent Metrics
    # do_sed_rename(ctx, datadog_metrics_replace, "./pkg/process/statsd/statsd.go")
    do_sed_rename(ctx, datadog_metrics_replace, "./vendor/github.com/DataDog/datadog-go/statsd/statsd.go")
    do_sed_rename(ctx, datadog_metrics_replace, "./vendor/github.com/DataDog/datadog-go/statsd/telemetry.go")

    # Cluster Agent
    cluster_agent_replace = '/www/! s/datadog/stackstate/g'
    do_sed_rename(ctx, cluster_agent_replace, "./cmd/cluster-agent/main.go")
    do_sed_rename(ctx, cluster_agent_replace, "./cmd/cluster-agent/app/*")
    do_sed_rename(ctx, cluster_agent_replace, "./cmd/cluster-agent/commands/*")
    do_sed_rename(ctx, cluster_agent_replace, "./cmd/agent/common/commands/*.go")
    do_sed_rename(ctx, 's/Datadog Cluster/StackState Cluster/g', "./cmd/cluster-agent/app/*")
    do_sed_rename(ctx, 's/Datadog Agent/StackState Agent/g', "./cmd/cluster-agent/app/*")
    do_sed_rename(ctx, 's/to Datadog/to StackState/g', "./cmd/cluster-agent/app/*")
    do_sed_rename(ctx, 's/"datadog-cluster"/"stackstate-cluster"/g', "./cmd/cluster-agent/app/*")

    # Cluster Agent - Kubernetes API client
    do_go_rename(ctx, '"\\"datadogtoken\\" -> \\"stackstatetoken\\""', "./pkg/util/kubernetes/apiserver")

    # Defaults
    do_go_rename(ctx, '"\\"/etc/datadog-agent\\" -> \\"/etc/stackstate-agent\\""', "./cmd/agent/common")
    do_go_rename(ctx, '"\\"/var/log/datadog/cluster-agent.log\\" -> \\"/var/log/stackstate-agent/cluster-agent.log\\""',
                 "./cmd/agent/common")
    do_go_rename(ctx, '"\\"datadog.yaml\\" -> \\"stackstate.yaml\\""', "./cmd/agent")
    do_go_rename(ctx, '"\\"datadog.conf\\" -> \\"stackstate.conf\\""', "./cmd/agent")
    do_go_rename(ctx,
                 '"\\"path to directory containing datadog.yaml\\" -> \\"path to directory containing stackstate.yaml\\""',
                 "./cmd")
    do_go_rename(ctx,
                 '"\\"unable to load Datadog config file: %s\\" -> \\"unable to load StackState config file: %s\\""',
                 "./cmd/agent/common")
    do_go_rename(ctx,
                 '"\\"unable to load Datadog config file: %w\\" -> \\"unable to load StackState config file: %w\\""',
                 "./cmd/agent/common")

    # Hardcoded checks and metrics
    do_sed_rename(ctx, sts_lower_replace, "./pkg/aggregator/aggregator.go")


# sts end

@task
def build(
    ctx,
    rebuild=False,
    build_include=None,
    build_exclude=None,
    race=False,
    development=True,
    skip_assets=False,
    policies_version=None,
    release_version="nightly-a7",
):
    """
    Build Cluster Agent

     Example invokation:
        inv cluster-agent.build
    """
    apply_branding(ctx)
    build_common(
        ctx,
        BIN_PATH,
        get_default_build_tags(build="cluster-agent"),
        "",
        rebuild,
        build_include,
        build_exclude,
        race,
        development,
        skip_assets,
    )

    # sts - ignore security policies (we don't use the security agent)

    # if policies_version is None:
    #     print("Loading release versions for {}".format(release_version))
    #     env = load_release_versions(ctx, release_version)
    #     if "SECURITY_AGENT_POLICIES_VERSION" in env:
    #         policies_version = env["SECURITY_AGENT_POLICIES_VERSION"]
    #         print("Security Agent polices for {}: {}".format(release_version, policies_version))
    #
    # build_context = "Dockerfiles/cluster-agent"
    # policies_path = "{}/security-agent-policies".format(build_context)
    # ctx.run("rm -rf {}".format(policies_path))
    # ctx.run("git clone {} {}".format(POLICIES_REPO, policies_path))
    # if policies_version != "master":
    #     ctx.run("cd {} && git checkout {}".format(policies_path, policies_version))


@task
def refresh_assets(ctx, development=True):
    """
    Clean up and refresh cluster agent's assets and config files
    """
    refresh_assets_common(ctx, BIN_PATH, [os.path.join("./Dockerfiles/cluster-agent", "dist")], development)


@task
def clean(ctx):
    """
    Remove temporary objects and binary artifacts
    """
    clean_common(ctx, "stackstate-cluster-agent")  # sts


@task
def integration_tests(ctx, install_deps=False, race=False, remote_docker=False, go_mod="mod"):
    """
    Run integration tests for cluster-agent
    """
    if install_deps:
        deps(ctx)

    # We need docker for the kubeapiserver integration tests
    tags = get_default_build_tags(build="cluster-agent") + ["docker"]

    test_args = {
        "go_mod": go_mod,
        "go_build_tags": " ".join(get_build_tags(tags, [])),
        "race_opt": "-race" if race else "",
        "exec_opts": "",
    }

    # since Go 1.13, the -exec flag of go test could add some parameters such as -test.timeout
    # to the call, we don't want them because while calling invoke below, invoke
    # thinks that the parameters are for it to interpret.
    # we're calling an intermediate script which only pass the binary name to the invoke task.
    if remote_docker:
        test_args["exec_opts"] = "-exec \"{}/test/integration/dockerize_tests.sh\"".format(os.getcwd())

    go_cmd = 'go test -mod={go_mod} {race_opt} -tags "{go_build_tags}" {exec_opts}'.format(**test_args)

    prefixes = [
        "./test/integration/util/kube_apiserver",
        "./test/integration/util/leaderelection",
    ]

    for prefix in prefixes:
        ctx.run("{} {}".format(go_cmd, prefix))


@task
def image_build(ctx, arch='amd64', tag=AGENT_TAG, push=False):
    """
    Build the docker image
    """

    dca_binary = glob.glob(os.path.join(BIN_PATH, "stackstate-cluster-agent"))  # sts
    # get the last debian package built
    if not dca_binary:
        print("No bin found in {}".format(BIN_PATH))
        print("See cluster-agent.build")
        raise Exit(code=1)
    latest_file = max(dca_binary, key=os.path.getctime)
    ctx.run("chmod +x {}".format(latest_file))

    build_context = "Dockerfiles/cluster-agent"
    exec_path = "{}/stackstate-cluster-agent.{}".format(build_context, arch)  # sts
    dockerfile_path = "{}/{}/Dockerfile".format(build_context, arch)

    shutil.copy2(latest_file, exec_path)
    ctx.run("docker build -t {} {} -f {}".format(tag, build_context, dockerfile_path))
    ctx.run("rm {}".format(exec_path))

    if push:
        ctx.run("docker push {}".format(tag))


@task
def version(ctx, url_safe=False, git_sha_length=7):
    """
    Get the agent version.
    url_safe: get the version that is able to be addressed as a url
    git_sha_length: different versions of git have a different short sha length,
                    use this to explicitly set the version
                    (the windows builder and the default ubuntu version have such an incompatibility)
    """
    version_common(ctx, url_safe, git_sha_length)


@task
def update_generated_code(ctx):
    """
    Re-generate 'pkg/clusteragent/custommetrics/api/generated/openapi/zz_generated.openapi.go'.
    """
    ctx.run("go install -mod=readonly k8s.io/kube-openapi/cmd/openapi-gen")
    ctx.run(
        "$GOPATH/bin/openapi-gen \
--logtostderr \
-i k8s.io/metrics/pkg/apis/custom_metrics,k8s.io/metrics/pkg/apis/custom_metrics/v1beta1,k8s.io/metrics/pkg/apis/custom_metrics/v1beta2,k8s.io/metrics/pkg/apis/external_metrics,k8s.io/metrics/pkg/apis/external_metrics/v1beta1,k8s.io/metrics/pkg/apis/metrics,k8s.io/metrics/pkg/apis/metrics/v1beta1,k8s.io/apimachinery/pkg/apis/meta/v1,k8s.io/apimachinery/pkg/api/resource,k8s.io/apimachinery/pkg/version,k8s.io/api/core/v1 \
-h ./tools/boilerplate.go.txt \
-p ./pkg/clusteragent/custommetrics/api/generated/openapi \
-O zz_generated.openapi \
-o ./ \
-r /dev/null"
    )
