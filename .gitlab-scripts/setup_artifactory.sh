## python artifactory dependency
mkdir ~/.pip/ && touch ~/.pip/pip.conf
echo "[global]" > ~/.pip/pip.conf
echo "extra-index-url = https://$artifactory_user:$artifactory_password@$ARTIFACTORY_PYPI_URL/simple" >> ~/.pip/pip.conf
