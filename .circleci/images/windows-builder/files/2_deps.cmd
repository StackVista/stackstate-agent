set OMNIBUS_BASE_DIR_WIN=c:\omni-base\manual
set OMNIBUS_BASE_DIR_WIN_OMNIBUS=c:/omni-base/manual

set WIN_CI_PROJECT_DIR=c:\workspaces\builds\ToVYBrDV\0\stackvista\stackstate-agent\
set WORKON_HOME=%GOPATH%\src\github.com\StackVista\stackstate-agent

cd %WORKON_HOME%
workon venv
inv -e deps
