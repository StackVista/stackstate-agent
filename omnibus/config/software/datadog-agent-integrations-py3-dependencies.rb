name 'datadog-agent-integrations-py3-dependencies'

dependency 'python3'
dependency 'setuptools3'
dependency 'openssl3'

if linux_target?

  # [STS] freetds, msodbcsql18, krb5, unixodbc not needed — STS integrations don't use SQL Server / Kerberos
  build do
    # gstatus binary used by the glusterfs integration
    command_on_repo_root "bazelisk run -- //deps/gstatus:install --destdir='#{install_dir}'"
    command_on_repo_root "bazelisk run -- //deps/nfsiostat:install --destdir='#{install_dir}'"
  end
end
