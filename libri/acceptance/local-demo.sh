#!/usr/bin/env bash

set -eou pipefail
#set -x  # useful for debugging

# optional settings (generally defaults should be fine, but sometimes useful for debugging)
LIBRI_LOG_LEVEL="${LIBRI_LOG_LEVEL:-INFO}"  # or DEBUG
LIBRI_TIMEOUT="${LIBRI_TIMEOUT:-5}"  # 10, or 20 for really sketchy network

# local and filesystem constants
LOCAL_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
LOCAL_TEST_DATA_DIR="${LOCAL_DIR}/data"
LOCAL_TEST_LOGS_DIR="${LOCAL_DIR}/logs"
mkdir -p "${LOCAL_TEST_LOGS_DIR}"

# determine which md5 command to use
MD5_CMD="md5sum"
if ( ! command -v ${MD5_CMD} > /dev/null ) && command -v "md5" > /dev/null; then
    # use BSD md5 if md5sum isn't available
    MD5_CMD="md5 -q"
fi
if ! command -v ${MD5_CMD} > /dev/null; then
    echo 'unable to determine which MD5 command to use'
    exit 1
fi

# get test data if it doesn't exist
if [[ ! -d "${LOCAL_TEST_DATA_DIR}" ]]; then
    ${LOCAL_DIR}/get-test-data.sh
fi

# container command constants
IMAGE="daedalus2718/libri:snapshot"
KEYCHAIN_DIR="/keychains"  # inside container
CONTAINER_TEST_DATA_DIR="/test-data"
LIBRI_PASSPHRASE="test passphrase"  # bypass command-line entry
N_LIBRARIANS=4

# clean up any existing libri containers
echo "cleaning up existing network and containers..."
docker ps | grep 'libri' | awk '{print $1}' | xargs -I {} docker stop {} || true
docker ps -a | grep 'libri' | awk '{print $1}' | xargs -I {} docker rm {} || true
docker network list | grep 'libri' | awk '{print $2}' | xargs -I {} docker network rm {} || true

echo
echo "creating libri docker network..."
docker network create libri

echo
echo "starting librarian peers..."
librarian_addrs=""
librarian_containers=""
for c in $(seq 0 $((${N_LIBRARIANS} - 1))); do
    port=$((20100+c))
    metricsPort=$((20200+c))
    name="librarian-${c}"
    docker run --name "${name}" --net=libri -d -p ${port}:${port} ${IMAGE} \
        librarian start \
        --nSubscriptions 2 \
        --logLevel "${LIBRI_LOG_LEVEL}" \
        --publicPort ${port} \
        --publicHost ${name} \
        --localPort ${port} \
        --localMetricsPort ${metricsPort} \
        --bootstraps "librarian-0:20100"
    if [[ c -eq 0 ]]; then
       librarian_addrs="${name}:${port}"
    else
       librarian_addrs="${name}:${port},${librarian_addrs}"
    fi
    librarian_containers="${name} ${librarian_containers}"
done
sleep 5  # TODO (drausin) add retry to healthcheck

echo
echo "testing librarians health..."
docker run --rm --net=libri ${IMAGE} test health \
    -a "${librarian_addrs}" \
    --logLevel "${LIBRI_LOG_LEVEL}" \
    --timeout "${LIBRI_TIMEOUT}"

echo
echo "testing librarians upload/download..."
docker run --rm --net=libri ${IMAGE} test io \
    -a "${librarian_addrs}" \
    -n 4 \
    --logLevel "${LIBRI_LOG_LEVEL}" \
    --timeout "${LIBRI_TIMEOUT}"

echo
echo "initializing author..."
docker create \
    --name author-data \
    -v ${KEYCHAIN_DIR} \
    -v ${CONTAINER_TEST_DATA_DIR} \
    -e LIBRI_PASSPHRASE="${LIBRI_PASSPHRASE}" \
    ${IMAGE}

docker cp ${LOCAL_TEST_DATA_DIR}/* author-data:${CONTAINER_TEST_DATA_DIR}
docker run \
    --rm \
     --net=libri \
    --volumes-from author-data \
    -e LIBRI_PASSPHRASE="${LIBRI_PASSPHRASE}" \
    ${IMAGE} \
    author init -k "${KEYCHAIN_DIR}"

echo
echo "uploading & downloading local files..."
for file in $(ls ${LOCAL_TEST_DATA_DIR}); do
    up_file="${CONTAINER_TEST_DATA_DIR}/${file}"
    docker run \
        --rm \
        --net=libri \
        --volumes-from author-data \
        -e LIBRI_PASSPHRASE="${LIBRI_PASSPHRASE}" \
        ${IMAGE} \
        author upload \
        -k "${KEYCHAIN_DIR}" \
        -a "${librarian_addrs}" \
        -f "${up_file}"  \
        --timeout "${LIBRI_TIMEOUT}" \
        --logLevel "${LIBRI_LOG_LEVEL}" 2>&1 | \
        tee ${LOCAL_TEST_LOGS_DIR}/${file}.log

    log_file="${LOCAL_TEST_LOGS_DIR}/${file}.log"
    down_file="${CONTAINER_TEST_DATA_DIR}/downloaded.${file}"
    envelope_key=$(grep envelope_key ${log_file} | sed -E 's/.*"envelope_key": "([^ "]*).*/\1/g')
    docker run \
        --rm \
        --net=libri \
        --volumes-from author-data \
        -e LIBRI_PASSPHRASE="${LIBRI_PASSPHRASE}" \
        ${IMAGE} \
        author download -k "${KEYCHAIN_DIR}" -a "${librarian_addrs}" -f "${down_file}" -e "${envelope_key}"

    # verify md5s (locally, since it's simpler)
    docker cp "author-data:${down_file}" "${LOCAL_TEST_DATA_DIR}/downloaded.${file}"
    up_md5=$(${MD5_CMD} "${LOCAL_TEST_DATA_DIR}/${file}" | awk '{print $1}')
    down_md5=$(${MD5_CMD} "${LOCAL_TEST_DATA_DIR}/downloaded.${file}" | awk '{print $1}')
    echo "uploaded MD5: ${up_md5}"
    echo "downloaded MD5: ${down_md5}"
    [[ "${up_md5}" = "${down_md5}" ]]
done

echo
echo "cleaning up..."
rm -f ${LOCAL_TEST_DATA_DIR}/downloaded.*
rm -f ${LOCAL_TEST_LOGS_DIR}/*
docker ps | grep 'libri' | awk '{print $1}' | xargs -I {} docker stop {} || true
docker ps -a | grep 'libri' | awk '{print $1}' | xargs -I {} docker rm {} || true
docker network list | grep 'libri' | awk '{print $2}' | xargs -I {} docker network rm {} || true

echo
echo "All tests passed."
