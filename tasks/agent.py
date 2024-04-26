"""
Agent namespaced tasks
"""


import ast
import glob
import os
import platform
import re
import shutil
import sys
import tempfile

from invoke import task
from invoke.exceptions import Exit, ParseError

from .build_tags import filter_incompatible_tags, get_build_tags, get_default_build_tags
from .flavor import AgentFlavor
from .go import deps
from .process_agent import build as process_agent_build
from .rtloader import clean as rtloader_clean
from .rtloader import install as rtloader_install
from .rtloader import make as rtloader_make
from .ssm import get_pfx_pass, get_signing_cert
from .trace_agent import build as trace_agent_build
from .utils import (
    REPO_PATH,
    bin_name,
    cache_version,
    generate_config,
    get_build_flags,
    get_version,
    has_both_python,
    load_release_versions,
    timed,
    do_go_rename,  # sts
    do_sed_rename,  # sts
    do_find_sed_rename,  # sts
    do_find_sed_rename_pattern_multi_ignore,  # sts
)
from .windows_resources import build_messagetable, build_rc, versioninfo_vars

# constants
BIN_PATH = os.path.join(".", "bin", "agent")
AGENT_TAG = "datadog/agent:master"

AGENT_CORECHECKS = [
    "container",
    "containerd",
    "container_image",
    "container_lifecycle",
    "cpu",
    "cri",
    "snmp",
    "docker",
    "disk",
    "file_handle",
    "go_expvar",
    "io",
    "jmx",
    "kubernetes_apiserver",
    "load",
    "memory",
    "ntp",
    "oom_kill",
    "oracle-dbm",
    "sbom",
    "systemd",
    "tcp_queue_length",
    "uptime",
    "winproc",
    "jetson",
    "telemetry",
    "orchestrator_pod",
]

WINDOWS_CORECHECKS = [
    "agentcrashdetect",
    "windows_registry",
    "winkmem",
    "wincrashdetect",
]

IOT_AGENT_CORECHECKS = [
    "cpu",
    "disk",
    "io",
    "load",
    "memory",
    "network",
    "ntp",
    "uptime",
    "systemd",
    "jetson",
]

CACHED_WHEEL_FILENAME_PATTERN = "datadog_{integration}-*.whl"
CACHED_WHEEL_DIRECTORY_PATTERN = "integration-wheels/{branch}/{hash}/{python_version}/"
CACHED_WHEEL_FULL_PATH_PATTERN = CACHED_WHEEL_DIRECTORY_PATTERN + CACHED_WHEEL_FILENAME_PATTERN
LAST_DIRECTORY_COMMIT_PATTERN = "git -C {integrations_dir} rev-list -1 HEAD {integration}"


@task
def apply_branding(ctx):
    """
    Apply stackstate branding
    """
    sts_camel_replace = 's/Data[dD]og/StackState/g'
    sts_lower_replace = 's/datadog/stackstate/g'
    datadog_metrics_replace = 's/"datadog./"stackstate./g'
    datadog_checks_replace = 's/"datadog_checks./"stackstate_checks./g'

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
    do_go_rename(ctx, '"\\"https://process.datadoghq.com\\" -> \\"http://localhost:7077\\""', "./pkg/config/setup")
    do_go_rename(ctx, '"\\"https://orchestrator.datadoghq.com\\" -> \\"http://localhost:7077\\""', "./pkg/config/setup")

    # [sts] turn of the metadata collection, the receiver does not recognize these payloads
    do_sed_rename(ctx, 's/"enable_metadata_collection"\\, true/"enable_metadata_collection"\\, false/g', "./pkg/config/setup/config.go")
    do_sed_rename(ctx, 's/"enable_gohai"\\, true/"enable_gohai"\\, false/g', "./pkg/config/setup/config.go")
    do_sed_rename(ctx, 's/"inventories_enabled"\\, true/"inventories_enabled"\\, false/g', "./pkg/config/setup/config.go")

    # [sts] switch DisableAPIKeyChecking to true to skip api key validation
    do_sed_rename(ctx, 's/DisableAPIKeyChecking:    false/DisableAPIKeyChecking:    true/g', "./comp/forwarder/defaultforwarder/default_forwarder.go")

    # cmd/agent/common/common_windows.go
    do_sed_rename(ctx, 's/"programdata\\\\\\\\\\\\\\\\datadog"/"programdata\\\\\\\\\\\\\\\\stackstate"/g',
                  "./cmd/agent/common/common_windows.go")
    do_sed_rename(ctx, 's/"ProgramData\\\\\\\\\\\\\\\\datadog"/"ProgramData\\\\\\\\\\\\\\\\StackState"/g',
                  "./cmd/agent/common/common_windows.go")
    do_sed_rename(ctx, 's/"Datadog"/"StackState"/g',
                  "./cmd/agent/common/common_windows.go")
    do_sed_rename(ctx, 's/"ProgramData\\\\\\\\\\\\\\\\StackVista"/"ProgramData\\\\\\\\\\\\\\\\StackState"/g',
                  "./cmd/agent/common/common_windows.go")
    do_sed_rename(ctx, 's/"SOFTWARE\\\\StackVista\\\\"/"SOFTWARE\\\\\StackState\\\\"/g',
                  "./cmd/agent/common/common_windows.go")
    do_sed_rename(ctx, 's/"datadog.conf"/"stackstate.conf"/g',
                  "./cmd/agent/common/common_windows.go")
#     do_sed_rename(ctx, 's/"\\\\StackState\\\\StackVista"/"\\\\StackState\\\\StackState Agent"/g',
#                   "./tools/windows/install-help/cal/stringtable.rc")
#     do_sed_rename(ctx, 's/"\\StackState\\StackVista"/"\\StackState\\StackState Agent"/g',
#                   "./tools/windows/install-help/cal/stringtable.rc")
#     do_sed_rename(ctx, 's/"\\\\\\\\\\\\\\\\StackState\\\\\\\\\\\\\\\\StackVista"/"\\\\\\\\\\\\\\\\StackState\\\\\\\\\\\\\\\\StackState Agent"/g',
#                   "./tools/windows/install-help/cal/stringtable.rc")

    # DD 7.33 Upstream Merge changes
    do_sed_rename(ctx, 's/"datadog.yaml"/"stackstate.yaml"/g',
                       "./cmd/agent/common/import.go")
    do_sed_rename(ctx, 's/"datadog.conf"/"stackstate.conf"/g',
                       "./cmd/agent/common/import.go")

    # systray.go
    do_sed_rename(ctx, 's/"programdata\\\\\\\\\\\\\\\\datadog"/"programdata\\\\\\\\\\\\\\\\stackstate"/g',
                  "./comp/systray/systray/systrayimpl/systray.go")
    do_sed_rename(ctx, 's/"ProgramData\\\\\\\\\\\\\\\\datadog"/"ProgramData\\\\\\\\\\\\\\\\StackState"/g',
                  "./comp/systray/systray/systrayimpl/systray.go")
    do_sed_rename(ctx, 's/"Datadog"/"Stackstate"/g',
                  "./comp/systray/systray/systrayimpl/systray.go")
    do_sed_rename(ctx, 's/"ProgramData\\\\\\\\\\\\\\\\StackVista"/"ProgramData\\\\\\\\\\\\\\\\StackState"/g',
                  "./comp/systray/systray/systrayimpl/systray.go")
    # pkg/config/setup/config_windows.go
    do_sed_rename(ctx, 's/"programdata\\\\\\\\\\\\\\\\datadog"/"programdata\\\\\\\\\\\\\\\\stackstate"/g',
                  "./pkg/config/setup/config_windows.go")
    do_sed_rename(ctx, 's/"ProgramData\\\\\\\\\\\\\\\\datadog"/"ProgramData\\\\\\\\\\\\\\\\StackState"/g',
                  "./pkg/config/setup/config_windows.go")
    do_sed_rename(ctx, 's/"Datadog"/"Stackstate"/g',
                  "./pkg/config/setup/config_windows.go")
    do_sed_rename(ctx, 's/"ProgramData\\\\\\\\\\\\\\\\StackVista"/"ProgramData\\\\\\\\\\\\\\\\StackState"/g',
                  "./pkg/config/setup/config_windows.go")
    do_sed_rename(ctx, 's/"SOFTWARE\\\\StackVista\\\\"/"SOFTWARE\\\\\StackState\\\\"/g',
                  "./pkg/config/setup/config_windows.go")
    do_sed_rename(ctx, 's/"datadog.conf"/"stackstate.conf"/g',
                  "./pkg/config/setup/config_windows.go")
#     do_sed_rename(ctx, 's/"Datadog Agent Service"/"StackState Agent Service"/g',
#                   "./cmd/agent/app/install_service_windows.go")
    do_find_sed_rename_pattern_multi_ignore(ctx, "*windows*.go", 's/"Datadog Agent Service"/"StackState Agent Service"/g', ".", [".git", "vendor"])
    do_find_sed_rename_pattern_multi_ignore(ctx, "*windows*.go", 's/"Datadog Agent service"/"StackState Agent service"/g', ".", [".git", "vendor"])
#     do_sed_rename(ctx, 's/"Datadog Agent Service"/"StackState Agent Service"/g',
#                   "./cmd/process-agent/main_windows.go")
#     do_sed_rename(ctx, 's/"Datadog Agent Service"/"StackState Agent Service"/g',
#                   "./cmd/trace-agent/main_windows.go")
    do_sed_rename(ctx, 's/"DatadogAgent"/"StackStateAgent"/g',
                  "./pkg/config/setup/config_windows.go")
    do_sed_rename(ctx, 's/"datadog-process-agent"/"stackstate-process-agent"/g',
                  "./cmd/process-agent/main_windows.go")
#     do_sed_rename(ctx, 's/"datadog-trace-agent"/"stackstate-trace-agent"/g',
#                   "./cmd/trace-agent/main_windows.go")
    do_sed_rename(ctx, 's/"c:\\\\\\\\\\\\\\\\programdata\\\\\\\\\\\\\\\\datadog\\\\\\\\\\\\\\\\datadog.yaml"/"c:\\\\\\\\\\\\\\\\programdata\\\\\\\\\\\\\\\\stackstate\\\\\\\\\\\\\\\\stackstate.yaml"/g',
                  "./cmd/process-agent/main_windows.go")
    do_sed_rename(ctx, 's/"c:\\\\\\\\\\\\\\\\programdata\\\\\\\\\\\\\\\\datadog\\\\\\\\\\\\\\\\system-probe.yaml"/"c:\\\\\\\\\\\\\\\\programdata\\\\\\\\\\\\\\\\stackstate\\\\\\\\\\\\\\\\system-probe.yaml"/g',
                  "./cmd/process-agent/main_windows.go")
    do_sed_rename(ctx, 's/"c:\\\\\\\\\\\\\\\\programdata\\\\\\\\\\\\\\\\datadog\\\\\\\\\\\\\\\\conf.d"/"c:\\\\\\\\\\\\\\\\programdata\\\\\\\\\\\\\\\\stackstate\\\\\\\\\\\\\\\\conf.d"/g',
                  "./cmd/process-agent/main_windows.go")
    do_sed_rename(ctx, 's/"c:\\\\\\\\\\\\\\\\programdata\\\\\\\\\\\\\\\\datadog\\\\\\\\\\\\\\\\logs\\\\\\\\\\\\\\\\process-agent.log"/"c:\\\\\\\\\\\\\\\\programdata\\\\\\\\\\\\\\\\stackstate\\\\\\\\\\\\\\\\logs\\\\\\\\\\\\\\\\process-agent.log"/g',
                  "./cmd/process-agent/main_windows.go")
    do_sed_rename(ctx, 's/"datadog.yaml"/"stackstate.yaml"/g',
                  "./cmd/process-agent/main_windows.go")
    # pkg/pidfile/pidfile_windows.go
    do_sed_rename(ctx, 's/"programdata\\\\\\\\\\\\\\\\datadog"/"programdata\\\\\\\\\\\\\\\\stackstate"/g',
                  "./pkg/pidfile/pidfile_windows.go")
    do_sed_rename(ctx, 's/"ProgramData\\\\\\\\\\\\\\\\datadog"/"ProgramData\\\\\\\\\\\\\\\\StackState"/g',
                  "./pkg/pidfile/pidfile_windows.go")
    do_sed_rename(ctx, 's/"ProgramData\\\\\\\\\\\\\\\\StackVista"/"ProgramData\\\\\\\\\\\\\\\\StackState"/g',
                  "./pkg/pidfile/pidfile_windows.go")
    do_sed_rename(ctx, 's/"Datadog"/"Stackstate"/g',
                  "./pkg/pidfile/pidfile_windows.go")
    do_sed_rename(ctx, 's/"datadog"/"stackstate"/g',
                  "./pkg/pidfile/pidfile_windows.go")
    do_sed_rename(ctx, 's/StackVista\\\\\\\\\\\\\\\\stackstate-agent.pid"/StackState\\\\\\\\\\\\\\\\stackstate-agent.pid"/g',
                  "./pkg/pidfile/pidfile_windows.go")
    # pkg/trace/flags/flags_windows.go
    do_find_sed_rename_pattern_multi_ignore(ctx, "*.go", 's/"programdata\\\\\\\\\\\\\\\\datadog"/"programdata\\\\\\\\\\\\\\\\stackstate"/g', ".", [".git", "vendor"])
    do_find_sed_rename_pattern_multi_ignore(ctx, "*.go", 's/"ProgramData\\\\\\\\\\\\\\\\datadog"/"ProgramData\\\\\\\\\\\\\\\\stackstate"/g', ".", [".git", "vendor"])
    do_find_sed_rename_pattern_multi_ignore(ctx, "*.go", 's/"programdata\\\\\\\\\\\\\\\\DataDog"/"programdata\\\\\\\\\\\\\\\\StackState"/g', ".", [".git", "vendor"])
    do_find_sed_rename_pattern_multi_ignore(ctx, "*.go", 's/"ProgramData\\\\\\\\\\\\\\\\DataDog"/"ProgramData\\\\\\\\\\\\\\\\StackState"/g', ".", [".git", "vendor"])
