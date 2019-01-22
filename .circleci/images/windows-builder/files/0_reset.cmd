set OMNIBUS_BASE_DIR_WIN=c:\omni-base\manual
set OMNIBUS_BASE_DIR_WIN_OMNIBUS=c:/omni-base/manual

set WIN_CI_PROJECT_DIR=c:\workspaces\builds\ToVYBrDV\0\stackvista\stackstate-agent\
set WORKON_HOME=%GOPATH%\src\github.com\StackVista\stackstate-agent

if exist .omnibus rd /s/q .omnibus
mkdir .omnibus\pkg
if exist \omnibus-ruby rd /s/q \omnibus-ruby
if exist %OMNIBUS_BASE_DIR_WIN% rd /s/q %OMNIBUS_BASE_DIR_WIN%
if exist \opt\stackstate-agent rd /s/q \opt\stackstate-agent
if exist %GOPATH%\src\github.com\StackVista\stackstate-agent rd /s/q %GOPATH%\src\github.com\StackVista\stackstate-agent
if exist %WORKON_HOME%\venv rd /s/q %WORKON_HOME%\venv
if exist %WIN_CI_PROJECT_DIR%\venv rd /s/q %WIN_CI_PROJECT_DIR%\venv
