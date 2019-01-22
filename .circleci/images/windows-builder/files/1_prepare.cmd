set OMNIBUS_BASE_DIR_WIN=c:\omni-base\manual
set OMNIBUS_BASE_DIR_WIN_OMNIBUS=c:/omni-base/manual

set WIN_CI_PROJECT_DIR=c:\workspaces\builds\ToVYBrDV\0\stackvista\stackstate-agent\
set WORKON_HOME=%GOPATH%\src\github.com\StackVista\stackstate-agent

mkdir %GOPATH%\src\github.com\StackVista\stackstate-agent
xcopy /q/h/e/s %WIN_CI_PROJECT_DIR%* %GOPATH%\src\github.com\StackVista\stackstate-agent
cd %GOPATH%\src\github.com\StackVista\stackstate-agent
mkvirtualenv venv
cd %GOPATH%\src\github.com\StackVista\stackstate-agent
echo cd %GOPATH%\src\github.com\StackVista\stackstate-agent
pip install -r requirements.txt
