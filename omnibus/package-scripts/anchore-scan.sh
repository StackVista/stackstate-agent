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

FILE="anchore-scan.txt"
ANCHORE="anchore/engine-cli:v0.9.2"
ANCHORE_DOCKER_INVOKE="docker run --rm -a stdout -e ANCHORE_CLI_USER=${ANCHORE_CLI_USER} -e ANCHORE_CLI_PASS=${ANCHORE_CLI_PASS} -e ANCHORE_CLI_URL=${ANCHORE_CLI_URL} ${ANCHORE}"
ANCHORE_PARSE="quay.io/stackstate/anchore-parser:c1c93c53"

${ANCHORE_DOCKER_INVOKE} anchore-cli image add "$IMAGE" > /dev/null
${ANCHORE_DOCKER_INVOKE} anchore-cli image wait "$IMAGE" > /dev/null
${ANCHORE_DOCKER_INVOKE} anchore-cli image vuln --vendor-only false "$IMAGE" all > $FILE
${ANCHORE_DOCKER_INVOKE} anchore-cli evaluate check "$IMAGE" --policy "cluster-agent-04x" --detail

if [ ! -f ${FILE} ]; then
    echo "File ${FILE} not found!"
    exit 1
fi

mkdir -p "${PWD}"/anchore-output
mv ${FILE} anchore-output

docker run --rm \
   -e ANCHORE_WEBHOOK="${ANCHORE_WEBHOOK}" \
   -e INPUT_FILE="anchore-output/${FILE}" \
   -e IMAGE_WHITELIST_FILE="anchore-whitelists/image-whitelist.json" \
   -e CVE_WHITELIST_FILE="anchore-whitelists/cve-whitelist.json" \
   -e WHITELIST_IMAGES_HAVE_TAGS="false" \
   -v "${PWD}"/anchore-whitelists:/usr/src/app/anchore-whitelists \
   -v "${PWD}"/anchore-output:/usr/src/app/anchore-output \
   ${ANCHORE_PARSE} python daily_high_crit_report.py

rm -rf anchore-output
