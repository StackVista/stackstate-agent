## python artifactory dependency
mkdir ~/.pip/ && touch ~/.pip/pip.conf
echo "[global]" > ~/.pip/pip.conf
export URL_TO_USE="${GITLAB_PACKAGE_REGISTRY_PYPI_URL}"
if [[ "${GITLAB_PACKAGE_REGISTRY_PYPI_URL%/}" != */simple ]]; then
    export URL_TO_USE=${GITLAB_PACKAGE_REGISTRY_PYPI_URL}/simple
fi
echo "extra-index-url = https://${GITLAB_PACKAGE_REGISTRY_USER}:${GITLAB_PACKAGE_REGISTRY_TOKEN}@${URL_TO_USE}" >> ~/.pip/pip.conf

echo " --------------------------------------------- "
echo "Pip configuration:"
cat ~/.pip/pip.conf
echo "Pip configuration done."
echo " --------------------------------------------- "
