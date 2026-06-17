name 'datadog-agent-dependencies'

description "Enforce building dependencies as soon as possible so they can be cached"

# [sts] STAC-24773: Bazel migration. The remaining `dependency '...'` lines below
# (jmxfetch, libpcap, systemd, snmp-traps, procps-ng) are progressively replaced by
# the //packages/agent/dependencies:install Bazel target invoked in the build do
# block at the bottom. Mirrors DD upstream's pattern in pristine 7.78.2 with
# flavor_flag computed for fips/heroku awareness.
if heroku_target?
  flavor_flag = "--//packages/agent:flavor=heroku"
else
  flavor_flag = fips_mode? ? "--//packages/agent:flavor=fips" : ""
end

# Linux-specific dependencies
if linux_target?
  dependency 'curl'
end
if fips_mode?
  dependency 'openssl-fips-provider'
end

# Bundled cacerts file (is this a good idea?)
dependency 'cacerts'

# Used for memory profiling with the `status py` agent subcommand
dependency 'pympler'

# [STS] StackState integrations are declared in agent.rb (project level)
# to avoid circular dependency: datadog-agent -> datadog-agent-dependencies -> integrations -> datadog-agent

# [sts] STAC-24773: STS does not ship Windows. The `if windows_target?` block that
# previously pulled in datadog-windows-{filter-driver,apminject,procmon-driver}
# has been removed along with the corresponding .rb recipe files. STS does NOT
# support Windows at all; this is dead code in our tree. Any Windows-conditional
# stanza in inherited DD files should be similarly stripped during the broader
# audit of restored .rb files (see follow-up STAC ticket).

build do
    # [sts] STAC-24773: Bazel install for jmxfetch / libpcap / systemd / snmp-traps
    # (and other Bazel-shipped deps). Runs alongside the legacy `dependency '...'`
    # chain in this transitional commit; subsequent commits drop the duplicated
    # legacy dependencies one-by-one.
    #
    # `--downloader_config=/dev/null` overrides DD's .adms/bazel/adms.mirror.cfg
    # which rewrites all GitHub / canonical URLs to go through DD's internal
    # mirror at depot-read-api-bzl.us1.ddbuild.io. That mirror is unreachable
    # from STS runners (private DD infrastructure). The /dev/null override tells
    # Bazel to use direct URLs, restoring fetches from github.com / bcr.bazel.build
    # / etc. Without this, every bazelisk invocation in STS CI times out on the
    # first module-resolution fetch (observed in the first build_deb attempt of
    # this commit before the override was added — depot-read-api-bzl times out
    # fetching github.com/aiuto/supply-chain/archive/refs/tags/dd_test.tar.gz).
    command_on_repo_root "bazelisk run #{flavor_flag} --downloader_config=/dev/null -- //packages/agent/dependencies:install --destdir=#{install_dir}"
end

build do
    # Delete empty folders that can still be present when building
    # without the omnibus cache.
    # When the cache gets used, git will transparently remove empty dirs for us
    # We do this here since we are done building our dependencies, but haven't
    # started creating the agent directories, which might be empty but that we
    # still want to keep
    command "find #{install_dir} -type d -empty -delete"
end
