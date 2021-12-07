if exist %APPDATA%\pip rd /s/q %APPDATA%\pip
mkdir %APPDATA%\pip
echo [global] > %APPDATA%\pip\pip.ini
echo extra-index-url = https://%artifactory_user%:%artifactory_password%@%ARTIFACTORY_PYPI_URL%/simple >> %APPDATA%\pip\pip.ini
