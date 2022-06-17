# Set administrator password
net user Administrator Passw0rd!
wmic useraccount where "name='Administrator'" set PasswordExpires=FALSE

# Disable Complex Passwords
# Reference: http://vlasenko.org/2011/04/27/removing-password-complexity-requirements-from-windows-server-2008-core/
$seccfg = [IO.Path]::GetTempFileName()
secedit /export /cfg $seccfg
(Get-Content $seccfg) | Foreach-Object {$_ -replace "PasswordComplexity\s*=\s*1", "PasswordComplexity=0"} | Set-Content $seccfg
secedit /configure /db $env:windir\security\new.sdb /cfg $seccfg /areas SECURITYPOLICY
del $seccfg
Write-Host "Complex Passwords have been disabled." -ForegroundColor Green

$admin = [adsi]("WinNT://./administrator, user")
$admin.PSBase.Invoke("SetPassword", "Passw0rd!")

net user ieuser 'Passw0rd!' /add /y
net localgroup administrators ieuser /add
net accounts /maxpwage:unlimited

# Disable Internet Explorer Security
# http://stackoverflow.com/a/9368555/2067999
# $AdminKey = "HKLM:\SOFTWARE\Microsoft\Active Setup\Installed Components\{A509B1A7-37EF-4b3f-8CFC-4F3A74704073}"
# $UserKey = "HKLM:\SOFTWARE\Microsoft\Active Setup\Installed Components\{A509B1A8-37EF-4b3f-8CFC-4F3A74704073}"
# Set-ItemProperty -Path $AdminKey -Name "IsInstalled" -Value 0
# Set-ItemProperty -Path $UserKey -Name "IsInstalled" -Value 0

add-type @"
using System.Net;
using System.Security.Cryptography.X509Certificates;
public class TrustAllCertsPolicy : ICertificatePolicy {
    public bool CheckValidationResult(
        ServicePoint srvPoint, X509Certificate certificate,
        WebRequest request, int certificateProblem) {
        return true;
    }
}
"@
$AllProtocols = [System.Net.SecurityProtocolType]'Ssl3,Tls,Tls11,Tls12'
[System.Net.ServicePointManager]::SecurityProtocol = $AllProtocols
[System.Net.ServicePointManager]::CertificatePolicy = New-Object TrustAllCertsPolicy

# Configure UAC to allow privilege elevation in remote shells
$Key = 'HKLM:\SOFTWARE\Microsoft\Windows\CurrentVersion\Policies\System'
$Setting = 'LocalAccountTokenFilterPolicy'
Set-ItemProperty -Path $Key -Name $Setting -Value 1 -Force

iex ((new-object net.webclient).DownloadString('https://chocolatey.org/install.ps1'))
choco feature enable -n allowGlobalConfirmation

Set-PSDebug -Trace 1

Set-ExecutionPolicy Bypass -Scope Process -Force;
./bootstrap_runner.ps1

# net user ${var.INSTANCE_USERNAME} '${var.INSTANCE_PASSWORD}' /add /y
# net localgroup administrators ${var.INSTANCE_USERNAME} /add

# Set-ExecutionPolicy Unrestricted -Scope LocalMachine -Force -ErrorAction Ignore

# $ErrorActionPreference = "stop"
