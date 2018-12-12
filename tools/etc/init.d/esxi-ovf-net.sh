#!/bin/bash

set -e
set -x

VMTOOLSD="/usr/bin/vmtoolsd"

function getinfo() {
    $VMTOOLSD --cmd "info-get guestinfo.net.$1" 2>/dev/null
}

function resolve_distro() {
distro_file=/etc/issue
if [[ -e /etc/redhat-release ]]
then
	distro_file=/etc/redhat-release
fi

distro=`head -n 1 $distro_file | grep -o -E -i "redhad|centos|fedora|debian|ubuntu"`
distro=`echo $distro | tr '[A-Z]' '[a-z]'` # convert to lower case

case "$distro" in
redhat)
	distro=redhat;;
centos)
	distro=redhat;;
fedora)
	distro=redhat;;
debian)
	distro=debian;;
ubuntu)
	distro=debian;;
*)
	echo "$0: Unsupported distro $distro" >&2
	exit 1;;
esac

if [[ $# -gt 0 ]]
then
	for supported_distro in "$@" # iterate over each argument
	do
		if [[ "$distro" == "$supported_distro" ]]
		then
			echo $distro
			exit 0
		fi
	done

	echo "$0: Unsupported distro $distro" >&2
	exit 1
else
	echo $distro
	exit 0
fi
}

#### main ###

if [ ! -x $VMTOOLSD ]
then
    echo "$VMTOOLSD does not exist or not executable"
    exit 1 
fi

dev=`getinfo dev`
ipaddr=`getinfo address`
netmask=`getinfo netmask`
gateway=`getinfo gateway`
set +e
nameservers=`getinfo nameservers`
set -e

distro=`resolve_distro`

if [ "$distro" == "debian" ]
then
netfile=/etc/network/interfaces
set +e
grep $ipaddr $netfile
if [ "$?" -eq 0 ]
then
	exit 0
fi
set -e

d=`date`
mv $netfile $netfile.bk
cat > $netfile << EOF
# Created by /esxi-ovf-net.sh
# $d

source /etc/network/interfaces.d/*

auto lo
iface lo inet loopback

auto $dev
iface $dev inet static
    address $ipaddr
    netmask $netmask
    gateway $gateway
EOF
set +e
if [ "$nameservers" != "" ]
then
	echo "    dns-nameservers $nameservers" >> $netfile
fi

ifdown $dev; ip addr flush dev $dev; ifup $dev
set -e
fi

if [ "$distro" == "redhat" ]
then
set +e
ifdown $dev
netfile=/etc/sysconfig/network-scripts/ifcfg-$dev
grep $ipaddr $netfile
if [ "$?" -eq 0 ]
then
	exit 0
fi
set -e
mv $netfile $netfile.bk
cat > $netfile << EOF
# Created by /esxi-ovf-net.sh
# $d
DEVICE="$dev"
BOOTPROTO="static"
GATEWAY="$gateway"
IPADDR="$ipaddr"
NETMASK="$netmask"
ONBOOT="yes"
EOF
ifup $dev
if [ "$nameservers" != "" ]
then
	echo "    nameservers $nameservers" > /etc/resolv.conf
fi
fi
