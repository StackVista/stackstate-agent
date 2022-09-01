python -m pip install --upgrade pip

pip install virtualenvwrapper-win

# choco install -v -y --ignore-checksums visualstudio2017community --package-parameters "--allWorkloads --includeRecommended --includeOptional --locale en-US"

C:\binaries\vs_buildtools.exe --quiet --wait --norestart --nocache --installPath C:\BuildTools --add Microsoft.VisualStudio.Workload.AzureBuildTools --remove Microsoft.VisualStudio.Component.Windows10SDK.10240 --remove Microsoft.VisualStudio.Component.Windows10SDK.10586 --remove Microsoft.VisualStudio.Component.Windows10SDK.14393 --remove Microsoft.VisualStudio.Component.Windows81SDK
C:\binaries\vs_buildtools2019.exe --quiet --wait --norestart --nocache --installPath "%ProgramFiles(x86)%\Microsoft Visual Studio\2019\BuildTools" --add Microsoft.VisualStudio.Workload.AzureBuildTools --remove Microsoft.VisualStudio.Component.Windows10SDK.10240 --remove Microsoft.VisualStudio.Component.Windows10SDK.10586 --remove Microsoft.VisualStudio.Component.Windows10SDK.14393 --remove Microsoft.VisualStudio.Component.Windows81SDK
C:\binaries\dotNetFx40_Full_setup.exe /q

#choco install -v -y visualstudio2017buildtools --pkgparameters="--norestart --wait --quiet --locale en-US --add Microsoft.VisualStudio.Workload.MSBuildTools --add Microsoft.VisualStudio.Workload.UniversalBuildTools --add Microsoft.VisualStudio.Workload.VCTools --add Microsoft.VisualStudio.ComponentGroup.NativeDesktop.Win81 --add Microsoft.VisualStudio.Workload.VCTools --add Microsoft.VisualStudio.Component.VC.Runtimes.x86.x64.Spectre --add Microsoft.VisualStudio.Component.Windows10SDK.17763"

#choco install -v -y --ignore-checksums visualstudio2019community --pkgparameters="--norestart --quiet --wait --locale en-US --add Microsoft.VisualStudio.Workload.MSBuildTools --add Microsoft.VisualStudio.Workload.NativeDesktop  --add Microsoft.VisualStudio.Component.VC.CMake.Project --add Microsoft.VisualStudio.Component.VC.CLI.Support --add Microsoft.VisualStudio.Workload.UniversalBuildTools --add Microsoft.VisualStudio.Workload.VCTools --add Microsoft.VisualStudio.ComponentGroup.NativeDesktop.Win81"

./files/InstallWixExtension.ps1

./files/InstallWixExtension2017.ps1

choco install -v -y --ignore-checksums visualstudio2019-workload-manageddesktop visualstudio2019-workload-nativedesktop visualstudio2019-workload-netcoretools visualstudio2019-workload-netweb visualstudio2019-workload-nativecrossplat visualstudio2019-workload-vctools netfx-4.7-devpack hg pkgconfiglite wget git wixtoolset
