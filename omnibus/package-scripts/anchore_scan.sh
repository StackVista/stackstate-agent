#!/bin/sh

usage()
{
    echo "Usage: $0 [-i <string>] [-n <0|1>]" 1>&2
    echo "       -i image: docker image to scan for vulnerabilities"
    echo "       -n notify-medium: Whether or not a slack message is sent for medium vulnerabilities"
    exit 1
}

IMAGE=""

while getopts ":i:n:" o; do
    case "${o}" in
        i)
            IMAGE=${OPTARG}
            ;;
        n)
            NOTIFY=${OPTARG}
            ;;
        *)
            usage
            ;;
    esac
done
shift $((OPTIND-1))

if [ -z "${IMAGE}" ] || [ -z "${NOTIFY}" ]; then
    usage
fi

# Split by colon and return first element
REPO_PATH=${IMAGE%:*}
# Replace all forward slashes with double underscores to prevent docker image looking like a directory path
# Double underscore is used by the anchore-parser project to reconstruct the image name (replaced with /)
REPO_PATH=${REPO_PATH//\//__}

# Split by colon and return last element
IMAGE_TAG=${IMAGE#*:}

FILE="${REPO_PATH}_${IMAGE_TAG}.json"
ANCHORE="anchore/engine-cli:v0.9.2"
ANCHORE_DOCKER_INVOKE="docker run --rm -a stdout -e ANCHORE_CLI_USER=${ANCHORE_CLI_USER} -e ANCHORE_CLI_PASS=${ANCHORE_CLI_PASS} -e ANCHORE_CLI_URL=${ANCHORE_CLI_URL} ${ANCHORE}"
ANCHORE_PARSE="quay.io/stackstate/anchore-parser:4d710056"

${ANCHORE_DOCKER_INVOKE} anchore-cli image add "$IMAGE"
${ANCHORE_DOCKER_INVOKE} anchore-cli image wait "$IMAGE"
${ANCHORE_DOCKER_INVOKE} anchore-cli --json image vuln --vendor-only false "$IMAGE" all > "${FILE}"
${ANCHORE_DOCKER_INVOKE} anchore-cli evaluate check "$IMAGE" --policy "cluster-agent-04x" --detail


if [ ! -f ${FILE} ]; then
    echo "File ${FILE} not found!"
    exit 1
fi

APP_DIR=$(dirname "${0}")
EXEC_DIR="${PWD}/${APP_DIR}"

mkdir -p "${EXEC_DIR}"/anchore_output
mv ${FILE} "${EXEC_DIR}"/anchore_output

docker run --rm \
   -e ANCHORE_WEBHOOK="${ANCHORE_WEBHOOK}" \
   -e INPUT_DIR="anchore_output" \
   -e IMAGE_WHITELIST_FILE="anchore-whitelists/image-whitelist.json" \
   -e MIXED_WHITELIST_FILE="anchore-whitelists/image-cve-mixed.json" \
   -e WHITELIST_IMAGES_HAVE_TAGS="false" \
   -v "${EXEC_DIR}"/anchore-whitelists:/usr/src/app/anchore-whitelists \
   -v "${EXEC_DIR}"/anchore_output:/usr/src/app/anchore_output \
   ${ANCHORE_PARSE} python reports/json_parsed/os_level_cves/cve_reports/json_os_high_crit_report.py

docker run --rm \
   -e ANCHORE_WEBHOOK="${ANCHORE_WEBHOOK}" \
   -e INPUT_DIR="anchore_output" \
   -e IMAGE_WHITELIST_FILE="anchore-whitelists/image-whitelist.json" \
   -e MIXED_WHITELIST_FILE="anchore-whitelists/image-cve-mixed.json" \
   -e WHITELIST_IMAGES_HAVE_TAGS="false" \
   -v "${EXEC_DIR}"/anchore-whitelists:/usr/src/app/anchore-whitelists \
   -v "${EXEC_DIR}"/anchore_output:/usr/src/app/anchore_output \
   ${ANCHORE_PARSE} python reports/json_parsed/non_os_level_cves/cve_reports/json_nos_high_crit_report.py

rm -rf "${EXEC_DIR}"/anchore_output
