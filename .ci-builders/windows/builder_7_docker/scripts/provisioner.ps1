# TODO Verify defender disabled
#Set-MpPreference -DisableRealtimeMonitoring $true
#Get-MpComputerStatus

#Enable-WindowsOptionalFeature -Online -FeatureName "NetFx3" -All -LogLevel WarningsInfo
Install-WindowsFeature NET-Framework-Core -Source C:\sxs

C:\Windows\Temp\VCForPython27.msi -ArgumentList "/quiet"

python -m pip install --upgrade pip

pip install virtualenvwrapper-win
