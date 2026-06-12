name 'datadog-agent-integrations-py3-dependencies'

dependency 'pip3'
dependency 'setuptools3'

if linux_target?

# sts: not needed dependency 'freetds'  # needed for SQL Server integration
# sts: not needed dependency 'msodbcsql18' # needed for SQL Server integrat
# odbc drivers used by the SQL Server integration
#   dependency 'freetds'
#   unless heroku_target?
    # dependency 'msodbcsql18' # needed for SQL Server integration
#   endion
  dependency 'nfsiostat'
  # gstatus binary used by the glusterfs integration
  dependency 'gstatus'
end
