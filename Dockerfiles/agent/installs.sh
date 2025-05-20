#/bin/bash

zypper --non-interactive refresh
zypper --non-interactive update
zypper --installroot /chroot -n in --no-recommends systemd libncurses6 libgcrypt20

rm -f /usr/sbin/runuser /usr/lib64/libdb-5.3.so

zypper --non-interactive clean --all

rm -rf /var/cache/zypp/* /tmp/* /var/tmp/*
