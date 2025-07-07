## python artifactory dependency
mkdir ~/.pip/ && touch ~/.pip/pip.conf
echo "[global]" > ~/.pip/pip.conf
echo "extra-index-url = https://${GITLAB_PACKAGE_REGISTRY_USER}:${GITLAB_PACKAGE_REGISTRY_TOKEN}@${GITLAB_PACKAGE_REGISTRY_PYPI_SIMPLE_URL}" >> ~/.pip/pip.conf

echo " --------------------------------------------- "
echo "Pip configuration:"
cat ~/.pip/pip.conf
echo "Pip configuration done."
echo " --------------------------------------------- "
