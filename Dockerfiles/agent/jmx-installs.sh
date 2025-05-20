#/bin/bash

zypper refresh

mkdir /usr/share/man/man1

zypper --installroot /chroot -n in --no-recommends java-1_8_0-openjdk-headless

zypper --non-interactive clean --all

rm -rf /var/cache/zypp/* /tmp/* /var/tmp/*
