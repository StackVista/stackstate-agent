#!/usr/bin/env ruby

# Copyright:: Your Company

name "libtool"
default_version "2.4.7"

license "GPL-2.0-or-later"
license_file "COPYING"
skip_transitive_dependency_licensing true

# Pin checksums per version to satisfy Omnibus security checks
version("2.4.7") { source sha256: "4f7f217f057ce655ff22559ad221a0fd8ef84ad1fc5fcb6990cecc333aa1635d" }

# Use a stable mirror for libtool source
source url: "https://mirror.clientvps.com/gnu/libtool/libtool-#{version}.tar.xz"

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