#     do_sed_rename(ctx, 's/"programdata\\\\\\\\\\\\\\\\datadog"/"programdata\\\\\\\\\\\\\\\\stackstate"/g',
#                   "./pkg/trace/flags/flags_windows.go")
#     do_sed_rename(ctx, 's/"ProgramData\\\\\\\\\\\\\\\\datadog"/"ProgramData\\\\\\\\\\\\\\\\StackState"/g',
#                   "./pkg/trace/flags/flags_windows.go")
#     do_sed_rename(ctx, 's/"Datadog"/"Stackstate"/g',
#                   "./pkg/trace/flags/flags_windows.go")
#     do_sed_rename(ctx, 's/"ProgramData\\\\\\\\\\\\\\\\StackVista"/"ProgramData\\\\\\\\\\\\\\\\StackState"/g',
#                   "./pkg/trace/flags/flags_windows.go")
#     do_sed_rename(ctx, 's/"c:\\\\\\\\\\\\\\\\programdata\\\\\\\\\\\\\\\\datadog\\\\\\\\\\\\\\\\datadog.yaml"/"c:\\\\\\\\\\\\\\\\programdata\\\\\\\\\\\\\\\\StackState\\\\\\\\\\\\\\\\stackstate.yaml"/g',
#                   "./pkg/trace/flags/flags_windows.go")
    do_sed_rename(ctx, 's/"Datadog"/"StackState"/g',
                  "./pkg/util/winutil/shutil.go")
    do_sed_rename(ctx, 's/"Datadog Agent"/"StackState Agent"/g',
                  "./pkg/util/winutil/shutil.go")
    do_sed_rename(ctx, 's/"SOFTWARE\\\\\\\\\\\\\\\\Datadog\\\\\\\\\\\\\\\\"/"SOFTWARE\\\\\\\\\\\\\\\\StackState\\\\\\\\\\\\\\\\"/g',
                  "./pkg/util/winutil/shutil.go")

    # ApiKeys
    dd_api_key_repl = '"\\"DD-Api-Key\\" -> \\"sts-api-key\\""'
    dd_api_key_repl_upper = '"\\"DD-API-KEY\\" -> \\"sts-api-key\\""'
    do_go_rename(
        ctx, dd_api_key_repl, "./cmd/process-agent")
    do_go_rename(
        ctx, dd_api_key_repl, "./comp/forwarder")
    do_go_rename(
        ctx, dd_api_key_repl, "./pkg/trace/api")
    do_go_rename(
        ctx, dd_api_key_repl, "./pkg/trace/writer")
    do_go_rename(
        ctx, dd_api_key_repl_upper, "./cmd/process-agent")
    do_go_rename(
        ctx, dd_api_key_repl_upper, "./comp/forwarder")
    do_go_rename(
        ctx, dd_api_key_repl_upper, "./pkg/trace/api")
    do_go_rename(
        ctx, dd_api_key_repl_upper, "./pkg/trace/writer")

    # Commands
    do_sed_rename(ctx, sts_lower_replace, "./cmd/agent/subcommands/integrations/command.go")
    do_sed_rename(ctx, sts_lower_replace, "./cmd/agent/subcommands/run/dependent_services_windows.go")
    do_sed_rename(ctx, sts_lower_replace, "./cmd/agent/subcommands/launchgui/command.go")
    do_sed_rename(ctx, 's/Datadog Agent/StackState Agent/g', "./cmd/agent/subcommands/launchgui/command.go")
    do_sed_rename(ctx, 's/Datadog Agent/StackState Agent/g', "./cmd/agent/subcommands/run/command.go")
    do_sed_rename(ctx, sts_lower_replace, "./cmd/agent/command/command.go")
    do_sed_rename(ctx, sts_lower_replace, "./cmd/agent/subcommands/integrations/command.go")
    do_sed_rename(ctx, 's/Datadog integration/StackState integration/g', "./cmd/agent/subcommands/integrations/command.go")
    do_go_rename(ctx, '"\\"Collect a flare and send it to Datadog\\" -> \\"Collect a flare and send it to StackState\\""', "./pkg/cli/subcommands/dcaflare")
    do_go_rename(ctx, '"\\"Collect a flare and send it to Datadog\\" -> \\"Collect a flare and send it to StackState\\""', "./cmd/agent/subcommands/flare")
    do_go_rename(ctx, '"\\"Collect a flare and send it to Datadog\\" -> \\"Collect a flare and send it to StackState\\""', "./cmd/security-agent/subcommands/flare")

    # Trace agent
    do_go_rename(ctx, '"\\"DD_PROXY_HTTPS\\" -> \\"STS_PROXY_HTTPS\\""', "./pkg/trace")
    do_go_rename(ctx, '"\\"DD_CONNECTION_LIMIT\\" -> \\"STS_CONNECTION_LIMIT\\""', "./pkg/trace")
    do_go_rename(ctx, '"\\"DD_APM_CONNECTION_LIMIT\\" -> \\"STS_APM_CONNECTION_LIMIT\\""', "./pkg/trace")
    do_go_rename(ctx, '"\\"DD_RECEIVER_PORT\\" -> \\"STS_RECEIVER_PORT\\""', "./pkg/trace")
    do_go_rename(ctx, '"\\"DD_APM_RECEIVER_PORT\\" -> \\"STS_APM_RECEIVER_PORT\\""', "./pkg/trace")
    do_go_rename(ctx, '"\\"DD_MAX_EPS\\" -> \\"STS_MAX_EPS\\""', "./pkg/trace")
    do_go_rename(ctx, '"\\"DD_MAX_TPS\\" -> \\"STS_MAX_TPS\\""', "./pkg/trace")
    do_go_rename(ctx, '"\\"DD_APM_MAX_TPS\\" -> \\"STS_APM_MAX_TPS\\""', "./pkg/trace")
    do_go_rename(ctx, '"\\"DD_IGNORE_RESOURCE\\" -> \\"STS_IGNORE_RESOURCE\\""', "./pkg/trace")
    do_go_rename(ctx, '"\\"DD_APM_IGNORE_RESOURCES\\" -> \\"STS_APM_IGNORE_RESOURCES\\""', "./pkg/trace")
    do_go_rename(ctx, '"\\"DD_API_KEY\\" -> \\"STS_API_KEY\\""', "./pkg/trace")
    do_go_rename(ctx, '"\\"DD_SITE\\" -> \\"STS_SITE\\""', "./pkg/trace")
    do_go_rename(ctx, '"\\"DD_APM_ENABLED\\" -> \\"STS_APM_ENABLED\\""', "./pkg/trace")
    do_go_rename(ctx, '"\\"DD_APM_ENABLED\\" -> \\"STS_APM_ENABLED\\""', "./pkg/config")
    do_go_rename(ctx, '"\\"DD_APM_DD_URL\\" -> \\"STS_APM_URL\\""', "./pkg/trace")
    do_go_rename(ctx, '"\\"DD_APM_DD_URL\\" -> \\"STS_APM_URL\\""', "./pkg/config")
    do_go_rename(ctx, '"\\"DD_HOSTNAME\\" -> \\"STS_HOSTNAME\\""', "./pkg/trace")
    do_go_rename(ctx, '"\\"DD_BIND_HOST\\" -> \\"STS_BIND_HOST\\""', "./pkg/trace")
    do_go_rename(ctx, '"\\"DD_DOGSTATSD_PORT\\" -> \\"STS_DOGSTATSD_PORT\\""', "./pkg/trace")
    do_go_rename(ctx, '"\\"DD_APM_NON_LOCAL_TRAFFIC\\" -> \\"STS_APM_NON_LOCAL_TRAFFIC\\""', "./pkg/trace")
    do_go_rename(ctx, '"\\"DD_LOG_LEVEL\\" -> \\"STS_LOG_LEVEL\\""', "./pkg/trace")
    do_go_rename(ctx, '"\\"DD_APM_ANALYZED_SPANS\\" -> \\"STS_APM_ANALYZED_SPANS\\""', "./pkg/trace")
    do_go_rename(ctx, '"\\"DD_APM_MAX_EPS\\" -> \\"STS_APM_MAX_EPS\\""', "./pkg/trace")
    do_go_rename(ctx, '"\\"DD_APM_ENV\\" -> \\"STS_APM_ENV\\""', "./pkg/trace")
    do_go_rename(ctx, '"\\"DD_APM_MAX_MEMORY\\" -> \\"STS_APM_MAX_MEMORY\\""', "./pkg/trace")
    do_go_rename(ctx, '"\\"datadog.trace_agent.sampler.exception.hits\\" -> \\"stackstate.trace_agent.sampler.exception.hits\\""', "./pkg/trace/sampler")
    do_go_rename(ctx, '"\\"datadog.trace_agent.sampler.exception.misses\\" -> \\"stackstate.trace_agent.sampler.exception.misses\\""', "./pkg/trace/sampler")
    do_go_rename(ctx, '"\\"datadog.trace_agent.sampler.exception.shrinks\\" -> \\"stackstate.trace_agent.sampler.exception.shrinks\\""', "./pkg/trace/sampler")
    do_go_rename(ctx, '"\\"datadog.trace_agent.sampler.kept\\" -> \\"stackstate.trace_agent.sampler.kept\\""', "./pkg/trace/sampler")
    do_go_rename(ctx, '"\\"datadog.trace_agent.sampler.seen\\" -> \\"stackstate.trace_agent.sampler.seen\\""', "./pkg/trace/sampler")
