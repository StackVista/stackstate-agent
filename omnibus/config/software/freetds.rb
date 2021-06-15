name "freetds"
default_version "1.1.36"

version "1.1.36" do
  source sha256: "d6ff072a7a37084baaead44785e03e13f551a527e36e3b5b715049b1ff3e59cc"
end

source url: "ftp://ftp.freetds.org/pub/freetds/stable/freetds-#{version}.tar.gz"

relative_path "freetds-#{version}"

build do
  ship_license "./COPYING"
  env = with_standard_compiler_flags(with_embedded_path)

  configure_args = [
    "--disable-readline",
  ]

  configure_command = configure_args.unshift("./configure").join(" ")

  command configure_command, env: env, in_msys_bash: true
  make env: env

  # Only `libtdsodbc.so/libtdsodbc.so.0.0.0` are needed for SQLServer integration.
  # Hence we only need to copy those.
  copy "src/odbc/.libs/libtdsodbc.so", "#{install_dir}/embedded/lib/libtdsodbc.so"
  copy "src/odbc/.libs/libtdsodbc.so.0.0.0", "#{install_dir}/embedded/lib/libtdsodbc.so.0.0.0"

end
