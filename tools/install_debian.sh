#!/bin/bash

echo "installing net configuration tools"

cp etc/init.d/esxi-ovf-net.sh /etc/init.d/
ln -s /etc/init.d/esxi-ovf-net.sh /etc/rc2.d/S10esxi-ovf-net