#     do_go_rename(ctx, '"\\"datadog.trace_agent.obfuscations\\" -> \\"stackstate.trace_agent.obfuscations\\""', "./pkg/trace/obfuscate")
    do_go_rename(ctx, '"\\"datadog.agent.python.version\\" -> \\"stackstate.agent.python.version\\""', "./pkg/collector/python")
    do_go_rename(ctx, '"\\"/var/log/datadog/trace-agent.log\\" -> \\"/var/log/stackstate-agent/trace-agent.log\\""', "./pkg/trace/config/")
    do_go_rename(ctx, '"\\"/opt/datadog-agent/embedded/bin/python\\" -> \\"/opt/stackstate-agent/embedded/bin/python\\""', "./pkg/trace/config/")
    do_go_rename(ctx, '"\\"PYTHONPATH=/opt/datadog-agent/agent\\" -> \\"PYTHONPATH=/opt/stackstate-agent/agent\\""', "./pkg/trace/config/")
    do_go_rename(ctx, '"\\"/var/log/datadog/agent.log\\" -> \\"/var/log/stackstate-agent/agent.log\\""', "./pkg/trace/config/")
    do_go_rename(ctx, '"\\"/opt/datadog-agent/bin/agent/agent\\" -> \\"/opt/stackstate-agent/bin/agent/agent\\""', "./pkg/trace/config/")
    do_go_rename(ctx, '"\\"/etc/dd-agent/datadog.conf\\" -> \\"/etc/sts-agent/stackstate.conf\\""', "./pkg/trace/config/")
    do_go_rename(ctx, '"\\"https://trace.agent.datadoghq.com\\" -> \\"http://localhost:7077\\""', "./pkg/trace/config/")

    do_go_rename(ctx, '"\\"Datadog Trace Agent\\" -> \\"Stackstate Trace Agent\\""', "./pkg/trace/writer/")
    do_go_rename(ctx, '"\\"https://github.com/StackVista/datadog-trace-agent\\" -> \\"https://github.com/Stackvista/stackstate-trace-agent\\""', "./pkg/trace/writer/")

    # ntp core check with default hosts replacecment
    do_go_rename(ctx, '"\\"0.datadog.pool.ntp.org\\" -> \\"0.stackstate.pool.ntp.org\\""', "./pkg/collector/corechecks/net/ntp.go")
    do_go_rename(ctx, '"\\"1.datadog.pool.ntp.org\\" -> \\"1.stackstate.pool.ntp.org\\""', "./pkg/collector/corechecks/net/ntp.go")
    do_go_rename(ctx, '"\\"2.datadog.pool.ntp.org\\" -> \\"2.stackstate.pool.ntp.org\\""', "./pkg/collector/corechecks/net/ntp.go")
    do_go_rename(ctx, '"\\"3.datadog.pool.ntp.org\\" -> \\"3.stackstate.pool.ntp.org\\""', "./pkg/collector/corechecks/net/ntp.go")

    # Trace agent
    apm_dd_url_replace = 's/apm_dd_url/apm_sts_url/g'
    do_find_sed_rename_pattern_multi_ignore(ctx, "*.go", apm_dd_url_replace, ".", [".git", "vendor"])
#     do_sed_rename(ctx, apm_dd_url_replace, "./pkg/trace/config/apply.go")
#     do_sed_rename(ctx, apm_dd_url_replace, "./pkg/trace/config/env.go")
#     do_sed_rename(ctx, apm_dd_url_replace, "./pkg/config/apm.go")
    do_find_sed_rename_pattern_multi_ignore(ctx, "*.go", 's/DD_APM_ENABLED/STS_APM_ENABLED/g', ".", [".git", "vendor"])
#     do_sed_rename(ctx, 's/DD_APM_ENABLED/STS_APM_ENABLED/g', "./pkg/trace/agent/run.go")
#     do_sed_rename(ctx, 's/DD_APM_ENABLED/STS_APM_ENABLED/g', "./pkg/config/apm.go")
    dd_agent_bin_replace = 's/dd_agent_bin/sts_agent_bin/g'
#     do_sed_rename(ctx, dd_agent_bin_replace, "./pkg/trace/config/apply.go")
    do_find_sed_rename_pattern_multi_ignore(ctx, "*.go", dd_agent_bin_replace, ".", [".git", "vendor"])
    DD_API_KEY_replace = 's/DD_API_KEY/STS_API_KEY/g'
#     do_sed_rename(ctx, DD_API_KEY_replace, "./pkg/trace/config/config.go")
    do_find_sed_rename_pattern_multi_ignore(ctx, "*.go", DD_API_KEY_replace, ".", [".git", "vendor"])
    DD_HOSTNAME_replace = 's/DD_HOSTNAME/STS_HOSTNAME/g'
    DD_HEADER_replace = 's/X-Datadog/X-Stackstate/g'
    do_find_sed_rename_pattern_multi_ignore(ctx, "*.go", DD_HOSTNAME_replace, ".", [".git", "vendor"])
    do_find_sed_rename_pattern_multi_ignore(ctx, "*.go", DD_HEADER_replace, ".", [".git", "vendor"])
#     do_sed_rename(ctx, DD_HOSTNAME_replace, "./pkg/trace/config/config.go")
#     do_sed_rename(ctx, DD_HEADER_replace, "./pkg/trace/api/profiles.go")
#     do_sed_rename(ctx, DD_HEADER_replace, "./pkg/trace/api/api.go")
#     do_sed_rename(ctx, DD_HEADER_replace, "./pkg/trace/test/runner.go")
#     do_sed_rename(ctx, DD_HEADER_replace, "./pkg/trace/writer/stats_test.go")
#     do_sed_rename(ctx, DD_HEADER_replace, "./pkg/trace/writer/trace.go")
    apm_dd_tags_replace = 's/"_dd/"_sts/g'
    do_sed_rename(ctx, apm_dd_tags_replace, "./pkg/trace/agent/agent.go")

    # pkg/trace/config/config_windows.go
#     do_sed_rename(ctx, 's/"programdata\\\\\\\\\\\\\\\\datadog"/"programdata\\\\\\\\\\\\\\\\stackstate"/g',
#                   "./pkg/trace/config/config_windows.go")
#     do_sed_rename(ctx, 's/"ProgramData\\\\\\\\\\\\\\\\datadog"/"ProgramData\\\\\\\\\\\\\\\\StackState"/g',
#                   "./pkg/trace/config/config_windows.go")
#     do_sed_rename(ctx, 's/"Datadog"/"Stackstate"/g',
#                   "./pkg/trace/config/config_windows.go")
#     do_sed_rename(ctx, 's/"ProgramData\\\\\\\\\\\\\\\\StackVista"/"ProgramData\\\\\\\\\\\\\\\\StackState"/g',
#                   "./pkg/trace/config/config_windows.go")
    do_find_sed_rename_pattern_multi_ignore(ctx, "*.go", 's/"datadog.conf"/"stackstate.conf"/g', ".", [".git", "vendor"])
#     do_sed_rename(ctx, 's/"datadog.conf"/"stackstate.conf"/g',
#                   "./pkg/trace/config/config_windows.go")
#     do_sed_rename(ctx, 's/"c:\\\\\\\\\\\\\\\\programdata\\\\\\\\\\\\\\\\datadog\\\\\\\\\\\\\\\\logs\\\\\\\\\\\\\\\\process-agent.log"/"c:\\\\\\\\\\\\\\\\programdata\\\\\\\\\\\\\\\\stackstate\\\\\\\\\\\\\\\\logs\\\\\\\\\\\\\\\\process-agent.log"/g',
#                   "./pkg/process/config/config_windows.go")
    do_find_sed_rename_pattern_multi_ignore(ctx, "*.go", 's/"c:\\\\\\\\\\\\\\\\Program Files\\\\\\\\\\\\\\\\Datadog\\\\\\\\\\\\\\\\Datadog Agent\\\\\\\\\\\\\\\\bin\\\\\\\\\\\\\\\\agent.exe"/"c:\\\\\\\\\\\\\\\\Program Files\\\\\\\\\\\\\\\\StackState\\\\\\\\\\\\\\\\StackState Agent\\\\\\\\\\\\\\\\bin\\\\\\\\\\\\\\\\agent.exe"/g',
                                            ".", [".git", "vendor"])
#     do_sed_rename(ctx, 's/"c:\\\\\\\\\\\\\\\\Program Files\\\\\\\\\\\\\\\\Datadog\\\\\\\\\\\\\\\\Datadog Agent\\\\\\\\\\\\\\\\bin\\\\\\\\\\\\\\\\agent.exe"/"c:\\\\\\\\\\\\\\\\Program Files\\\\\\\\\\\\\\\\StackState\\\\\\\\\\\\\\\\StackState Agent\\\\\\\\\\\\\\\\bin\\\\\\\\\\\\\\\\agent.exe"/g',
#                   "./pkg/process/config/config_windows.go")
    do_find_sed_rename_pattern_multi_ignore(ctx, "*.go", 's/"c:\\\\\\\\\\\\\\\\Program Files\\\\\\\\\\\\\\\\Datadog\\\\\\\\\\\\\\\\Datadog Agent\\\\\\\\\\\\\\\\bin\\\\\\\\\\\\\\\\agent.exe"/"c:\\\\\\\\\\\\\\\\Program Files\\\\\\\\\\\\\\\\StackState\\\\\\\\\\\\\\\\StackState Agent\\\\\\\\\\\\\\\\bin\\\\\\\\\\\\\\\\agent.exe"/g',
                                            ".", [".git", "vendor"])
#     do_sed_rename(ctx, 's/c:\\\\\\\\\\\\\\\\programdata\\\\\\\\\\\\\\\\datadog/c:\\\\\\\\\\\\\\\\programdata\\\\\\\\\\\\\\\\StackState/g',
#                   "./pkg/config/setup/config_windows.go")
#     do_sed_rename(ctx, 's/datadog\\\\\\\\\\\\\\\\logs\\\\\\\\\\\\\\\\trace-agent.log"/StackState\\\\\\\\\\\\\\\\logs\\\\\\\\\\\\\\\\trace-agent.log"/g',
#                   "./pkg/trace/config/config_windows.go")
    do_find_sed_rename_pattern_multi_ignore(ctx, "*.go", 's/datadog\\\\\\\\\\\\\\\\logs\\\\\\\\\\\\\\\\trace-agent.log"/StackState\\\\\\\\\\\\\\\\logs\\\\\\\\\\\\\\\\trace-agent.log"/g',
                                            ".", [".git", "vendor"])
#     do_sed_rename(ctx, 's/Datadog\\\\\\\\\\\\\\\\Datadog Agent\\\\\\\\\\\\\\\\bin\\\\\\\\\\\\\\\\agent.exe"/StackState\\\\\\\\\\\\\\\\StackState Agent\\\\\\\\\\\\\\\\bin\\\\\\\\\\\\\\\\agent.exe"/g',
#                   "./pkg/trace/config/config_windows.go")


    # Trace Agent Metrics # sts - stop caring about the trace agent.
#     do_sed_rename(ctx, datadog_metrics_replace, "./pkg/trace/api/api.go")
#     do_sed_rename(ctx, datadog_metrics_replace, "./pkg/trace/api/responses.go")
#     do_sed_rename(ctx, datadog_metrics_replace, "./pkg/trace/api/listener.go")
#     do_sed_rename(ctx, datadog_metrics_replace, "./pkg/trace/sampler/rare_sampler.go")
#     do_sed_rename(ctx, datadog_metrics_replace, "./pkg/trace/agent/run.go")
#     do_sed_rename(ctx, datadog_metrics_replace, "./pkg/trace/agent/agent.go")
#     do_go_rename(ctx, '"\\"datadog.conf\\" -> \\"stackstate.conf\\""', "./pkg/trace/agent")
#     do_sed_rename(ctx, datadog_metrics_replace, "./pkg/trace/event/sampler_max_eps.go")
#     do_sed_rename(ctx, datadog_metrics_replace, "./pkg/trace/writer/trace.go")
#     do_sed_rename(ctx, datadog_metrics_replace, "./pkg/trace/writer/stats.go")
#     do_sed_rename(ctx, datadog_metrics_replace, "./pkg/trace/writer/stats_test.go")
#     do_sed_rename(ctx, datadog_metrics_replace, "./pkg/trace/info/stats.go")
#     # do_sed_rename(ctx, datadog_metrics_replace, "./pkg/process/statsd/statsd.go")
#     do_sed_rename(ctx, 's/"Datadog Trace Agent\/%s\/%s"/"Stackstate Trace Agent-%s-%s"/g',
#                   "./pkg/trace/writer/sender.go")

    # Defaults
    do_go_rename(ctx, '"\\"/etc/datadog-agent\\" -> \\"/etc/stackstate-agent\\""', "./cmd/agent/common")
    do_go_rename(ctx, '"\\"/var/log/datadog/agent.log\\" -> \\"/var/log/stackstate-agent/agent.log\\""', "./cmd/agent/common")
    do_go_rename(ctx, '"\\"/var/log/datadog/cluster-agent.log\\" -> \\"/var/log/stackstate-agent/cluster-agent.log\\""', "./cmd/agent/common")
    do_go_rename(ctx, '"\\"datadog.yaml\\" -> \\"stackstate.yaml\\""', "./cmd/agent")
    do_go_rename(ctx, '"\\"Datadog.yaml\\" -> \\"stackstate.yaml\\""', "./cmd/agent")
    do_go_rename(ctx, '"\\"datadog.yaml\\" -> \\"stackstate.yaml\\""', "./pkg/config")
    do_go_rename(ctx, '"\\"datadog.conf\\" -> \\"stackstate.conf\\""', "./cmd/agent")
    do_go_rename(ctx, '"\\"path to directory containing datadog.yaml\\" -> \\"path to directory containing stackstate.yaml\\""', "./cmd")
    do_go_rename(ctx, '"\\"unable to load Datadog config file: %s\\" -> \\"unable to load StackState config file: %s\\""', "./cmd/agent/common")
    do_go_rename(ctx, '"\\"unable to load Datadog config file: %w\\" -> \\"unable to load StackState config file: %w\\""', "./cmd/agent/common")
