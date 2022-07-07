FROM vinogradoff/windows-ruby:2.4.3.2


SHELL ["powershell", "-Command", "$ErrorActionPreference = 'Stop'; $ProgressPreference = 'SilentlyContinue';"]
WORKDIR C:/
ADD scripts .
ADD binaries C:/Windows/Temp
RUN C:/bootstrap.ps1
RUN C:/provisioner.ps1
