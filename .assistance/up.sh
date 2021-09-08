cd .assistance || exit

if [ -z "$1" ]; then
    export VERSION=2
else
    export VERSION=$1
fi

vagrant destroy --force && vagrant up