#     do_go_rename(ctx, '"\\"Starting Datadog Agent v%v\\" -> \\"Starting StackState Agent v%v\\""', "./cmd/agent/app")

    # Dist config templates
    do_sed_rename(ctx, sts_lower_replace, "./cmd/agent/dist/conf.d/go_expvar.d/agent_stats.yaml.example")
    do_sed_rename(ctx, sts_lower_replace, "./cmd/agent/dist/conf.d/apm.yaml.default")
    do_sed_rename(ctx, 's/dd/sts/g', "./cmd/agent/dist/dd-agent")
    do_sed_rename(ctx, sts_lower_replace, "./cmd/agent/dist/dd-agent")

    # Hardcoded checks and metrics
    do_sed_rename(ctx, sts_lower_replace, "./pkg/aggregator/aggregator.go")

    # Windows defaults
    do_find_sed_rename_pattern_multi_ignore(ctx, "*.rc", sts_camel_replace,
                                                ".", [".git", "vendor"])
#     do_sed_rename(ctx, sts_camel_replace, "./cmd/agent/windows_resources/agent.rc")
#     do_sed_rename(ctx, sts_camel_replace, "./cmd/trace-agent/windows/resources/trace-agent.rc")
    do_find_sed_rename_pattern_multi_ignore(ctx, "*install_service_windows.go", sts_camel_replace,
                                                    ".", [".git", "vendor"])
#     do_sed_rename(ctx, sts_camel_replace, "./cmd/agent/app/install_service_windows.go")
#     do_sed_rename(ctx, sts_lower_replace, "./cmd/agent/app/dependent_services_windows.go")
    # replace strings NOT containing certain pattern
    do_sed_rename(ctx, '/config/! s/Data[dD]og/StackState/g', "./cmd/agent/common/common_windows.go")
    do_sed_rename(ctx, sts_lower_replace, "./cmd/agent/common/common_windows.go")
    do_sed_rename(ctx, 's/dd_url/sts_url/', "./cmd/agent/common/common_windows.go")
    do_sed_rename(ctx, sts_lower_replace, "./cmd/dogstatsd/main_windows.go")
    do_sed_rename(ctx, sts_camel_replace, "./pkg/config/setup/config_windows.go")

    # Windows MSI installation
    do_sed_rename(ctx, sts_camel_replace, "./omnibus/resources/agent/msi/localization-en-us.wxl.erb")
    do_sed_rename(ctx, 's/"datadog\.yaml\.example"/"stackstate\.yaml\.example"/', "./omnibus/resources/agent/msi/source.wxs.erb")
    do_sed_rename(ctx, 's/datadoghq\.com/www\.stackstate\.com/', "./omnibus/resources/agent/msi/source.wxs.erb")
    do_sed_rename(ctx, sts_camel_replace, "./omnibus/resources/agent/msi/source.wxs.erb")
    do_sed_rename(ctx, sts_lower_replace, "./omnibus/resources/agent/msi/source.wxs.erb")
    do_sed_rename(ctx, 's/DATADOG/STACKSTATE/', "./omnibus/resources/agent/msi/source.wxs.erb")
    do_sed_rename(ctx, 's/dd_url/sts_url/', "./omnibus/resources/agent/msi/source.wxs.erb")
    do_sed_rename(ctx, 's/\[.*DD_URL\]/\[STS_URL\]/', "./omnibus/resources/agent/msi/source.wxs.erb")
    do_sed_rename(ctx, 's/"SOFTWARE\\\\Datadog\\\\Datadog Agent"/"SOFTWARE\\\\StackState\\\\StackState Agent"/g',
                  "./omnibus/resources/iot/msi/source.wxs.erb")
    do_sed_rename(ctx, 's/"Datadog Agent"/"StackState Agent"/g',
                  "./omnibus/resources/iot/msi/source.wxs.erb")
    do_sed_rename(ctx, 's/"datadog.yaml.example"/"stackstate.yaml.example"/g',
                  "./omnibus/resources/iot/msi/source.wxs.erb")
    do_sed_rename(ctx, sts_camel_replace, "./omnibus/resources/agent/msi/bundle.wxs.erb")
    do_sed_rename(ctx, 's/dd_logo_side\\.png/sts_logo_side\\.png/', "./omnibus/resources/agent/msi/bundle.wxs.erb")
#     do_sed_rename(ctx, 's/StackVista Agent/Stackstate Agent/g',
#                   "./tools/windows/install-help/cal/stopservices.cpp")
#     do_sed_rename(ctx, 's/StackVista Trace Agent/Stackstate Trace Agent/g',
#                   "./tools/windows/install-help/cal/stopservices.cpp")
#     do_sed_rename(ctx, 's/"StackVista Process Agent"/"StackState Process Agent"/g',
#                   "./tools/windows/install-help/cal/stopservices.cpp")
#     do_sed_rename(ctx, 's/"Send process metrics to StackVista"/"Send process metrics to StackState"/g',
#                   "./tools/windows/install-help/cal/stopservices.cpp")
#     do_sed_rename(ctx, 's/"Send tracing metrics to StackVista"/"Send tracing metrics to StackState"/g',
#                   "./tools/windows/install-help/cal/stopservices.cpp")
#     do_sed_rename(ctx, 's/"Send metrics to StackVista"/"Send metrics to StackState"/g',
#                   "./tools/windows/install-help/cal/stopservices.cpp")
    do_sed_rename(ctx, 's/StackVista Trace Agent/Stackstate Trace Agent/g',
                  "./cmd/trace-agent/windows/resources/trace-agent.rc")

    # datadogagent references
#     do_sed_rename(ctx, 's/datadogagent/stackstateagent/g',
#                   "./tools/windows/install-help/cal/stopservices.cpp")
    do_sed_rename(ctx, 's/datadogagent/stackstateagent/g',
                  "./Dockerfiles/agent/install.ps1")

    # tools/windows/install-help/uninstall-cmd/cmdline.cpp
#     do_sed_rename(ctx, 's/"C:\\\\\\\\\\\\\\\\Program Files\\\\\\\\\\\\\\\\Datadog\\\\\\\\\\\\\\\\Datadog Agent\\\\\\\\\\\\\\\\"/"C:\\\\\\\\\\\\\\\\Program Files\\\\\\\\\\\\\\\\StackState\\\\\\\\\\\\\\\\StackState Agent\\\\\\\\\\\\\\\\"/g',
#                   "./tools/windows/install-help/install-cmd/cmdline.cpp")
#     do_sed_rename(ctx, 's/"C:\\\\\\\\\\\\\\\\ProgramData\\\\\\\\\\\\\\\\Datadog\\\\\\\\\\\\\\\\"/"C:\\\\\\\\\\\\\\\\ProgramData\\\\\\\\\\\\\\\\StackState\\\\\\\\\\\\\\\\"/g',
#                   "./tools/windows/install-help/install-cmd/cmdline.cpp")
#     do_sed_rename(ctx, 's/"C:\\\\\\\\\\\\\\\\Program Files\\\\\\\\\\\\\\\\Datadog\\\\\\\\\\\\\\\\Datadog Agent\\\\\\\\\\\\\\\\"/"C:\\\\\\\\\\\\\\\\Program Files\\\\\\\\\\\\\\\\StackState\\\\\\\\\\\\\\\\StackState Agent\\\\\\\\\\\\\\\\"/g',
#                   "./tools/windows/install-help/uninstall-cmd/cmdline.cpp")
#     do_sed_rename(ctx, 's/"C:\\\\\\\\\\\\\\\\ProgramData\\\\\\\\\\\\\\\\Datadog\\\\\\\\\\\\\\\\"/"C:\\\\\\\\\\\\\\\\ProgramData\\\\\\\\\\\\\\\\StackState\\\\\\\\\\\\\\\\"/g',
#                   "./tools/windows/install-help/uninstall-cmd/cmdline.cpp")

    # chocolatey/tools-offline/chocolateyinstall.ps1
    do_sed_rename(ctx, 's/"HKLM:\\\\SOFTWARE\\\\Datadog\\\\Datadog Agent"/"HKLM:\\\\SOFTWARE\\\\StackState\\\\StackState Agent"/g',
                  "./chocolatey/tools-online/chocolateyinstall.ps1")
    do_sed_rename(ctx, 's/"HKLM:\\\\SOFTWARE\\\\Datadog\\\\Datadog Agent"/"HKLM:\\\\SOFTWARE\\\\StackState\\\\StackState Agent"/g',
                  "./chocolatey/tools-offline/chocolateyinstall.ps1")

    # Dockerfiles/agent/install.ps1
    do_sed_rename(ctx, 's/C:\\\\Program Files\\\\Datadog\\\\Datadog Agent/C:\\\\Program Files\\\\StackState\\\\StackState Agent/g',
                  "./Dockerfiles/agent/install.ps1")
    do_sed_rename(ctx, 's/C:\/Program Files\/Datadog\/Datadog Agent/C:\/Program Files\/StackState\/StackState Agent/g',
                  "./Dockerfiles/agent/install.ps1")
    do_sed_rename(ctx, 's/C:\/Program Files\/Datadog/C:\/Program Files\/StackState/g',
                  "./Dockerfiles/agent/install.ps1")
    do_sed_rename(ctx, 's/C:\/ProgramData\/Datadog/C:\/ProgramData\/StackState/g',
                  "./Dockerfiles/agent/install.ps1")
    do_sed_rename(ctx, 's/"Datadog Agent"/"StackState Agent"/g',
                  "./Dockerfiles/agent/install.ps1")
    do_sed_rename(ctx, 's/"datadog-process-agent"/"stackstate-process-agent"/g',
                  "./Dockerfiles/agent/install.ps1")
    do_sed_rename(ctx, 's/"datadog-trace-agent"/"stackstate-trace-agent"/g',
                  "./Dockerfiles/agent/install.ps1")

    # tools/windows/install-help/cal/strings.cpp
#     do_sed_rename(ctx, 's/datadog.yaml/stackstate.yaml/g',
#                   "./tools/windows/install-help/cal/strings.cpp")

    # Windows SysTray and GUI
    tray_replace = 's/ddtray/ststray/'
    do_sed_rename(ctx, sts_lower_replace, "./comp/systray/systray/systrayimpl/doservicecontrol.go")
    do_sed_rename(ctx, sts_camel_replace, "./comp/systray/systray/systrayimpl/systray.go")
    do_sed_rename(ctx, tray_replace, "./comp/systray/systray/systrayimpl/systray.go")
    do_sed_rename(ctx, sts_camel_replace, "./cmd/systray/systray.rc")
    do_sed_rename(ctx, tray_replace, "./cmd/systray/systray.rc")
    do_sed_rename(ctx, tray_replace, "./omnibus/resources/agent/msi/source.wxs.erb")
    do_sed_rename(ctx, tray_replace, "./tasks/systray.py")
    do_sed_rename(ctx, tray_replace, "./omnibus/config/projects/agent.rb")
    do_sed_rename(ctx, tray_replace, "./omnibus/config/software/datadog-agent.rb")
    do_sed_rename(ctx, sts_lower_replace, "./cmd/agent/gui/views/templates/index.tmpl")
    do_sed_rename(ctx, 's/"StackVista Agent 6"/"StackState Agent 2"/', "./cmd/agent/gui/views/templates/index.tmpl")
    do_sed_rename(ctx, sts_camel_replace, "./cmd/agent/gui/views/templates/index.tmpl")
    do_sed_rename(ctx, sts_camel_replace, "./cmd/agent/gui/views/private/js/javascript.js")

    # TODO: overbranding - fix either dll name or reference
    do_sed_rename(ctx, 's/libstackstate-agent-three/libdatadog-agent-three/g', "./omnibus/resources/agent/msi/source.wxs.erb")
    do_sed_rename(ctx, 's/libstackstate-agent-two/libdatadog-agent-two/g', "./omnibus/resources/agent/msi/source.wxs.erb")

    # stackstate_checks
    do_find_sed_rename_pattern_multi_ignore(ctx, "*.go", 's/datadog_checks/stackstate_checks/g',
                                                ".", [".git", "vendor"])
