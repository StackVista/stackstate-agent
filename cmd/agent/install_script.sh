#!/bin/bash
# (C) Datadog, Inc. 2010-2016
# (C) StackState
# All rights reserved
# Licensed under Simplified BSD License (see LICENSE)
# StackState Agent installation script: install and set up the Agent on supported Linux distributions
# using the package manager and StackState repositories.

set -e
install_script_version=1.1.0
logfile="ddagent-install.log"

PKG_NAME="stackstate-agent"
PKG_USER="stackstate-agent"
ETCDIR="/etc/stackstate-agent"
CONF="$ETCDIR/stackstate.yaml"

logfile="$PKG_NAME-install.log"

if [ $(command -v curl) ]; then
    dl_cmd="curl -f"
else
    dl_cmd="wget --quiet"
fi

# A2923DFF56EDA6E76E55E492D3A80E30382E94DE expires in 2022
# D75CEA17048B9ACBF186794B32637D44F14F620E expires in 2032
APT_GPG_KEYS=("A2923DFF56EDA6E76E55E492D3A80E30382E94DE" "D75CEA17048B9ACBF186794B32637D44F14F620E")

# DATADOG_RPM_KEY_E09422B3.public expires in 2022
# DATADOG_RPM_KEY_20200908.public expires in 2024
RPM_GPG_KEYS=("DATADOG_RPM_KEY_E09422B3.public" "DATADOG_RPM_KEY_20200908.public")

# RPM_GPG_KEYS_A6 contains keys we only install for the A6 repo.
# DATADOG_RPM_KEY.public is only useful to install old (< 6.14) Agent packages.
RPM_GPG_KEYS_A6=("DATADOG_RPM_KEY.public")

# Set up a named pipe for logging
npipe=/tmp/$$.tmp
mknod $npipe p

# Log all output to a log for error checking
tee <$npipe $logfile &
exec 1>&-
exec 1>$npipe 2>&1
trap "rm -f $npipe" EXIT

# Colours
readonly C_NOC="\\033[0m"    # No colour
readonly C_RED="\\033[0;31m" # Red
readonly C_GRN="\\033[0;32m" # Green
readonly C_BLU="\\033[0;34m" # Blue
readonly C_PUR="\\033[0;35m" # Purple
readonly C_CYA="\\033[0;36m" # Cyan
readonly C_WHI="\\033[0;37m" # White

### Helper functions
print_red () { local i; for i in "$@"; do echo -e "${C_RED}${i}${C_NOC}"; done }
print_grn () { local i; for i in "$@"; do echo -e "${C_GRN}${i}${C_NOC}"; done }
print_blu () { local i; for i in "$@"; do echo -e "${C_BLU}${i}${C_NOC}"; done }
print_pur () { local i; for i in "$@"; do echo -e "${C_PUR}${i}${C_NOC}"; done }
print_cya () { local i; for i in "$@"; do echo -e "${C_CYA}${i}${C_NOC}"; done }
print_whi () { local i; for i in "$@"; do echo -e "${C_WHI}${i}${C_NOC}"; done }

function on_error() {
    print_red "$ERROR_MESSAGE
It looks like you hit an issue when trying to install the StackState Agent v2.

Basic information about the Agent are available at:

    https://l.stackstate.com/agent-install-docs-link

If you're still having problems, please send an email to info@stackstate.com
with the contents of $logfile and we'll do our very best to help you
solve your problem.\n"
}
trap on_error ERR

if [ -n "$STS_API_KEY" ]; then
    api_key=$STS_API_KEY
fi

if [ -n "$STS_SITE" ]; then
    site="$STS_SITE"
fi

sts_url="http://localhost/stsAgent"
if [ -n "$STS_URL" ]; then
    sts_url=$STS_URL
fi

no_start=
if [ -n "$STS_INSTALL_ONLY" ]; then
    no_start=true
fi

no_repo=
if [ ! -z "$STS_INSTALL_NO_REPO" ]; then
    no_repo=true
fi

if [ -n "$STS_HOSTNAME" ]; then
    hostname=$STS_HOSTNAME
fi

# comma-separated list of tags
default_host_tags="os:linux"
if [ -n "$HOST_TAGS" ]; then
    host_tags="$default_host_tags,$HOST_TAGS"
else
    host_tags=$default_host_tags
fi

if [ -n "$CODE_NAME" ]; then
    code_name=$CODE_NAME
else
    code_name="stable"
fi

if [ -z "$DEBIAN_REPO" ]; then
    # for offline script remember default production repo address
    DEBIAN_REPO="https://stackstate-agent-3.s3.amazonaws.com"
fi

if [ -z "$YUM_REPO" ]; then
    # for offline script remember default production repo address
    YUM_REPO="https://stackstate-agent-3-rpm.s3.amazonaws.com"
