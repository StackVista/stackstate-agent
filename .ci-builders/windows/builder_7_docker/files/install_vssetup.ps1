Set-StrictMode -Version Latest
$ErrorActionPreference = "Stop"
$PSDefaultParameterValues['*:ErrorAction'] = 'Stop'

[Net.ServicePointManager]::SecurityProtocol = [Net.ServicePointManager]::SecurityProtocol -bor [Net.SecurityProtocolType]::Tls12

Set-PSRepository -Name "PSGallery" -InstallationPolicy Trusted
Install-Module Pscx -AllowClobber
Install-Module VSSetup -Scope CurrentUser