#     do_go_rename(ctx, '"\\"datadog_checks\\" -> \\"stackstate_checks\\""', "./cmd/agent/app")
#     do_sed_rename(ctx, 's/datadog_checks_base/stackstate_checks_base/g', "./cmd/agent/app/integrations.go")
#     do_go_rename(ctx, '"\\"datadog_checks\\" -> \\"stackstate_checks\\""', "./pkg/collector/python")
    do_go_rename(ctx, '"\\"An error occurred while grabbing the python datadog integration list\\" -> \\"An error occurred while grabbing the python StackState integration list\\""', "./pkg/collector/python")
    #    do_sed_rename(ctx, datadog_checks_replace, "./pkg/collector/python/loader.go")
    do_sed_rename(ctx, datadog_metrics_replace, "./pkg/collector/runner/runner.go")
    do_sed_rename(ctx, datadog_metrics_replace, "./pkg/collector/worker/worker.go")

    # cluster agent client
    do_go_rename(ctx, '"\\"datadog-cluster-agent\\" -> \\"stackstate-cluster-agent\\""', "./pkg/config")
    do_sed_rename(ctx, 's/Datadog Cluster Agent/StackState Cluster Agent/g', "./pkg/util/clusteragent/clusteragent.go")
    do_sed_rename(ctx, 's/Datadog Cluster Agent/StackState Cluster Agent/g', "./pkg/status/render/templates/clusteragent.tmpl")

    # kubernetes openmetrics annotations
    do_sed_rename(ctx, 's/ad.datadoghq.com/ad.stackstate.com/g', "./pkg/autodiscovery/common/utils/kubelet.go")
    do_sed_rename(ctx, 's/ad.datadoghq.com/ad.stackstate.com/g', "./pkg/autodiscovery/listeners/kubelet.go")
    do_sed_rename(ctx, 's/ad.datadoghq.com/ad.stackstate.com/g', "./pkg/autodiscovery/listeners/kube_services.go")
    do_sed_rename(ctx, 's/ad.datadoghq.com/ad.stackstate.com/g', "./pkg/autodiscovery/listeners/kube_endpoints.go")
    do_sed_rename(ctx, 's/ad.datadoghq.com/ad.stackstate.com/g', "./pkg/autodiscovery/providers/kube_common.go")
#     do_sed_rename(ctx, 's/ad.datadoghq.com/ad.stackstate.com/g', "./pkg/autodiscovery/providers/kubelet.go")
    do_sed_rename(ctx, 's/ad.datadoghq.com/ad.stackstate.com/g', "./pkg/autodiscovery/providers/kube_services.go")
    do_sed_rename(ctx, 's/ad.datadoghq.com/ad.stackstate.com/g', "./pkg/autodiscovery/providers/kube_endpoints.go")
    do_sed_rename(ctx, 's/ad.datadoghq.com/ad.stackstate.com/g', "./pkg/tagger/collectors/workloadmeta_extract.go")
    do_sed_rename(ctx, 's/ad.datadoghq.com/ad.stackstate.com/g', "./pkg/util/kubernetes/kubelet/kubelet.go")

    # docker/ecs openmetrics annotations
    do_sed_rename(ctx, 's/com.datadoghq.ad/com.stackstate.ad/g', "./pkg/autodiscovery/listeners/common.go")
    do_sed_rename(ctx, 's/com.datadoghq.ad/com.stackstate.ad/g', "./pkg/autodiscovery/providers/container.go")

    # rtloader branding
    do_sed_rename(ctx, datadog_checks_replace, "./rtloader/two/two.cpp")
    do_sed_rename(ctx, datadog_checks_replace, "./rtloader/three/three.cpp")

    # omnibus
    do_sed_rename(ctx, 's/\/opt\/datadog/\/opt\/stackstate/g', "./omnibus/config/projects/agent.rb")
    do_sed_rename(ctx, 's/\\\\\\\\\\\\\\\\etc\\\\\\\\\\\\\\\\datadog-agent\\\\\\\\\\\\\\\\extra_package_files/\\\\\\\\\\\\\\\\etc\\\\\\\\\\\\\\\\stackstate-agent\\\\\\\\\\\\\\\\extra_package_files/g',
                  "./omnibus/config/projects/agent.rb")
    do_sed_rename(ctx, 's/\\\\\\\\\\\\\\\\etc\\\\\\\\\\\\\\\\datadog-agent\\\\\\\\\\\\\\\\extra_package_files/\\\\\\\\\\\\\\\\etc\\\\\\\\\\\\\\\\stackstate-agent\\\\\\\\\\\\\\\\extra_package_files/g',
                  "./omnibus/config/projects/agent-binaries.rb")
    do_sed_rename(ctx, 's/\\\\\\\\\\\\\\\\etc\\\\\\\\\\\\\\\\datadog-agent\\\\\\\\\\\\\\\\extra_package_files/\\\\\\\\\\\\\\\\etc\\\\\\\\\\\\\\\\stackstate-agent\\\\\\\\\\\\\\\\extra_package_files/g',
                  "./omnibus/config/projects/iot-agent.rb")
    do_sed_rename(ctx, 's/StackVista\\\\\\\\\\\\\\\\stackstate-agent/Stackvista\\\\\\\\\\\\\\\\stackstate-agent/g', "./omnibus/config/projects/agent.rb")
    do_sed_rename(ctx, 's/\/opt\/datadog/\/opt\/stackstate/g', "./omnibus/config/projects/iot-agent.rb")
    do_sed_rename(ctx, 's/\/opt\/datadog/\/opt\/stackstate/g', "./omnibus/config/software/datadog-agent-finalize.rb")
    do_sed_rename(ctx, 's/DataDog\/datadog-agent/StackVista\/stackstate-agent/g', "./omnibus/config/software/datadog-cf-finalize.rb")
    do_sed_rename(ctx, 's/\/opt\/datadog/\/opt\/stackstate/g', "./omnibus/config/templates/datadog-agent/sysvinit_debian.erb")

    do_find_sed_rename_pattern_multi_ignore(ctx, "*postinst", 's/\/opt\/datadog/\/opt\/stackstate/g',
                                                    ".", [".git", "vendor"])
    do_find_sed_rename_pattern_multi_ignore(ctx, "*posttrans", 's/\/opt\/datadog/\/opt\/stackstate/g',
                                                    ".", [".git", "vendor"])
    do_find_sed_rename_pattern_multi_ignore(ctx, "*preinst", 's/\/opt\/datadog/\/opt\/stackstate/g',
                                                    ".", [".git", "vendor"])
    do_find_sed_rename_pattern_multi_ignore(ctx, "*postrm", 's/\/opt\/datadog/\/opt\/stackstate/g',
                                                    ".", [".git", "vendor"])

#     do_sed_rename(ctx, 's/\/opt\/datadog/\/opt\/stackstate/g', "./omnibus/package-scripts/dogstatsd/postinst")
#     do_sed_rename(ctx, 's/\/opt\/datadog/\/opt\/stackstate/g', "./omnibus/package-scripts/dogstatsd/posttrans")
#     do_sed_rename(ctx, 's/\/opt\/datadog/\/opt\/stackstate/g', "./omnibus/package-scripts/dogstatsd/preinst")
#     do_sed_rename(ctx, 's/\/opt\/datadog/\/opt\/stackstate/g', "./omnibus/package-scripts/iot-agent/postinst")
#     do_sed_rename(ctx, 's/\/opt\/datadog/\/opt\/stackstate/g', "./omnibus/package-scripts/iot-agent/postrm")
#     do_sed_rename(ctx, 's/\/opt\/datadog/\/opt\/stackstate/g', "./omnibus/package-scripts/iot-agent/posttrans")
#     do_sed_rename(ctx, 's/\/opt\/datadog/\/opt\/stackstate/g', "./omnibus/package-scripts/iot-agent/preinst")
    do_sed_rename(ctx, 's/datadog\.yaml/stackstate\.yaml/', "./omnibus/config/software/datadog-iot-agent.rb")

    do_sed_rename(ctx, 's/datadog\.yaml/stackstate\.yaml/', "./omnibus/config/templates/datadog-agent/systemd.process.service.erb")
    do_sed_rename(ctx, 's/datadog/stackstate/', "./omnibus/config/templates/datadog-agent/systemd.process.service.erb")

    do_sed_rename(ctx, 's/datadog\.yaml/stackstate\.yaml/', "./omnibus/config/templates/datadog-agent/systemd.security.service.erb")
    do_sed_rename(ctx, 's/datadog/stackstate/', "./omnibus/config/templates/datadog-agent/systemd.security.service.erb")

    do_sed_rename(ctx, 's/datadog\.yaml/stackstate\.yaml/', "./omnibus/config/templates/datadog-agent/systemd.sysprobe.service.erb")
    do_sed_rename(ctx, 's/datadog/stackstate/', "./omnibus/config/templates/datadog-agent/systemd.sysprobe.service.erb")

    do_sed_rename(ctx, 's/datadog-agent/stackstate-agent/', "./omnibus/config/software/datadog-iot-agent.rb")
    do_sed_rename(ctx, 's/StackVista/Stackvista/', "./omnibus/config/software/datadog-iot-agent.rb")
    do_sed_rename(ctx, 's/\/var\/log\/datadog/\/var\/log\/stackstate/', "./omnibus/config/software/datadog-iot-agent.rb")
    do_sed_rename(ctx, 's/datadog-iot-agent\/src/stackstate-iot-agent\/src/', "./omnibus/config/software/datadog-iot-agent.rb")
    do_sed_rename(ctx, 's/DataDog\/datadog-agent\/tools\/windows\/decompress_merge_module.ps1/StackVista\/stackstate-agent\/tools\/windows\/decompress_merge_module.ps1/',
                  "./omnibus/config/software/vc_redist_14.rb")
    do_sed_rename(ctx, 's/DataDog\/datadog-agent/StackVista\/stackstate-agent/',
                  "./omnibus/config/software/vc_redist.rb")
    do_sed_rename(ctx, 's/DataDog\/datadog-agent\/bin\/agent/StackVista\/stackstate-agent\/bin\/agent/',
                  "./omnibus/config/software/datadog-agent.rb")
    do_sed_rename(ctx, 's/\/etc\/datadog-agent/\/etc\/stackstate-agent/',
                  "./omnibus/config/software/datadog-agent.rb")
    do_sed_rename(ctx, 's/datadog-agent\/src\/github\.com\/DataDog\/datadog-agent\/rtloader/stackstate-agent\/src\/github\.com\/StackVista\/stackstate-agent\/rtloader/',
                  "./omnibus/config/software/datadog-agent.rb")
    do_sed_rename(ctx, 's/"datadog.yaml"/"stackstate.yaml"/g',
                  "./omnibus/config/projects/agent.rb")
    do_sed_rename(ctx, 's/opt\\\\datado-agent/opt\\\\stackstate-agent/',
                  "./omnibus/resources/iot/msi/localbuild/rebuild.bat")
    do_sed_rename(ctx, 's/opt\\\\datadog-agent/opt\\\\stackstate-agent/',
                  "./omnibus/resources/agent/msi/localbuild/rebuild.bat")
    do_sed_rename(ctx, 's/src\\\\etc\\\\datadog-agent/src\\\\etc\\\\stackstate-agent/',
                  "./omnibus/resources/iot/msi/localbuild/rebuild.bat")
    do_sed_rename(ctx, 's/src\\\\etc\\\\datadog-agent/src\\\\etc\\\\stackstate-agent/',
                  "./omnibus/resources/agent/msi/localbuild/rebuild.bat")