fi

if [ -n "$SKIP_SSL_VALIDATION" ]; then
    skip_ssl_validation=$SKIP_SSL_VALIDATION
fi

if [ ! $api_key ]; then
    print_red "API key not available in STS_API_KEY environment variable.\n"
    exit 1
fi

keyserver="hkp://keyserver.ubuntu.com:80"
backup_keyserver="hkp://pool.sks-keyservers.net:80"
# use this env var to specify another key server, such as
# hkp://p80.pool.sks-keyservers.net:80 for example.
if [ -n "$DD_KEYSERVER" ]; then
  keyserver="$DD_KEYSERVER"
fi

INSTALL_MODE="REPO"
if [ ! -z "$1" ]; then
    if test -f "$1"; then
        print_blu "Trying to install from local package $1"
        INSTALL_MODE="FILE"
        LOCAL_PKG_NAME="$1"
    fi
fi

# OS/Distro Detection
# Try lsb_release, fallback with /etc/issue then uname command
KNOWN_DISTRIBUTION="(Debian|Ubuntu|RedHat|CentOS|openSUSE|Amazon|Arista|SUSE)"
DISTRIBUTION=$(lsb_release -d 2>/dev/null | grep -Eo $KNOWN_DISTRIBUTION || grep -Eo $KNOWN_DISTRIBUTION /etc/issue 2>/dev/null || grep -Eo $KNOWN_DISTRIBUTION /etc/Eos-release 2>/dev/null || grep -m1 -Eo $KNOWN_DISTRIBUTION /etc/os-release 2>/dev/null || uname -s)

if [ $DISTRIBUTION = "Darwin" ]; then
    print_red "This script does not support installing on the Mac.

Please use the 1-step script available at https://app.datadoghq.com/account/settings#agent/mac."
    exit 1

elif [ -f /etc/debian_version -o "$DISTRIBUTION" == "Debian" -o "$DISTRIBUTION" == "Ubuntu" ]; then
    OS="Debian"
elif [ -f /etc/redhat-release -o "$DISTRIBUTION" == "RedHat" -o "$DISTRIBUTION" == "CentOS" -o "$DISTRIBUTION" == "Amazon" ]; then
    OS="RedHat"
# Some newer distros like Amazon may not have a redhat-release file
elif [ -f /etc/system-release -o "$DISTRIBUTION" == "Amazon" ]; then
    OS="RedHat"
# Arista is based off of Fedora14/18 but do not have /etc/redhat-release
elif [ -f /etc/Eos-release -o "$DISTRIBUTION" == "Arista" ]; then
    OS="RedHat"
# openSUSE and SUSE use /etc/SuSE-release or /etc/os-release
elif [ -f /etc/SuSE-release -o "$DISTRIBUTION" == "SUSE" -o "$DISTRIBUTION" == "openSUSE" ]; then
    OS="SUSE"
fi

# Root user detection
if [ $(echo "$UID") = "0" ]; then
    sudo_cmd=''
else
    sudo_cmd='sudo'
fi

# Install the necessary package sources
if [ $OS = "RedHat" ]; then
    if [ -z "$no_repo" ]; then
    print_blu "* Installing YUM sources for StackState\n"
    $sudo_cmd sh -c "echo -e '[stackstate]\nname = StackState\nbaseurl = $YUM_REPO/$code_name/\nenabled=1\ngpgcheck=1\npriority=1\ngpgkey=$YUM_REPO/public.key' > /etc/yum.repos.d/stackstate.repo"
    fi

    gpgkeys=''
    separator='\n       '
    for key_path in "${RPM_GPG_KEYS[@]}"; do
      gpgkeys="${gpgkeys:+"${gpgkeys}${separator}"}https://${yum_url}/${key_path}"
    done
    if [ "$agent_major_version" -eq 6 ]; then
      for key_path in "${RPM_GPG_KEYS_A6[@]}"; do
        gpgkeys="${gpgkeys:+"${gpgkeys}${separator}"}https://${yum_url}/${key_path}"
      done
    fi

    $sudo_cmd sh -c "echo -e '[datadog]\nname = Datadog, Inc.\nbaseurl = https://${yum_url}/${yum_version_path}/${ARCHI}/\nenabled=1\ngpgcheck=1\nrepo_gpgcheck=0\npriority=1\ngpgkey=${gpgkeys}' > /etc/yum.repos.d/datadog.repo"

    printf "\033[34m* Installing the Datadog Agent package\n\033[0m\n"
    $sudo_cmd yum -y clean metadata
    if [ $INSTALL_MODE = "REPO" ]; then
        $sudo_cmd yum -y --disablerepo='*' --enablerepo='stackstate' install $PKG_NAME || $sudo_cmd yum -y install $PKG_NAME
    else
        $sudo_cmd yum -y localinstall $LOCAL_PKG_NAME
    fi

