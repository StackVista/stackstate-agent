param([String]$goVersion="1.14.1")

choco feature enable -n allowGlobalConfirmation


Write-Host "Go Version: $goVersion"
choco install -y golang --version $goVersion
choco install -y python2 --version 2.7.14 --pkgparameters="/InstallDir:c:\\python27-x64"
choco install -y miniconda3 awscli rsat dotnetcore-sdk nuget.commandline conemu sysinternals dep cmake 7zip
