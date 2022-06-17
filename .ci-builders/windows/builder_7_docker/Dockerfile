FROM vinogradoff/windows-ruby:2.4.3.2


SHELL ["powershell", "-Command", "$ErrorActionPreference = 'Stop'; $ProgressPreference = 'SilentlyContinue';"]
WORKDIR C:/
ADD scripts .
RUN C:/bootstrap.ps1
