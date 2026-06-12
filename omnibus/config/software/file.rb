#
# Copyright 2012-2014 Chef Software, Inc.
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

name "file"
default_version "5.46"

dependency 'zlib'
dependency 'bzip2'
dependency 'liblzma'

license "BSD"
license_file "COPYING"
skip_transitive_dependency_licensing true

version("5.46") { source sha256: "c9cc77c7c560c543135edc555af609d5619dbef011997e988ce40a3d75d86088" }

source url: "https://distfiles.macports.org/file/file-#{version}.tar.gz"

relative_path "file-#{version}"

build do
  env = with_standard_compiler_flags(with_embedded_path)

  configure_options = []
  configure(*configure_options, env: env)

  make "-j #{workers}", env: env
  make "install", env: env

  # The file binary depends on libseccomp.so.2 which is a system library.
  # This is safe as libseccomp is widely available on Linux systems.
  whitelist_file "#{install_dir}/embedded/bin/file"
end
