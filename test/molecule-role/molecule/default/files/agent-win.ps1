#Download and Run MSI package for Automated install
$uri = "https://s3-eu-west-1.amazonaws.com/stackstate-agent-2-test/windows/master/stackstate-agent-2.0.0.git.443.ef0c11ef-1-x86_64.msi"
$out = "c:\stackstate-agent.msi"

Function Download_MSI_STS_Installer{
Invoke-WebRequest -uri $uri -OutFile $out
$msifile = Get-ChildItem -Path $out -File -Filter '*.ms*'
write-host "StackState MSI $msifile "
}

Function Install_STS{
$FileExists = Test-Path $msifile -IsValid
$DataStamp = get-date -Format yyyyMMddTHHmmss
$logFile = '{0}-{1}.log' -f $msifile.fullname,$DataStamp
$stsHostname='agent-win'
$MSIArguments = @(
    "/i"
    ('"{0}"' -f $msifile.fullname)
    "/qn"
    "/norestart"
    "/L*v"
    $logFile
    " STS_API_KEY=API_KEY STS_URL=https://test-stackstate-agent.sts/stsAgent STS_HOSTNAME=$stsHostname SKIP_SSL_VALIDATION=true "
)
If ($FileExists -eq $True)
{
Start-Process "msiexec.exe" -ArgumentList $MSIArguments -passthru | wait-process
write-host "Finished msi "$msifile
}

Else {Write-Host "File doesn't exists"}
}
Download_MSI_STS_Installer
Install_STS
