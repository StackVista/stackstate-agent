REM set WIN_CI_PROJECT_DIR=%CD%
REM set WORKON_HOME=%WIN_CI_PROJECT_DIR%

IF EXIST c:\deps GOTO C_DEPS_EXIST
call %WIN_CI_PROJECT_DIR%\pkg_configs.cmd
:C_DEPS_EXIST

if exist .omnibus rd /s/q .omnibus
mkdir .omnibus\pkg
if exist \omnibus-ruby rd /s/q \omnibus-ruby
if exist %OMNIBUS_BASE_DIR_WIN% rd /s/q %OMNIBUS_BASE_DIR_WIN%
if exist \opt\stackstate-agent rd /s/q \opt\stackstate-agent
if exist %GOPATH%\src\github.com\StackVista\stackstate-agent rd /s/q %GOPATH%\src\github.com\StackVista\stackstate-agent
mkdir %GOPATH%\src\github.com\StackVista\stackstate-agent
xcopy /q/h/e/s * %GOPATH%\src\github.com\StackVista\stackstate-agent
cd %GOPATH%\src\github.com\StackVista\stackstate-agent
call mkvirtualenv venv
cd %GOPATH%\src\github.com\StackVista\stackstate-agent
echo cd %GOPATH%\src\github.com\StackVista\stackstate-agent
pip install -r requirements.txt