elif [ $OS = "Debian" ]; then
    print_blu "* Installing apt-transport-https\n"
    $sudo_cmd apt-get update || print_red "'apt-get update' failed, the script will not install the latest version of apt-transport-https."
    $sudo_cmd apt-get install -y apt-transport-https || print_red "> 'apt-transport-https' was not installed"
    # Only install dirmngr if it's available in the cache
    # it may not be available on Ubuntu <= 14.04 but it's not required there
    cache_output=$(apt-cache search dirmngr)
    if [ ! -z "$cache_output" ]; then
        $sudo_cmd apt-get install -y dirmngr
    fi

    print_blu "* Configuring APT package sources for StackState\n"
    $sudo_cmd sh -c "echo 'deb $DEBIAN_REPO $code_name main' > /etc/apt/sources.list.d/stackstate.list"
    if [[ $INSTALL_MODE == "REPO" ]]; then
        $sudo_cmd apt-key adv --recv-keys --keyserver hkp://keyserver.ubuntu.com:80 B3CC4376
    else
        $sudo_cmd apt-key adv --recv-keys --keyserver hkp://keyserver.ubuntu.com:80 B3CC4376 || print_red "> Failed to install apt repo key (no internet connection?). Please install separately for further repo updates"
    fi
    printf "\033[34m\n* Installing APT package sources for Datadog\n\033[0m\n"
    $sudo_cmd sh -c "echo 'deb https://${apt_url}/ ${apt_repo_version}' > /etc/apt/sources.list.d/datadog.list"

    for key in "${APT_GPG_KEYS[@]}"; do
      for retries in {0..4}; do
        $sudo_cmd apt-key adv --recv-keys --keyserver "${keyserver}" "${key}" && break
        if [ "$retries" -eq 4 ]; then
          ERROR_MESSAGE="ERROR
  Couldn't fetch Datadog public key ${key}.
  This might be due to a connectivity error with the key server
  or a temporary service interruption.
  *****
  "
          false
        fi
        printf "\033[33m\napt-key failed to retrieve Datadog's public key ${key}, retrying in 5 seconds...\n\033[0m\n"
        sleep 5
        if [ "$retries" -eq 1 ]; then
          printf "\033[34mSwitching to backup keyserver\n\033[0m\n"
          keyserver="${backup_keyserver}"
        fi
      done
    done

    print_blu "* Installing the StackState Agent v2 package\n"
    ERROR_MESSAGE="ERROR
Failed to update the sources after adding the StackState repository.
This may be due to any of the configured APT sources failing -
see the logs above to determine the cause.
If the failing repository is StackState, please contact StackState support.
*****
"

  # Try to guess if we're installing on SUSE 11, as it needs a different flow to work
  if cat /etc/SuSE-release 2>/dev/null | grep VERSION | grep 11; then
    SUSE11="yes"
  fi

  echo -e "\033[34m\n* Importing the Datadog GPG Keys\n\033[0m"
  if [ "$SUSE11" == "yes" ]; then
    # SUSE 11 special case
    for key_path in "${RPM_GPG_KEYS[@]}"; do
      $sudo_cmd curl -o "/tmp/${key_path}" "https://${yum_url}/${key_path}"
      $sudo_cmd rpm --import "/tmp/${key_path}"
    done
    if [ "$agent_major_version" -eq 6 ]; then
      for key_path in "${RPM_GPG_KEYS_A6[@]}"; do
        $sudo_cmd curl -o "/tmp/${key_path}" "https://${yum_url}/${key_path}"
        $sudo_cmd rpm --import "/tmp/${key_path}"
      done
    fi
  else
    for key_path in "${RPM_GPG_KEYS[@]}"; do
      $sudo_cmd rpm --import "https://${yum_url}/${key_path}"
    done
    if [ "$agent_major_version" -eq 6 ]; then
      for key_path in "${RPM_GPG_KEYS_A6[@]}"; do
        $sudo_cmd rpm --import "https://${yum_url}/${key_path}"
      done
    fi
  fi

  gpgkeys=''
  separator='\n       '
  for key_path in "${RPM_GPG_KEYS[@]}"; do
    gpgkeys="${gpgkeys:+"${gpgkeys}${separator}"}https://${yum_url}/${key_path}"
  done
  if [ "$agent_major_version" -eq 6 ]; then
    for key_path in "${RPM_GPG_KEYS_A6[@]}"; do
      gpgkeys="${gpgkeys:+"${gpgkeys}${separator}"}https://${yum_url}/${key_path}"
    done
  fi

  echo -e "\033[34m\n* Installing YUM Repository for Datadog\n\033[0m"
  $sudo_cmd sh -c "echo -e '[datadog]\nname=datadog\nenabled=1\nbaseurl=https://${yum_url}/suse/${yum_version_path}/${ARCHI}\ntype=rpm-md\ngpgcheck=1\nrepo_gpgcheck=0\ngpgkey=${gpgkeys}' > /etc/zypp/repos.d/datadog.repo"

  echo -e "\033[34m\n* Refreshing repositories\n\033[0m"
  $sudo_cmd zypper --non-interactive --no-gpg-checks refresh datadog

  echo -e "\033[34m\n* Installing Datadog Agent\n\033[0m"
  $sudo_cmd zypper --non-interactive install "$agent_flavor"

