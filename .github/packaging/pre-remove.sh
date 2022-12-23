#!/bin/sh

service_name=syringe@127.0.0.1

remove() {
    printf "\033[32m Pre Remove of a normal remove\033[0m\n"
    systemctl stop ${service_name} ||:
    rm -f /usr/lib/systemd/system/syringe@.service
}

purge() {
    printf "\033[32m Pre Remove purge, deb only\033[0m\n"
    systemctl stop ${service_name} ||:
    rm -f /usr/lib/systemd/system/syringe@.service
}

upgrade() {
    printf "\033[32m Pre Remove of an upgrade\033[0m\n"
    systemctl stop ${service_name} ||:
}

echo "$@"

action="$1"

case "$action" in
  "0" | "remove")
    remove
    ;;
  "1" | "upgrade")
    upgrade
    ;;
  "purge")
    purge
    ;;
  *)
    remove
    ;;
esac