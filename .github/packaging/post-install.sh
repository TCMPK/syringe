#!/bin/sh

service_name=syringe@127.0.0.1

cleanup() {
    # This is where you remove files that were not needed on this platform / system
}

cleanInstall() {
    printf "\033[32m Post Install of an clean install\033[0m\n"
    printf "\033[32m Reload the service unit from disk\033[0m\n"
    systemctl daemon-reload ||:
    printf "\033[32m Unmask the service\033[0m\n"
    systemctl unmask ${service_name} ||:
    printf "\033[32m Set the preset flag for the service unit\033[0m\n"
    systemctl preset ${service_name} ||:
    printf "\033[32m Set the enabled flag for the service unit\033[0m\n"
    systemctl enable ${service_name} ||:
    systemctl restart ${service_name} ||:
}

upgrade() {
    printf "\033[32m Post Install of an upgrade\033[0m\n"
    systemctl restart ${service_name} ||:
}

# Step 2, check if this is a clean install or an upgrade
action="$1"
if  [ "$1" = "configure" ] && [ -z "$2" ]; then
  # Alpine linux does not pass args, and deb passes $1=configure
  action="install"
elif [ "$1" = "configure" ] && [ -n "$2" ]; then
    # deb passes $1=configure $2=<current version>
    action="upgrade"
fi

case "$action" in
  "1" | "install")
    cleanInstall
    ;;
  "2" | "upgrade")
    upgrade
    ;;
  *)
    # $1 == version being installed
    cleanInstall
    ;;
esac

cleanup