else
    print_red "Your OS or distribution is not supported yet.\n"
    exit 1
fi

# Set the configuration
if [ ! -e $CONF ]; then
    $sudo_cmd cp $CONF.example $CONF
fi
if [ $api_key ]; then
    print_blu "* Adding your API key to the Agent configuration: $CONF\n"
    $sudo_cmd sh -c "sed -i 's/api_key:.*/api_key: $api_key/' $CONF"
fi
if [ $sts_url ]; then
    sts_url_esc=$(sed 's/[/.&]/\\&/g' <<<"$sts_url")
    print_blu "* Adding StackState url to the Agent configuration: $CONF\n"
    $sudo_cmd sh -c "sed -i 's/sts_url:.*/sts_url: $sts_url_esc/' $CONF"
fi
if [ $hostname ]; then
    print_blu "* Adding your STS_HOSTNAME to the Agent configuration: $CONF\n"
    $sudo_cmd sh -c "sed -i 's/# hostname:.*/hostname: $hostname/' $CONF"
fi
if [ $host_tags ]; then
    print_blu "* Adding your HOST TAGS to the Agent configuration: $CONF\n"
    formatted_host_tags="['"$(echo "$host_tags" | sed "s/,/','/g")"']" # format `env:prod,foo:bar` to yaml-compliant `['env:prod','foo:bar']`
    $sudo_cmd sh -c "sed -i \"s/# tags:.*/tags: "$formatted_host_tags"/\" $CONF"
fi
if [ $skip_ssl_validation ]; then
    print_blu "* Skipping SSL validation in the Agent configuration: $CONF\n"
    $sudo_cmd sh -c "sed -i 's/# skip_ssl_validation:.*/skip_ssl_validation: $skip_ssl_validation/' $CONF"
fi

function version_gt() {
    test "$(printf '%s\n' "$@" | sort -V | head -n 1)" != "$1"
}

#Minimum kernel version required for network tracer https://github.com/StackVista/tcptracer-bpf/blob/master/pkg/tracer/common/common_linux.go#L28
min_required_kernel="4.3.0"
current_kernel=$(uname -r)
if version_gt $min_required_kernel $current_kernel; then
    print_cya "* The network tracer does not support your kernel version (min required $min_required_kernel), disabling it\n"
    $sudo_cmd sh -c "sed -i \"s/network_tracing_enabled:.*/network_tracing_enabled: 'false'/\" $CONF"
fi

$sudo_cmd chown $PKG_USER:$PKG_USER $CONF
$sudo_cmd chmod 640 $CONF

# Use systemd by default
restart_cmd="$sudo_cmd systemctl restart $PKG_NAME.service"
stop_instructions="$sudo_cmd systemctl stop $PKG_NAME"
start_instructions="$sudo_cmd systemctl start $PKG_NAME"

# Try to detect Upstart, this works most of the times but still a best effort
if /sbin/init --version 2>&1 | grep -q upstart; then
    restart_cmd="$sudo_cmd start $PKG_NAME"
    stop_instructions="$sudo_cmd stop $PKG_NAME"
    start_instructions="$sudo_cmd start $PKG_NAME"
elif [[ -d /etc/rc.d/ || -d /etc/init.d/ ]]; then
    # Use sysv-init
    restart_cmd="$sudo_cmd service $PKG_NAME restart"
    stop_instructions="$sudo_cmd service $PKG_NAME stop"
    start_instructions="$sudo_cmd service $PKG_NAME start"
fi

if [ $no_start ]; then
    print_blu "
* STS_INSTALL_ONLY environment variable set: the newly installed version of the agent
will not be started. You will have to do it manually using the following
command:

    $restart_cmd
\n"
    exit
fi

print_blu "* Starting the Agent...\n"
eval $restart_cmd

# Metrics are submitted, echo some instructions and exit
print_grn "
Your Agent is running and functioning properly. It will continue to run in the
background and submit metrics to StackState.

If you ever want to stop the Agent, run:

    $stop_instructions

And to run it again run:

    $start_instructions
\n"