@task
def build(
    ctx,
    rebuild=False,
    race=False,
    build_include=None,
    build_exclude=None,
    flavor=AgentFlavor.base.name,
    development=True,
    skip_assets=False,
    embedded_path=None,
    rtloader_root=None,
    python_home_2=None,
    python_home_3=None,
    major_version='3',
    python_runtimes='3',
    arch='x64',
    exclude_rtloader=False,
    go_mod="vendor",
    windows_sysprobe=False,
    cmake_options='',
):
    """
    Build the agent. If the bits to include in the build are not specified,
    the values from `invoke.yaml` will be used.

    Example invokation:
        inv agent.build --build-exclude=systemd
    """
    flavor = AgentFlavor[flavor]

    if not exclude_rtloader and not flavor.is_iot():
        # If embedded_path is set, we should give it to rtloader as it should install the headers/libs
        # in the embedded path folder because that's what is used in get_build_flags()
        rtloader_make(ctx, python_runtimes=python_runtimes, install_prefix=embedded_path, cmake_options=cmake_options)
        rtloader_install(ctx)

    ldflags, gcflags, env = get_build_flags(
        ctx,
        embedded_path=embedded_path,
        rtloader_root=rtloader_root,
        python_home_2=python_home_2,
        python_home_3=python_home_3,
        major_version=major_version,
        python_runtimes=python_runtimes,
    )

    if sys.platform == 'win32':
        # Important for x-compiling
        env["CGO_ENABLED"] = "1"

        if arch == "x86":
            env["GOARCH"] = "386"

        build_messagetable(ctx, arch=arch)
        vars = versioninfo_vars(ctx, major_version=major_version, python_runtimes=python_runtimes, arch=arch)
        build_rc(
            ctx,
            "cmd/agent/windows_resources/agent.rc",
            arch=arch,
            vars=vars,
            out="cmd/agent/rsrc.syso",
        )

    if flavor.is_iot():
        # Iot mode overrides whatever passed through `--build-exclude` and `--build-include`
        build_tags = get_default_build_tags(build="agent", arch=arch, flavor=flavor)
    else:
        build_include = (
            get_default_build_tags(build="agent", arch=arch, flavor=flavor)
            if build_include is None
            else filter_incompatible_tags(build_include.split(","), arch=arch)
        )
        build_exclude = [] if build_exclude is None else build_exclude.split(",")
        build_tags = get_build_tags(build_include, build_exclude)
        print("-------------------------------------------")
        print(f"Building Agent with tags: {build_tags}")
        print("-------------------------------------------")

    cmd = "go build -mod={go_mod} {race_opt} {build_type} -tags \"{go_build_tags}\" "

    cmd += "-o {agent_bin} -gcflags=\"{gcflags}\" -ldflags=\"{ldflags}\" {REPO_PATH}/cmd/{flavor}"
    args = {
        "go_mod": go_mod,
        "race_opt": "-race" if race else "",
        "build_type": "-a" if rebuild else "",
        "go_build_tags": " ".join(build_tags),
        "agent_bin": os.path.join(BIN_PATH, bin_name("agent")),
        "gcflags": gcflags,
        "ldflags": ldflags,
        "REPO_PATH": REPO_PATH,
        "flavor": "iot-agent" if flavor.is_iot() else "agent",
    }
    if sys.platform.startswith('win'):
        ctx.run("echo %cd%", env=env)
    print("cmd: %s" % cmd.format(**args))
    ctx.run(cmd.format(**args), env=env)

    # Remove cross-compiling bits to render config
    env.update({"GOOS": "", "GOARCH": ""})

    # Render the Agent configuration file template
    build_type = "agent-py3"
    if flavor.is_iot():
        build_type = "iot-agent"
    elif has_both_python(python_runtimes):
        build_type = "agent-py2py3"

    generate_config(ctx,
                    build_type=build_type,
                    output_file="./cmd/agent/dist/stackstate.yaml", # sts
                    env=env)

    # On Linux and MacOS, render the system-probe configuration file template
    if sys.platform != 'win32' or windows_sysprobe:
        generate_config(ctx, build_type="system-probe", output_file="./cmd/agent/dist/system-probe.yaml", env=env)

    if not skip_assets:
        refresh_assets(ctx, build_tags, development=development, flavor=flavor.name, windows_sysprobe=windows_sysprobe)


@task
def refresh_assets(_, build_tags, development=True, flavor=AgentFlavor.base.name, windows_sysprobe=False):
    """
    Clean up and refresh Collector's assets and config files
    """
    flavor = AgentFlavor[flavor]
    # ensure BIN_PATH exists
    if not os.path.exists(BIN_PATH):
        os.mkdir(BIN_PATH)

    dist_folder = os.path.join(BIN_PATH, "dist")
    if os.path.exists(dist_folder):
        shutil.rmtree(dist_folder)
    os.mkdir(dist_folder)

    if "python" in build_tags:
        shutil.copytree("./cmd/agent/dist/checks/", os.path.join(dist_folder, "checks"), dirs_exist_ok=True)
        shutil.copytree("./cmd/agent/dist/utils/", os.path.join(dist_folder, "utils"), dirs_exist_ok=True)
        shutil.copy("./cmd/agent/dist/config.py", os.path.join(dist_folder, "config.py"))
    if not flavor.is_iot():
        shutil.copy("./cmd/agent/dist/dd-agent", os.path.join(dist_folder, "dd-agent"))
        # copy the dd-agent placeholder to the bin folder
        bin_ddagent = os.path.join(BIN_PATH, "sts-agent")  # sts
        shutil.move(os.path.join(dist_folder, "dd-agent"), bin_ddagent)

    # System probe not supported on windows
    if sys.platform.startswith('linux') or windows_sysprobe:
        shutil.copy("./cmd/agent/dist/system-probe.yaml", os.path.join(dist_folder, "system-probe.yaml"))
    shutil.copy("./cmd/agent/dist/stackstate.yaml", os.path.join(dist_folder, "stackstate.yaml"))

    for check in AGENT_CORECHECKS if not flavor.is_iot() else IOT_AGENT_CORECHECKS:
        check_dir = os.path.join(dist_folder, f"conf.d/{check}.d/")
        shutil.copytree(f"./cmd/agent/dist/conf.d/{check}.d/", check_dir, dirs_exist_ok=True)

    ## add additional windows-only corechecks, only on windows. Otherwise the check loader
    ## on linux will throw an error because the module is not found, but the config is.
    if sys.platform == 'win32':
        for check in WINDOWS_CORECHECKS:
            check_dir = os.path.join(dist_folder, f"conf.d/{check}.d/")
            shutil.copytree(f"./cmd/agent/dist/conf.d/{check}.d/", check_dir, dirs_exist_ok=True)

    if "apm" in build_tags:
        shutil.copy("./cmd/agent/dist/conf.d/apm.yaml.default", os.path.join(dist_folder, "conf.d/apm.yaml.default"))
    if "process" in build_tags:
        shutil.copy(
            "./cmd/agent/dist/conf.d/process_agent.yaml.default",
            os.path.join(dist_folder, "conf.d/process_agent.yaml.default"),
        )

    shutil.copytree("./cmd/agent/gui/views", os.path.join(dist_folder, "views"), dirs_exist_ok=True)
    if development:
        shutil.copytree("./dev/dist/", dist_folder, dirs_exist_ok=True)


@task
def run(
    ctx,
    rebuild=False,
    race=False,
    build_include=None,
    build_exclude=None,
    flavor=AgentFlavor.base.name,
    skip_build=False,
    config_path=None,
):
    """
    Execute the agent binary.

    By default it builds the agent before executing it, unless --skip-build was
    passed. It accepts the same set of options as agent.build.
    """
    if not skip_build:
        build(ctx, rebuild, race, build_include, build_exclude, flavor)

    agent_bin = os.path.join(BIN_PATH, bin_name("agent"))
    config_path = os.path.join(BIN_PATH, "dist", "datadog.yaml") if not config_path else config_path
    ctx.run(f"{agent_bin} run -c {config_path}")


@task
def exec(
    ctx,
    subcommand,
    config_path=None,
):
    """
    Execute 'agent <subcommand>' against the currently running Agent.

    This works against an agent run via `inv agent.run`.
    Basically this just simplifies creating the path for both the agent binary and config.
    """
    agent_bin = os.path.join(BIN_PATH, bin_name("agent"))
    config_path = os.path.join(BIN_PATH, "dist", "datadog.yaml") if not config_path else config_path
    ctx.run(f"{agent_bin} -c {config_path} {subcommand}")


@task
def system_tests(_):
    """
    Run the system testsuite.
    """
    pass


@task
def image_build(ctx, arch='amd64', base_dir="omnibus", python_version="2", skip_tests=False, tag=None, push=False):
    """
    Build the docker image
    """
    BOTH_VERSIONS = ["both", "2+3"]
    VALID_VERSIONS = ["2", "3"] + BOTH_VERSIONS
    if python_version not in VALID_VERSIONS:
        raise ParseError("provided python_version is invalid")

    build_context = "Dockerfiles/agent"
    base_dir = base_dir or os.environ.get("OMNIBUS_BASE_DIR")
    pkg_dir = os.path.join(base_dir, 'pkg')
    deb_glob = f'stackstate-agent*_{arch}.deb'
    dockerfile_path = f"{build_context}/Dockerfile"
    list_of_files = glob.glob(os.path.join(pkg_dir, deb_glob))
    # get the last debian package built
    if not list_of_files:
        print(f"No debian package build found in {pkg_dir}")
        print("See agent.omnibus-build")
        raise Exit(code=1)
    latest_file = max(list_of_files, key=os.path.getctime)
    shutil.copy2(latest_file, build_context)

    if tag is None:
        tag = AGENT_TAG

    common_build_opts = f"-t {tag} -f {dockerfile_path}"
    if python_version not in BOTH_VERSIONS:
        common_build_opts = f"{common_build_opts} --build-arg PYTHON_VERSION={python_version}"

    # Build with the testing target
    if not skip_tests:
        ctx.run(f"docker build {common_build_opts} --platform linux/{arch} --target testing {build_context}")

    # Build with the release target
    ctx.run(f"docker build {common_build_opts} --platform linux/{arch} --target release {build_context}")
    if push:
        ctx.run(f"docker push {tag}")

    ctx.run(f"rm {build_context}/{deb_glob}")


@task
def hacky_dev_image_build(
    ctx,
    base_image=None,
    target_image="agent",
    target_tag="latest",
    process_agent=False,
    trace_agent=False,
    push=False,
    signed_pull=False,
):
    if base_image is None:
        import requests
        import semver

        # Try to guess what is the latest release of the agent
        latest_release = semver.VersionInfo(0)
        tags = requests.get("https://gcr.io/v2/datadoghq/agent/tags/list")
        for tag in tags.json()['tags']:
            if not semver.VersionInfo.isvalid(tag):
                continue
            ver = semver.VersionInfo.parse(tag)
            if ver.prerelease or ver.build:
                continue
            if ver > latest_release:
                latest_release = ver
        base_image = f"gcr.io/datadoghq/agent:{latest_release}"

    # Extract the python library of the docker image
    with tempfile.TemporaryDirectory() as extracted_python_dir:
        ctx.run(
            f"docker run --rm '{base_image}' bash -c 'tar --create /opt/stackstate-agent/embedded/{{bin,lib,include}}/*python*' | tar --directory '{extracted_python_dir}' --extract"
        )

        os.environ["DELVE"] = "1"
        os.environ["LD_LIBRARY_PATH"] = (
            os.environ.get("LD_LIBRARY_PATH", "") + f":{extracted_python_dir}/opt/stackstate-agent/embedded/lib"
        )
        build(
            ctx,
            cmake_options=f'-DPython3_ROOT_DIR={extracted_python_dir}/opt/stackstate-agent/embedded -DPython3_FIND_STRATEGY=LOCATION',
        )
        ctx.run(
            f'perl -0777 -pe \'s|{extracted_python_dir}(/opt/stackstate-agent/embedded/lib/python\\d+\\.\\d+/../..)|substr $1."\\0"x length$&,0,length$&|e or die "pattern not found"\' -i dev/lib/libdatadog-agent-three.so'
        )
        if process_agent:
            process_agent_build(ctx)
        if trace_agent:
            trace_agent_build(ctx)

    copy_extra_agents = ""
    if process_agent:
        copy_extra_agents += "COPY bin/process-agent/process-agent /opt/stackstate-agent/embedded/bin/process-agent\n"
    if trace_agent:
        copy_extra_agents += "COPY bin/trace-agent/trace-agent /opt/stackstate-agent/embedded/bin/trace-agent\n"

    with tempfile.NamedTemporaryFile(mode='w') as dockerfile:
        dockerfile.write(
            f'''FROM ubuntu:latest AS src

COPY . /usr/src/stackstate-agent

RUN find /usr/src/stackstate-agent -type f \\! -name \\*.go -print0 | xargs -0 rm
RUN find /usr/src/stackstate-agent -type d -empty -print0 | xargs -0 rmdir

FROM ubuntu:latest AS bin

ENV DEBIAN_FRONTEND=noninteractive
RUN apt-get update && \
    apt-get install -y patchelf

COPY bin/agent/agent                            /opt/stackstate-agent/bin/agent/agent
COPY dev/lib/libdatadog-agent-rtloader.so.0.1.0 /opt/stackstate-agent/embedded/lib/libdatadog-agent-rtloader.so.0.1.0
COPY dev/lib/libdatadog-agent-three.so          /opt/stackstate-agent/embedded/lib/libdatadog-agent-three.so

RUN patchelf --set-rpath /opt/stackstate-agent/embedded/lib /opt/stackstate-agent/bin/agent/agent
RUN patchelf --set-rpath /opt/stackstate-agent/embedded/lib /opt/stackstate-agent/embedded/lib/libdatadog-agent-rtloader.so.0.1.0
RUN patchelf --set-rpath /opt/stackstate-agent/embedded/lib /opt/stackstate-agent/embedded/lib/libdatadog-agent-three.so

FROM golang:latest AS dlv

RUN go install github.com/go-delve/delve/cmd/dlv@latest

FROM {base_image}

ENV DEBIAN_FRONTEND=noninteractive
RUN apt-get update && \
    apt-get install -y bash-completion less vim tshark && \
    apt-get clean

ENV DELVE_PAGER=less

COPY --from=dlv /go/bin/dlv /usr/local/bin/dlv
COPY --from=src /usr/src/stackstate-agent {os.getcwd()}
COPY --from=bin /opt/stackstate-agent/bin/agent/agent                                 /opt/stackstate-agent/bin/agent/agent
COPY --from=bin /opt/stackstate-agent/embedded/lib/libdatadog-agent-rtloader.so.0.1.0 /opt/stackstate-agent/embedded/lib/libdatadog-agent-rtloader.so.0.1.0
COPY --from=bin /opt/stackstate-agent/embedded/lib/libdatadog-agent-three.so          /opt/stackstate-agent/embedded/lib/libdatadog-agent-three.so
{copy_extra_agents}
RUN agent completion bash > /usr/share/bash-completion/completions/agent

ENV DD_SSLKEYLOGFILE=/tmp/sslkeylog.txt
'''
        )
        dockerfile.flush()

        target_image_name = f'{target_image}:{target_tag}'
        pull_env = {}
        if signed_pull:
            pull_env['DOCKER_CONTENT_TRUST'] = '1'
        ctx.run(f'docker build -t {target_image_name} -f {dockerfile.name} .', env=pull_env)

        if push:
            ctx.run(f'docker push {target_image_name}')


