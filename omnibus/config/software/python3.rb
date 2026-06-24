name "python3"

default_version "3.13.14"

# [sts] STAC-24773 Phase D1: Python via Bazel @cpython (replaces omnibus source build).
# Mirrors origin/base-7.78.2 with --downloader_config=/dev/null on every bazelisk
# invocation (STS runner egress workaround; see datadog-agent-dependencies.rb).
if heroku_target?
  flavor_flag = "--//packages/agent:flavor=heroku"
else
  flavor_flag = fips_mode? ? "--//packages/agent:flavor=fips" : ""
end

unless windows?
  build do
    # Temporary deps. When we fix auto-rpath fixing these will disappear.
    command_on_repo_root "bazelisk run #{flavor_flag} --downloader_config=/dev/null -- @bzip2//:install --destdir='#{install_dir}'"

    command_on_repo_root "bazelisk run #{flavor_flag} --downloader_config=/dev/null -- @xz//:install --destdir='#{install_dir}'"
    sh_lib = if linux_target? then "liblzma.so" else "liblzma.dylib" end
    command_on_repo_root "bazelisk run #{flavor_flag} --downloader_config=/dev/null -- //bazel/rules:replace_prefix --prefix '#{install_dir}/embedded' " \
      "#{install_dir}/embedded/lib/#{sh_lib}"

    command_on_repo_root "bazelisk run #{flavor_flag} --downloader_config=/dev/null -- @sqlite3//:install --destdir='#{install_dir}'"
    sh_lib = if linux_target? then "libsqlite3.so" else "libsqlite3.dylib" end
    command_on_repo_root "bazelisk run #{flavor_flag} --downloader_config=/dev/null -- //bazel/rules:replace_prefix --prefix '#{install_dir}/embedded' " \
       "#{install_dir}/embedded/lib/#{sh_lib}"
  end
end
dependency "openssl3"

build do
  # 2.0 is the license version here, not the python version
  license "Python-2.0"

  if !windows_target?
    env = with_standard_compiler_flags(with_embedded_path)
    command_on_repo_root "bazelisk run #{flavor_flag} --downloader_config=/dev/null -- @cpython//:install --destdir='#{install_dir}'"
    sh_ext = if linux_target? then "so" else "dylib" end
    command_on_repo_root "bazelisk run #{flavor_flag} --downloader_config=/dev/null -- //bazel/rules:replace_prefix --prefix '#{install_dir}/embedded'" \
      " #{install_dir}/embedded/lib/libpython3.*#{sh_ext}" \
      " #{install_dir}/embedded/lib/python3.13/lib-dynload/*.so" \
      " #{install_dir}/embedded/bin/python3*"
    python = "#{install_dir}/embedded/bin/python3"

    # Python curses modules depend on system ncurses libraries (libncursesw.so.6, …)
    # which are widely available on Linux systems. These are safe dependencies.
    if linux_target?
      major, minor, = version.split(".")
      block do
        Dir.glob("#{install_dir}/embedded/lib/python#{major}.#{minor}/lib-dynload/_curses*.so").each do |file|
          whitelist_file file
        end
        Dir.glob("#{install_dir}/embedded/lib/python#{major}.#{minor}/lib-dynload/_curses_panel*.so").each do |file|
          whitelist_file file
        end
      end
    end
  else
    command_on_repo_root "bazelisk run #{flavor_flag} --downloader_config=/dev/null -- @cpython//:install --destdir=#{install_dir}"
    python = "#{windows_safe_path(python_3_embedded)}\\python.exe"
  end

  # Upgrade pip to 26.0.1 to address CVE-2026-1703 (path traversal in pip < 26.0
  # when installing malicious wheel archives). Python 3.13 ships with pip 25.3 via
  # ensurepip, which is vulnerable. Replaces the deleted omnibus pip3.rb recipe.
  command "#{python} -m pip install pip==26.0.1"

  # @cpython//:install creates pip3 -> pip{major}.{minor}, but pip self-upgrade
  # only refreshes the versioned script and drops the unversioned symlinks.
  # stackstate-agent-integrations-py3.rb invokes embedded/bin/pip3 directly.
  if !windows_target?
    major, minor, = version.split(".")
    block "recreate pip symlinks after pip self-upgrade" do
      Dir.chdir "#{install_dir}/embedded/bin" do
        # File.exist? is false for broken symlinks; pip self-upgrade can leave pip -> pip3 dangling.
        ["pip3", "pip"].each do |f|
          File.delete(f) if File.exist?(f) || File.symlink?(f)
        end
        File.symlink "pip#{major}.#{minor}", "pip3"
        File.symlink "pip3", "pip"
      end
    end
  end
end
