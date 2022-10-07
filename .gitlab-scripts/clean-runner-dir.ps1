echo "Getting runner name"
$runner=$(.\.gitlab-scripts\get-runner-name.ps1)

echo "Runner Name is $runner"

echo "Changing run directory to C:\workspaces\builds"
cd C:\workspaces\builds

sleep 10
echo "Attempt to fix folder attributes"
cmd /c "attrib -r /s $runner"

echo "Deleting runner folder: Get-ChildItem -Path C:\workspaces\builds\$runner -Recurse | Remove-Item -force -recurse"
Get-ChildItem -Path C:\workspaces\builds\$runner -Recurse | Remove-Item -force -recurse