@task
def integration_tests(ctx, install_deps=False, race=False, remote_docker=False, go_mod="mod", arch="x64"):
    """
    Run integration tests for the Agent
    """
    if install_deps:
        deps(ctx)

    if sys.platform == 'win32':
        return _windows_integration_tests(ctx, race=race, go_mod=go_mod, arch=arch)
    else:
        # TODO: See if these will function on Windows
        return _linux_integration_tests(ctx, race=race, remote_docker=remote_docker, go_mod=go_mod, arch=arch)


def _windows_integration_tests(ctx, race=False, go_mod="mod", arch="x64"):
    test_args = {
        "go_mod": go_mod,
        "go_build_tags": " ".join(get_default_build_tags(build="test", arch=arch)),
        "race_opt": "-race" if race else "",
        "exec_opts": "",
    }

    go_cmd = 'go test -mod={go_mod} {race_opt} -tags "{go_build_tags}" {exec_opts}'.format(**test_args)  # noqa: FS002

    tests = [
        {
            # Run eventlog tests with the Windows API, which depend on the EventLog service
            "dir": "./pkg/util/winutil/",
            'prefix': './eventlog/...',
            'extra_args': '-evtapi Windows',
        },
        {
            # Run eventlog tailer tests with the Windows API, which depend on the EventLog service
            "dir": ".",
            'prefix': './pkg/logs/tailers/windowsevent/...',
            'extra_args': '-evtapi Windows',
        },
        {
            # Run eventlog check tests with the Windows API, which depend on the EventLog service
            "dir": ".",
            'prefix': './pkg/collector/corechecks/windows_event_log/...',
            'extra_args': '-evtapi Windows',
        },
    ]

    for test in tests:
        with ctx.cd(f"{test['dir']}"):
            ctx.run(f"{go_cmd} {test['prefix']} {test['extra_args']}")


def _linux_integration_tests(ctx, race=False, remote_docker=False, go_mod="mod", arch="x64"):
    test_args = {
        "go_mod": go_mod,
        "go_build_tags": " ".join(get_default_build_tags(build="test", arch=arch)),
        "race_opt": "-race" if race else "",
        "exec_opts": "",
    }

    # since Go 1.13, the -exec flag of go test could add some parameters such as -test.timeout
    # to the call, we don't want them because while calling invoke below, invoke
    # thinks that the parameters are for it to interpret.
    # we're calling an intermediate script which only pass the binary name to the invoke task.
    if remote_docker:
        test_args["exec_opts"] = f"-exec \"{os.getcwd()}/test/integration/dockerize_tests.sh\""

    go_cmd = 'go test -mod={go_mod} {race_opt} -tags "{go_build_tags}" {exec_opts}'.format(**test_args)  # noqa: FS002

    prefixes = [
        "./test/integration/config_providers/...",
        "./test/integration/corechecks/...",
        "./test/integration/listeners/...",
        "./test/integration/util/kubelet/...",
    ]

    for prefix in prefixes:
        ctx.run(f"{go_cmd} {prefix}")


def get_omnibus_env(
    ctx,
    skip_sign=False,
    release_version="nightly",
    major_version='3',
    python_runtimes='3',
    hardened_runtime=False,
    system_probe_bin=None,
    go_mod_cache=None,
    flavor=AgentFlavor.base,
    pip_config_file="pip.conf",
):
    env = load_release_versions(ctx, release_version)

    # If the host has a GOMODCACHE set, try to reuse it
    if not go_mod_cache and os.environ.get('GOMODCACHE'):
        go_mod_cache = os.environ.get('GOMODCACHE')

    if go_mod_cache:
        env['OMNIBUS_GOMODCACHE'] = go_mod_cache

    if int(major_version) > 6:
        env['OMNIBUS_OPENSSL_SOFTWARE'] = 'openssl3'

    env_override = ['INTEGRATIONS_CORE_VERSION', 'OMNIBUS_SOFTWARE_VERSION']
    for key in env_override:
        value = os.environ.get(key)
        # Only overrides the env var if the value is a non-empty string.
        if value:
            env[key] = value

    if sys.platform == 'win32' and os.environ.get('SIGN_WINDOWS'):
        # get certificate and password from ssm
        pfxfile = get_signing_cert(ctx)
        pfxpass = get_pfx_pass(ctx)
        env['SIGN_PFX'] = str(pfxfile)
        env['SIGN_PFX_PW'] = str(pfxpass)

    if sys.platform == 'darwin':
        # Target MacOS 10.12
        env['MACOSX_DEPLOYMENT_TARGET'] = '10.12'

    if skip_sign:
        env['SKIP_SIGN_MAC'] = 'true'
    if hardened_runtime:
        env['HARDENED_RUNTIME_MAC'] = 'true'

    env['PACKAGE_VERSION'] = get_version(
        ctx, include_git=True, url_safe=True, major_version=major_version, include_pipeline_id=True
    )
    env['MAJOR_VERSION'] = major_version
    env['PY_RUNTIMES'] = python_runtimes

    # Since omnibus and the invoke task won't run in the same folder
    # we need to input the absolute path of the pip config file
    env['PIP_CONFIG_FILE'] = os.path.abspath(pip_config_file)

    if system_probe_bin:
        env['SYSTEM_PROBE_BIN'] = system_probe_bin
    env['AGENT_FLAVOR'] = flavor.name

    # We need to override the workers variable in omnibus build when running on Kubernetes runners,
    # otherwise, ohai detect the number of CPU on the host and run the make jobs with all the CPU.
    if os.environ.get('KUBERNETES_CPU_REQUEST'):
        env['OMNIBUS_WORKERS_OVERRIDE'] = str(int(os.environ.get('KUBERNETES_CPU_REQUEST')) + 1)
    # Forward the DEPLOY_AGENT variable so that we can use a higher compression level for deployed artifacts
    if os.environ.get('DEPLOY_AGENT'):
        env['DEPLOY_AGENT'] = os.environ.get('DEPLOY_AGENT')

    return env


def omnibus_run_task(
    ctx, task, target_project, base_dir, env, omnibus_s3_cache=False, log_level="info", host_distribution=None
):
    with ctx.cd("omnibus"):
        overrides_cmd = ""
        if base_dir:
            overrides_cmd = f"--override=base_dir:{base_dir}"
        if host_distribution:
            overrides_cmd += f" --override=host_distribution:{host_distribution}"

        omnibus = "bundle exec omnibus"
        if sys.platform == 'win32':
            omnibus = "bundle exec omnibus.bat"
        elif sys.platform == 'darwin':
            # HACK: This is an ugly hack to fix another hack made by python3 on MacOS
            # The full explanation is available on this PR: https://github.com/StackVista/stackstate-agent/pull/5010.
            omnibus = "unset __PYVENV_LAUNCHER__ && bundle exec omnibus"

        if omnibus_s3_cache:
            populate_s3_cache = "--populate-s3-cache"
        else:
            populate_s3_cache = ""

        cmd = "{omnibus} {task} {project_name} --log-level={log_level} {populate_s3_cache} {overrides}"
        args = {
            "omnibus": omnibus,
            "task": task,
            "project_name": target_project,
            "log_level": log_level,
            "overrides": overrides_cmd,
            "populate_s3_cache": populate_s3_cache,
        }

        ctx.run(cmd.format(**args), env=env)


def bundle_install_omnibus(ctx, gem_path=None, env=None, max_try=2):
    with ctx.cd("omnibus"):
        # make sure bundle install starts from a clean state
        try:
            os.remove("Gemfile.lock")
        except Exception:
            pass

        cmd = "bundle install"
        if gem_path:
            cmd += f" --path {gem_path}"

        for trial in range(max_try):
            res = ctx.run(cmd, env=env, warn=True)
            if res.ok:
                return
            if not should_retry_bundle_install(res):
                return
            print(f"Retrying bundle install, attempt {trial + 1}/{max_try}")


def should_retry_bundle_install(res):
    # We sometimes get a Net::HTTPNotFound error when fetching the
    # license-scout gem. This is a transient error, so we retry the bundle install
    if "Net::HTTPNotFound:" in res.stderr:
        return True
    return False


# hardened-runtime needs to be set to False to build on MacOS < 10.13.6, as the -o runtime option is not supported.
@task(
    help={
        'skip-sign': "On macOS, use this option to build an unsigned package if you don't have Datadog's developer keys.",
        'hardened-runtime': "On macOS, use this option to enforce the hardened runtime setting, adding '-o runtime' to all codesign commands",
    }
)
def omnibus_build(
    ctx,
    flavor=AgentFlavor.base.name,
    agent_binaries=False,
    log_level="info",
    base_dir=None,
    gem_path=None,
    skip_deps=False,
    skip_sign=False,
    release_version="nightly",
    major_version='7',
    python_runtimes='3',
    omnibus_s3_cache=False,
    hardened_runtime=False,
    system_probe_bin=None,
    go_mod_cache=None,
    python_mirror=None,
    pip_config_file="pip.conf",
    host_distribution=None,
):
    """
    Build the Agent packages with Omnibus Installer.
    """

    flavor = AgentFlavor[flavor]
    if not skip_deps:
        with timed(quiet=True) as deps_elapsed:
            deps(ctx)

    # sts
    apply_branding(ctx)

    # base dir (can be overridden through env vars, command line takes precedence)
    base_dir = base_dir or os.environ.get("OMNIBUS_BASE_DIR")

    if base_dir is not None and sys.platform == 'win32':
        # On Windows, prevent backslashes in the base_dir path otherwise omnibus will fail with
        # error 'no matched files for glob copy' at the end of the build.
        base_dir = base_dir.replace(os.path.sep, '/')

    env = get_omnibus_env(
        ctx,
        skip_sign=skip_sign,
        release_version=release_version,
        major_version=major_version,
        python_runtimes=python_runtimes,
        hardened_runtime=hardened_runtime,
        system_probe_bin=system_probe_bin,
        go_mod_cache=go_mod_cache,
        flavor=flavor,
        pip_config_file=pip_config_file,
    )

    target_project = "agent"
    if flavor.is_iot():
        target_project = "iot-agent"
    elif agent_binaries:
        target_project = "agent-binaries"

    # Get the python_mirror from the PIP_INDEX_URL environment variable if it is not passed in the args
    python_mirror = python_mirror or os.environ.get("PIP_INDEX_URL")

    # If a python_mirror is set then use it for pip by adding it in the pip.conf file
    pip_index_url = f"[global]\nindex-url = {python_mirror}" if python_mirror else ""

    # We're passing the --index-url arg through a pip.conf file so that omnibus doesn't leak the token
    with open(pip_config_file, 'w') as f:
        f.write(pip_index_url)

    with timed(quiet=True) as bundle_elapsed:
        bundle_install_omnibus(ctx, gem_path, env)

    with timed(quiet=True) as omnibus_elapsed:
        omnibus_run_task(
            ctx=ctx,
            task="build",
            target_project=target_project,
            base_dir=base_dir,
            env=env,
            omnibus_s3_cache=omnibus_s3_cache,
            log_level=log_level,
            host_distribution=host_distribution,
        )

    # Delete the temporary pip.conf file once the build is done
    os.remove(pip_config_file)

    print("Build component timing:")
    if not skip_deps:
        print(f"Deps:    {deps_elapsed.duration}")
    print(f"Bundle:  {bundle_elapsed.duration}")
    print(f"Omnibus: {omnibus_elapsed.duration}")


