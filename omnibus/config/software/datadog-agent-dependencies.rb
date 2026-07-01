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
# [sts] STAC-24773 B2: curl + nghttp2 migrated to Bazel install. Mirrors DD
# pristine 7.78.2's Linux build do block; --downloader_config=/dev/null added
# per the Bazel-egress workaround (see comment on the //packages/agent/
# dependencies:install line below). nghttp2 must install BEFORE curl because
# the resulting libcurl.so links against libnghttp2.so.
if linux_target?
  build do
    command_on_repo_root "bazelisk run #{flavor_flag} --downloader_config=/dev/null -- @nghttp2//:install --destdir='#{install_dir}'"
    command_on_repo_root "bazelisk run #{flavor_flag} --downloader_config=/dev/null -- //bazel/rules:replace_prefix --prefix '#{install_dir}/embedded'" \
      " #{install_dir}/embedded/lib/libnghttp2.so"

    command_on_repo_root "bazelisk run #{flavor_flag} --downloader_config=/dev/null -- @curl//:install --destdir='#{install_dir}'"
    command_on_repo_root "bazelisk run #{flavor_flag} --downloader_config=/dev/null -- //bazel/rules:replace_prefix --prefix '#{install_dir}/embedded'" \
      " #{install_dir}/embedded/lib/libcurl.so" \
      " #{install_dir}/embedded/bin/curl"

    # [sts] STAC-24773 C2: libxml2 + libxslt are Bazel-installed (replacing
    # omnibus libxml2.rb / libxslt.rb) early enough for stackstate-agent-
    # integrations-py{2,3} arm pip install of lxml AND openscap's later
    # link against them. Mirrors the install + replace_prefix pattern that
    # nghttp2 / curl above use — `pkg_filegroup all_files` route via
    # //packages/agent/dependencies:install was tried first (pipeline
    # 2619669055) and FAILED omnibus health check because pkg_install
    # ships the .so as-is without RPATH normalization, leaving DT_NEEDED
    # libz/liblzma/libicuuc resolving to system /lib instead of
    # /opt/.../embedded/lib. The install + replace_prefix two-step is the
    # correct pattern for any Bazel-built shared library STS needs in
    # /opt/.../embedded/.
    command_on_repo_root "bazelisk run #{flavor_flag} --downloader_config=/dev/null -- @libxml2//:install --destdir='#{install_dir}'"
    command_on_repo_root "bazelisk run #{flavor_flag} --downloader_config=/dev/null -- //bazel/rules:replace_prefix --prefix '#{install_dir}/embedded'" \
      " #{install_dir}/embedded/lib/libxml2.so"

    command_on_repo_root "bazelisk run #{flavor_flag} --downloader_config=/dev/null -- @libxslt//:install --destdir='#{install_dir}'"
    command_on_repo_root "bazelisk run #{flavor_flag} --downloader_config=/dev/null -- //bazel/rules:replace_prefix --prefix '#{install_dir}/embedded'" \
      " #{install_dir}/embedded/lib/libxslt.so" \
      " #{install_dir}/embedded/lib/libexslt.so"
  end
end
if fips_mode?
  dependency 'openssl-fips-provider'
end

# [sts] STAC-24773 C5: secret-generic-connector via Bazel prebuilt (replaces
# omnibus secret-generic-connector.rb). Mirrors upstream ABLD-251 cutover;
# mode 0500 is set in deps/secret_connector/BUILD.bazel (agent.rb chmod is
# redundant but kept for belt-and-suspenders).
unless heroku_target?
  build do
    command_on_repo_root "bazelisk run #{flavor_flag} --downloader_config=/dev/null -- //deps/secret_connector:install --destdir=#{install_dir}"
  end
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
