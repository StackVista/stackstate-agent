#!/usr/bin/env ruby

# Copyright:: Your Company

name "libtool"
default_version "2.4.7"

license "GPL-2.0-or-later"
license_file "COPYING"
skip_transitive_dependency_licensing true

# Prefer the GNU mirror redirector to avoid flaky single-host timeouts
# You can pin a checksum by adding a version block with `source sha256:` if desired.
source url: "https://ftpmirror.gnu.org/libtool/libtool-#{version}.tar.xz"

relative_path "libtool-#{version}"

dependency "m4"

build do
  env = with_standard_compiler_flags(with_embedded_path)

  configure_options = [
    "--disable-nls",
  ]

  configure(*configure_options, env: env)
  make "-j #{workers}", env: env
  make "-j #{workers} install", env: env
end


