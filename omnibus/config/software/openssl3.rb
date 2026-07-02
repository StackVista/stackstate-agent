#
# Copyright 2023 Chef Software, Inc.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.
#

name "openssl3"

license "Apache-2.0"
license_file "LICENSE.txt"
skip_transitive_dependency_licensing true

dependency "cacerts"

# [sts] STAC-25143: OpenSSL via Bazel @openssl (replaces omnibus source build).
# Mirrors origin/base-7.78.2's openssl3.rb, with --downloader_config=/dev/null
# on every bazelisk invocation (STS runner egress workaround, same pattern as
# STAC-24773 D1 in python3.rb / datadog-agent-dependencies.rb). zlib is now
# installed inline via @zlib//:install and is no longer an omnibus dependency.
#
# Version is kept at 3.5.7 to match deps/openssl/version.bzl (STS is a patch
# release ahead of upstream base-7.78.2's 3.5.6; bumped in commit 42e0e0b8bf
# for CVE-2025-9230 / openssl issue #30728).
default_version "3.5.7"

relative_path "openssl-#{version}"

build do
  flavor_flag = fips_mode? ? "--//packages/agent:flavor=fips" : ""

  command_on_repo_root "bazelisk run #{flavor_flag} --downloader_config=/dev/null -- @openssl//:install --destdir=#{install_dir}"

  unless windows?
    command_on_repo_root "bazelisk run --downloader_config=/dev/null -- @zlib//:install --destdir=#{install_dir}"
    # build_agent_dmg.sh sets INSTALL_DIR to some temporary folder.
    # This messes up openssl's internal paths. So we have to use another variable
    # so that replace_prefix and fix_openssl_paths set path correctly inside of the
    # openssl binaries on macos
    real_install_dir = if mac_os_x? then "/opt/datadog-agent" else install_dir end
    lib_extension = if linux_target? then ".so" else ".dylib" end

    files_to_patch = [
      "lib/libssl#{lib_extension}",
      "lib/libcrypto#{lib_extension}",
      "bin/openssl",
    ]
    if fips_mode?
      files_to_patch.append("lib/ossl-modules/*#{lib_extension}", "lib/engines-3/*#{lib_extension}")
    end

    files_to_patch = files_to_patch.map { |path| "#{install_dir}/embedded/#{path}" }

    command_on_repo_root "bazelisk run --downloader_config=/dev/null -- //bazel/rules:replace_prefix --prefix #{install_dir}/embedded #{files_to_patch.join(' ')}"

    command_on_repo_root "bazelisk run --downloader_config=/dev/null -- //deps/openssl:fix_openssl_paths --destdir #{real_install_dir}/embedded" \
      " #{install_dir}/embedded/lib/libssl#{lib_extension}" \
      " #{install_dir}/embedded/lib/libcrypto#{lib_extension}"
  end
  if fips_mode?
    command_on_repo_root "bazelisk run --downloader_config=/dev/null -- @openssl_fips//:install --destdir=#{install_dir}"
    if windows?
      command_on_repo_root "bazelisk run --downloader_config=/dev/null -- @openssl_fips//:configure_fips --destdir=\"#{install_dir}/embedded3\" --embedded_ssl_dir=\"C:/Program Files/Datadog/Datadog Agent/embedded3/ssl\""
    else
      command_on_repo_root "bazelisk run --downloader_config=/dev/null -- @openssl_fips//:configure_fips --destdir=#{install_dir}/embedded"
      command_on_repo_root "bazelisk run --downloader_config=/dev/null -- //bazel/rules:replace_prefix --prefix #{install_dir}/embedded" \
        " #{install_dir}/embedded/lib/ossl-modules/fips.so"
    end
  end
end
