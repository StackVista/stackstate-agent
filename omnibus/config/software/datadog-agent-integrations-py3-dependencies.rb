name 'datadog-agent-integrations-py3-dependencies'

dependency 'pip3'
dependency 'setuptools3'

# sts: not needed dependency 'confluent-kafka-python'

if arm_target?
  # same with libffi to build the cffi wheel
  dependency 'libffi'
  # same with libxml2 and libxslt to build the lxml wheel
  dependency 'libxml2'
  dependency 'libxslt'
end

if osx_target?
  dependency 'unixodbc'
end

if linux_target?
  # add nfsiostat script
  # sts: not needed dependency 'unixodbc'
  # sts: not needed dependency 'freetds'  # needed for SQL Server integration
  # sts: not needed dependency 'msodbcsql18' # needed for SQL Server integration
  dependency 'nfsiostat'
  # add libkrb5 for all integrations supporting kerberos auth with `requests-kerberos`
  dependency 'libkrb5'
  # needed for glusterfs
  dependency 'gstatus'
end

if redhat_target? && !arm_target?
  dependency 'pydantic-core-py3'
end

if linux_target?
  # We need to use cython<3.0.0 to build oracledb
  # sts: not needed dependency 'oracledb-py3'
end
