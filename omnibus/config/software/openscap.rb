# Unless explicitly stated otherwise all files in this repository are licensed
# under the Apache License Version 2.0.
# This product includes software developed at Datadog (https:#www.datadoghq.com/).
# Copyright 2016-present Datadog, Inc.

name 'openscap'
default_version '1.4.3'

# [sts] STS delta vs upstream base-7.78.2:
#   - `--downloader_config=/dev/null` on every bazelisk call overrides
#     .adms/bazel/adms.mirror.cfg, which rewrites GitHub/canonical URLs to
#     DD's internal depot-read-api-bzl mirror (unreachable from STS runners).
#   - libxml2 + libxslt are provided by //packages/agent/dependencies:install
#     (@libxml2//:all_files + @libxslt//:all_files) rather than installed here.

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

  command_on_repo_root "bazelisk run --downloader_config=/dev/null -- @libyaml//:install --destdir='#{install_dir}'"
  sh_lib = if linux_target? then "libyaml.so" else "libyaml.dylib" end
  command_on_repo_root "bazelisk run --downloader_config=/dev/null -- //bazel/rules:replace_prefix --prefix '#{install_dir}/embedded' " \
    "#{install_dir}/embedded/lib/#{sh_lib}"

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