@task
def build_dep_tree(ctx, git_ref=""):
    """
    Generates a file representing the Golang dependency tree in the current
    directory. Use the "--git-ref=X" argument to specify which tag you would like
    to target otherwise current repo state will be used.
    """
    saved_branch = None
    if git_ref:
        print(f"Tag {git_ref} specified. Checking out the branch...")

        result = ctx.run("git rev-parse --abbrev-ref HEAD", hide='stdout')
        saved_branch = result.stdout

        ctx.run(f"git checkout {git_ref}")
    else:
        print("No tag specified. Using the current state of repository.")

    try:
        ctx.run("go run tools/dep_tree_resolver/go_deps.go")
    finally:
        if saved_branch:
            ctx.run(f"git checkout {saved_branch}", hide='stdout')


@task
def omnibus_manifest(
    ctx,
    platform=None,
    arch=None,
    flavor=AgentFlavor.base.name,
    agent_binaries=False,
    log_level="info",
    base_dir=None,
    gem_path=None,
    skip_sign=False,
    release_version="nightly",
    major_version='7',
    python_runtimes='3',
    hardened_runtime=False,
    system_probe_bin=None,
    go_mod_cache=None,
):
    flavor = AgentFlavor[flavor]
    # base dir (can be overridden through env vars, command line takes precedence)
    base_dir = base_dir or os.environ.get("OMNIBUS_BASE_DIR")

    env = get_omnibus_env(
        ctx,
        skip_sign=skip_sign,
        release_version=release_version,
        major_version=major_version,
        python_runtimes=python_runtimes,
        hardened_runtime=hardened_runtime,
        system_probe_bin=system_probe_bin,
        go_mod_cache=go_mod_cache,
        flavor=flavor,
    )

    target_project = "agent"
    if flavor.is_iot():
        target_project = "iot-agent"
    elif agent_binaries:
        target_project = "agent-binaries"

    bundle_install_omnibus(ctx, gem_path, env)

    task = "manifest"
    if platform is not None:
        task += f" --platform-family={platform} --platform={platform} "
    if arch is not None:
        task += f" --architecture={arch} "

    omnibus_run_task(
        ctx=ctx,
        task=task,
        target_project=target_project,
        base_dir=base_dir,
        env=env,
        omnibus_s3_cache=False,
        log_level=log_level,
    )


@task
def check_supports_python_version(_, check_dir, python):
    """
    Check if a Python project states support for a given major Python version.
    """
    import toml
    from packaging.specifiers import SpecifierSet

    if python not in ['2', '3']:
        raise Exit("invalid Python version", code=2)

    project_file = os.path.join(check_dir, 'pyproject.toml')
    setup_file = os.path.join(check_dir, 'setup.py')
    if os.path.isfile(project_file):
        with open(project_file, 'r') as f:
            data = toml.loads(f.read())

        project_metadata = data['project']
        if 'requires-python' not in project_metadata:
            print('True', end='')
            return

        specifier = SpecifierSet(project_metadata['requires-python'])
        # It might be e.g. `>=3.8` which would not immediatelly contain `3`
        for minor_version in range(100):
            if specifier.contains(f'{python}.{minor_version}'):
                print('True', end='')
                return
        else:
            print('False', end='')
    elif os.path.isfile(setup_file):
        with open(setup_file, 'r') as f:
            tree = ast.parse(f.read(), filename=setup_file)

        prefix = f'Programming Language :: Python :: {python}'
        for node in ast.walk(tree):
            if isinstance(node, ast.keyword) and node.arg == 'classifiers':
                classifiers = ast.literal_eval(node.value)
                print(any(cls.startswith(prefix) for cls in classifiers), end='')
                return
        else:
            print('False', end='')
    else:
        raise Exit('not a Python project', code=1)


@task
def clean(ctx):
    """
    Remove temporary objects and binary artifacts
    """
    # go clean
    print("Executing go clean")
    ctx.run("go clean")

    # remove the bin/agent folder
    print("Remove agent binary folder")
    ctx.run("rm -rf ./bin/agent")

    print("Cleaning rtloader")
    rtloader_clean(ctx)


@task
def version(
    ctx,
    url_safe=False,
    omnibus_format=False,
    git_sha_length=7,
    major_version='7',
    version_cached=False,
    pipeline_id=None,
    include_git=True,
    include_pre=True,
):
    """
    Get the agent version.
    url_safe: get the version that is able to be addressed as a url
    omnibus_format: performs the same transformations omnibus does on version names to
                    get the exact same string that's used in package names
    git_sha_length: different versions of git have a different short sha length,
                    use this to explicitly set the version
                    (the windows builder and the default ubuntu version have such an incompatibility)
    version_cached: save the version inside a "agent-version.cache" that will be reused
                    by each next call of version.
    """
    if version_cached:
        cache_version(ctx, git_sha_length=git_sha_length)

    version = get_version(
        ctx,
        include_git=include_git,
        url_safe=url_safe,
        git_sha_length=git_sha_length,
        major_version=major_version,
        include_pipeline_id=True,
        pipeline_id=pipeline_id,
        include_pre=include_pre,
    )
    if omnibus_format:
        # See: https://github.com/StackVista/omnibus-ruby/blob/datadog-5.5.0/lib/omnibus/packagers/deb.rb#L599
        # In theory we'd need to have one format for each package type (deb, rpm, msi, pkg).
        # However, there are a few things that allow us in practice to have only one variable for everything:
        # - the deb and rpm safe version formats are identical (the only difference is an additional rule on Wind River Linux, which doesn't apply to us).
        #   Moreover, of the two rules, we actually really only use the first one (because we always use inv agent.version --url-safe).
        # - the msi version name uses the raw version string. The only difference with the deb / rpm versions
        #   is therefore that dashes are replaced by tildes. We're already doing the reverse operation in agent-release-management
        #   to get the correct msi name.
        # - the pkg version name uses the raw version + a variation of the second rule (where a dash is used in place of an underscore).
        #   Once again, replacing tildes by dashes (+ replacing underscore by dashes if we ever end up using the second rule for some reason)
        #   in agent-release-management is enough. We're already replacing tildes by dashes in agent-release-management.
        # TODO: investigate if having one format per package type in the agent.version method makes more sense.
        version = re.sub('-', '~', version)
        version = re.sub(r'[^a-zA-Z0-9\.\+\:\~]+', '_', version)
    print(version)


@task
def get_integrations_from_cache(ctx, python, bucket, branch, integrations_dir, target_dir, integrations, awscli="aws"):
    """
    Get cached integration wheels for given integrations.
    python: Python version to retrieve integrations for
    bucket: S3 bucket to retrieve integration wheels from
    branch: namespace in the bucket to get the integration wheels from
    integrations_dir: directory with Git repository of integrations
    target_dir: local directory to put integration wheels to
    integrations: comma-separated names of the integrations to try to retrieve from cache
    awscli: AWS CLI executable to call
    """
    integrations_hashes = {}
    for integration in integrations.strip().split(","):
        integration_path = os.path.join(integrations_dir, integration)
        if not os.path.exists(integration_path):
            raise Exit(f"Integration {integration} given, but doesn't exist in {integrations_dir}", code=2)
        last_commit = ctx.run(
            LAST_DIRECTORY_COMMIT_PATTERN.format(integrations_dir=integrations_dir, integration=integration),
            hide="both",
            echo=False,
        )
        integrations_hashes[integration] = last_commit.stdout.strip()

    print(f"Trying to retrieve {len(integrations_hashes)} integration wheels from cache")
    # On windows, maximum length of a command line call is 8191 characters, therefore
    # we do multiple syncs that fit within that limit (we use 8100 as a nice round number
    # and just to make sure we don't do any of-by-one errors that would break this).
    # WINDOWS NOTES: on Windows, the awscli is usually in program files, so we have to wrap the
    # executable in quotes; also we have to not put the * in quotes, as there's no
    # expansion on it, unlike on Linux
    exclude_wildcard = "*" if platform.system().lower() == "windows" else "'*'"
    sync_command_prefix = (
        f"\"{awscli}\" s3 sync s3://{bucket} {target_dir} --no-sign-request --exclude {exclude_wildcard}"
    )
    sync_commands = [[[sync_command_prefix], len(sync_command_prefix)]]
    for integration, hash in integrations_hashes.items():
        include_arg = " --include " + CACHED_WHEEL_FULL_PATH_PATTERN.format(
            hash=hash,
            integration=integration,
            python_version=python,
            branch=branch,
        )
        if len(include_arg) + sync_commands[-1][1] > 8100:
            sync_commands.append([[sync_command_prefix], len(sync_command_prefix)])
        sync_commands[-1][0].append(include_arg)
        sync_commands[-1][1] += len(include_arg)

    for sync_command in sync_commands:
        ctx.run("".join(sync_command[0]))

    found = []
    # move all wheel files directly to the target_dir, so they're easy to find/work with in Omnibus
    for integration in sorted(integrations_hashes):
        hash = integrations_hashes[integration]
        original_path_glob = os.path.join(
            target_dir,
            CACHED_WHEEL_FULL_PATH_PATTERN.format(
                hash=hash,
                integration=integration,
                python_version=python,
                branch=branch,
            ),
        )
        files_matched = glob.glob(original_path_glob)
        if len(files_matched) == 0:
            continue
        elif len(files_matched) > 1:
            raise Exit(
                f"More than 1 wheel for integration {integration} matched by {original_path_glob}: {files_matched}"
            )
        wheel_path = files_matched[0]
        print(f"Found cached wheel for integration {integration}")
        shutil.move(wheel_path, target_dir)
        found.append(f"datadog_{integration}")

    print(f"Found {len(found)} cached integration wheels")
    with open(os.path.join(target_dir, "found.txt"), "w") as f:
        f.write('\n'.join(found))


@task
def upload_integration_to_cache(ctx, python, bucket, branch, integrations_dir, build_dir, integration, awscli="aws"):
    """
    Upload a built integration wheel for given integration.
    python: Python version the integration is built for
    bucket: S3 bucket to upload the integration wheel to
    branch: namespace in the bucket to upload the integration wheels to
    integrations_dir: directory with Git repository of integrations
    build_dir: directory containing the built integration wheel
    integration: name of the integration being cached
    awscli: AWS CLI executable to call
    """
    matching_glob = os.path.join(build_dir, CACHED_WHEEL_FILENAME_PATTERN.format(integration=integration))
    files_matched = glob.glob(matching_glob)
    if len(files_matched) == 0:
        raise Exit(f"No wheel for integration {integration} found in {build_dir}")
    elif len(files_matched) > 1:
        raise Exit(f"More than 1 wheel for integration {integration} matched by {matching_glob}: {files_matched}")

    wheel_path = files_matched[0]

    last_commit = ctx.run(
        LAST_DIRECTORY_COMMIT_PATTERN.format(integrations_dir=integrations_dir, integration=integration),
        hide="both",
        echo=False,
    )
    hash = last_commit.stdout.strip()

    target_name = CACHED_WHEEL_DIRECTORY_PATTERN.format(
        hash=hash, python_version=python, branch=branch
    ) + os.path.basename(wheel_path)
    print(f"Caching wheel {target_name}")
    # NOTE: on Windows, the awscli is usually in program files, so we have the executable
    ctx.run(f"\"{awscli}\" s3 cp {wheel_path} s3://{bucket}/{target_name} --acl public-read")
