cd .assistance || exit 0
. ./env
case $1 in
        "provision")
                vagrant provision --provision-with "$2"
                ;;
        "destroy")
                vagrant destroy --force && vagrant up
                ;;
esac
