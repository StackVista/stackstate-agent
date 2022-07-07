
#[Net.ServicePointManager]::SecurityProtocol = [Net.ServicePointManager]::SecurityProtocol -bor [Net.SecurityProtocolType]::Tls12
#Set-ExecutionPolicy Bypass -Scope Process -Force
#iex ((New-Object System.Net.WebClient).DownloadString('https://chocolatey.org/install.ps1'))


choco install -y awscli
#aws s3 cp s3://vcpython27/VCForPython27.msi C:\\Windows\\Temp\\VCForPython27.msi
#aws s3 cp s3://vcpython27/VC_redist.x86.exe C:\\Windows\\Temp\\VC_redist.x86.exe
#aws s3 cp s3://vcpython27/vs_Community.exe C:\\Windows\\Temp\\vs_Community.exe

# Install Gitlab Runner
choco install -y gitlab-runner

