name "pylint2"
# Ship 1.x as 2.x only supports python 3
default_version "1.9.5"

dependency "pip2"

build do
  # pylint is only called in a subprocess by the Agent, so the Agent doesn't have to be GPL as well
  ship_license "https://dd-agent-omnibus.s3.amazonaws.com/licenses/GPL-2.0"

  # aliases for the pips
  if windows?
    pip2 = "#{windows_safe_path(python_2_embedded)}\\Scripts\\pip.exe"
    python2 = "#{windows_safe_path(python_2_embedded)}\\python.exe"
  else
    pip2 = "#{install_dir}/embedded/bin/pip2"
    python2 = "#{install_dir}/embedded/bin/python2"
  end

  # pin 2 dependencies of pylint:
  # - configparser: later versions (up to v3.7.1) are broken
  # - lazy-object-proxy 1.7.0 broken on python 2 https://github.com/ionelmc/python-lazy-object-proxy/issues/61
  if windows?
    command "#{python2} -m pip install configparser==3.5.0 lazy-object-proxy==1.6.0"
    command "#{python2} -m pip install pylint==#{version}"
  else
    command "#{pip2} install configparser==3.5.0 lazy-object-proxy==1.6.0"
    command "#{pip2} install pylint==#{version}"
  end
end
