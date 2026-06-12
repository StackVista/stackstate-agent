## python artifactory dependency
# pip.conf gets the URL only — no credentials interpolated into the URL.
# Auth goes into ~/.netrc instead, which pip / pip-compile / pip-tools all
# honour for HTTP basic auth without echoing credentials into output
# requirements files. See STAC-24642.

missing=
for var in GITLAB_PACKAGE_REGISTRY_PYPI_SIMPLE_URL GITLAB_PACKAGE_REGISTRY_USER GITLAB_PACKAGE_REGISTRY_READONLY_PASSWORD; do
  if [ -z "${!var}" ]; then
    missing="$missing $var"
  fi
done
if [ -n "$missing" ]; then
  echo "ERROR: Required environment variables not set:$missing" >&2
  exit 1
fi

mkdir -p ~/.pip
cat > ~/.pip/pip.conf <<EOF
[global]
extra-index-url = https://${GITLAB_PACKAGE_REGISTRY_PYPI_SIMPLE_URL}
EOF

# Hostname only (strip the path that follows the first '/') for the netrc machine entry.
NETRC_HOST="${GITLAB_PACKAGE_REGISTRY_PYPI_SIMPLE_URL%%/*}"
cat > ~/.netrc <<EOF
machine ${NETRC_HOST}
login ${GITLAB_PACKAGE_REGISTRY_USER}
password ${GITLAB_PACKAGE_REGISTRY_READONLY_PASSWORD}
EOF
chmod 600 ~/.netrc


echo " --------------------------------------------- "
echo "Pip configuration:"
cat ~/.pip/pip.conf
echo "(auth for $NETRC_HOST in ~/.netrc, not shown)"
echo "Pip configuration done."
echo " --------------------------------------------- "

unset NETRC_HOST
