REM set WIN_CI_PROJECT_DIR=%CD%
REM set WORKON_HOME=%WIN_CI_PROJECT_DIR%

set

dir

call %WORKON_HOME%\venv\Scripts\activate.bat

inv -e deps
