#!/bin/bash

read -p "Did you run 'beest cleanup' and 'beest destroy' before running this cleanup script [Y/N]? " -n 1 -r
echo
if [[ ! $REPLY =~ ^[Yy]$ ]]
then
    echo "Please destroy your resources before cleaning up the beest files."
    exit 1
fi

function cleanup_test_directories () {
    # Delete Files
    find "./tests" -type f -name '*.gv' "-$1"
    find "./tests" -type f -name 'test_*-*.json' "-$1"
    find "./tests" -type f -name 'test_*-*.xml' "-$1"

    # Delete Directories
    if [ $1 == "delete" ]; then
        find "./tests" -type d -name '.pytest_cache' -exec rm -rf {} +
    else
        find "./tests" -type d -name '.pytest_cache' "-$1"
    fi
}

function cleanup_sut_directories () {
    # Delete Files
    find "./sut/yards" -type f -name 'ansible_inventory' "-$1"
    find "./sut/yards" -type f -name 'conf.yaml' "-$1"
    find "./sut/yards" -type f -name '*_id_rsa' "-$1"
    find "./sut/yards" -type f -name 'stackstate-values.yml' "-$1"
    find "./sut/yards" -type f -name 'sts-toolbox.yml' "-$1"
    find "./sut/yards" -type f -name 'tf.deploy' "-$1"

    # Delete Directories
    if [ $1 == "delete" ]; then
        find "./sut/yards" -type d -name '.terraform' -exec rm -rf {} +
    else
        find "./sut/yards" -type d -name '.terraform' "-$1"
    fi
}

cleanup_test_directories "print"
cleanup_sut_directories "print"

echo
read -p "Continue with deleting the files mentioned above [Y/N]? " -n 1 -r
echo
if [[ ! $REPLY =~ ^[Yy]$ ]]
then
    exit 1
fi

cleanup_test_directories "delete"
cleanup_sut_directories "delete"

echo "Beest files has been deleted and cleaned-up"
