# Unless explicitly stated otherwise all files in this repository are licensed
# under the Apache License Version 2.0.
# This product includes software developed at Datadog (https:#www.datadoghq.com/).
# Copyright 2016-present Datadog, Inc.

name 'openscap'
default_version '1.4.3'

# [sts] STAC-24773 Phase B1: openscap.rb migrated to Bazel install. Mirrors DD
# pristine 7.78.2's structure (zero `dependency` lines; all chain deps pulled
# inline via bazelisk targets in the build block below). Implicit version bump
# 1.4.2 -> 1.4.3 — pkg/compliance/evaluator_xccdf.go exercises openscap at
# runtime; sandbox compliance smoke gate before B2.
#
# `--downloader_config=/dev/null` overrides DD's .adms/bazel/adms.mirror.cfg
# which rewrites GitHub/canonical URLs to DD's internal depot-read-api-bzl
# mirror (unreachable from STS runners). Same workaround as the bazelisk line
# in datadog-agent-dependencies.rb.

# [sts] STAC-24773 B1 fix-up: libyaml is omnibus-managed because python3.rb
# consumes it; python3.rb stays per Phase D deferral. The `bazelisk run
# @libyaml//:install` line from pristine 7.78.2 was REMOVED below because it
# would collide with the omnibus-installed libyaml.so symlink (FileExistsError
# observed on libyaml.so.0 in pipeline 2613584162). The Bazel-built openscap
# binary dynamically links libyaml.so at runtime — omnibus 0.2.2 and Bazel
# 0.2.5 are ABI-compatible.
#
# libxml2 + libxslt were also omnibus-managed in B1 fix-up 2 for the same
# reason; C2 migrated them to Bazel (provided by //packages/agent/
# dependencies:install via @libxml2//:all_files + @libxslt//:all_files) and
# removed the corresponding `dependency` lines from this file.
dependency 'libyaml'

build do
  command_on_repo_root "bazelisk run --downloader_config=/dev/null -- @acl//:install --destdir='#{install_dir}'"
  command_on_repo_root "bazelisk run --downloader_config=/dev/null -- //bazel/rules:replace_prefix --prefix '#{install_dir}/embedded'" \
    " #{install_dir}/embedded/lib/libacl.so"

  command_on_repo_root "bazelisk run --downloader_config=/dev/null -- @attr//:install --destdir='#{install_dir}'"
  command_on_repo_root "bazelisk run --downloader_config=/dev/null -- //bazel/rules:replace_prefix --prefix '#{install_dir}/embedded'" \
    " #{install_dir}/embedded/lib/libattr.so"

  command_on_repo_root "bazelisk run --downloader_config=/dev/null -- @dbus//:install --destdir='#{install_dir}'"

  command_on_repo_root "bazelisk run --downloader_config=/dev/null -- @libselinux//:install --destdir='#{install_dir}'"
  command_on_repo_root "bazelisk run --downloader_config=/dev/null -- //bazel/rules:replace_prefix --prefix '#{install_dir}/embedded'" \
    " #{install_dir}/embedded/lib/libselinux.so"

  command_on_repo_root "bazelisk run --downloader_config=/dev/null -- @libsepol//:install --destdir='#{install_dir}'"
  command_on_repo_root "bazelisk run --downloader_config=/dev/null -- //bazel/rules:replace_prefix --prefix '#{install_dir}/embedded'" \
    " #{install_dir}/embedded/lib/libsepol.so"

  # [sts] @libyaml//:install + replace_prefix REMOVED — see note at top of file.

  command_on_repo_root "bazelisk run --downloader_config=/dev/null -- @pcre2//:install --destdir=#{install_dir}"
  command_on_repo_root "bazelisk run --downloader_config=/dev/null -- //bazel/rules:replace_prefix " \
    "--prefix #{install_dir}/embedded " \
    "#{install_dir}/embedded/lib/libpcre2*.so"

  command_on_repo_root "bazelisk run --downloader_config=/dev/null -- @popt//:install --destdir='#{install_dir}'"
  command_on_repo_root "bazelisk run --downloader_config=/dev/null -- //bazel/rules:replace_prefix --prefix '#{install_dir}/embedded'" \
    " #{install_dir}/embedded/lib/libpopt.so"

  command_on_repo_root "bazelisk run --downloader_config=/dev/null -- @rpm//:install --destdir='#{install_dir}'"
  command_on_repo_root "bazelisk run --downloader_config=/dev/null -- //bazel/rules:replace_prefix --prefix '#{install_dir}/embedded'" \
    " #{install_dir}/embedded/lib/librpm.so" \
    " #{install_dir}/embedded/lib/librpmio.so"

  command_on_repo_root "bazelisk run --downloader_config=/dev/null -- @util-linux//:blkid_install --destdir='#{install_dir}'"
  command_on_repo_root "bazelisk run --downloader_config=/dev/null -- //bazel/rules:replace_prefix --prefix '#{install_dir}/embedded'" \
    " #{install_dir}/embedded/lib/libblkid.so"

  command_on_repo_root "bazelisk run --downloader_config=/dev/null -- @gpg-error//:install --destdir='#{install_dir}'"
  command_on_repo_root "bazelisk run --downloader_config=/dev/null -- @gcrypt//:install --destdir='#{install_dir}'"
  command_on_repo_root "bazelisk run --downloader_config=/dev/null -- //bazel/rules:replace_prefix --prefix '#{install_dir}/embedded'" \
    " #{install_dir}/embedded/lib/libgcrypt.so" \
    " #{install_dir}/embedded/lib/libgpg-error.so" \

  # [sts] @libxml2//:install + replace_prefix REMOVED — see note at top of file.

  # [sts] @libxslt//:install + replace_prefix REMOVED — see note at top of file.

  command_on_repo_root "bazelisk run --downloader_config=/dev/null -- @xmlsec//:install --destdir='#{install_dir}'"
  command_on_repo_root "bazelisk run --downloader_config=/dev/null -- //bazel/rules:replace_prefix --prefix '#{install_dir}/embedded'" \
    " #{install_dir}/embedded/lib/libxmlsec1*.so" \

  command_on_repo_root "bazelisk run --downloader_config=/dev/null -- @openscap//:install --destdir='#{install_dir}'"
  command_on_repo_root "bazelisk run --downloader_config=/dev/null -- //bazel/rules:replace_prefix --prefix '#{install_dir}/embedded'" \
    " #{install_dir}/embedded/lib/libopenscap.so" \
    " #{install_dir}/embedded/lib/libopenscap_sce.so" \
    " #{install_dir}/embedded/bin/oscap" \
    " #{install_dir}/embedded/bin/oscap-io"

end
