#!/bin/bash

set -e

function cmd() {
    echo $1
    $1
}

which ovftool
if [ "$?" != 0 ]
then
  echo "Need to install OVF tools."
  echo "OVF tool is not installed."
  exit 1
fi


echo "Getting go modules"
echo "------------------"
cmd "go get github.com/hashicorp/terraform/helper/schema"
cmd "go get github.com/jamesharr/expect"
cmd "go get github.com/mattn/go-sqlite3"
cmd "go get github.com/pkg/errors"
cmd "go get gopkg.in/ini.v1"

echo ""
echo "Will compile and install"
echo "------------------------"
echo go build -o terraform-provider-esxi-ovf
go build -o terraform-provider-esxi-ovf
echo mkdir -p ~/.terraform.d/plugins/linux_amd64
mkdir -p ~/.terraform.d/plugins/linux_amd64
echo mv terraform-provider-esxi-ovf ~/.terraform.d/plugins/linux_amd64
mv terraform-provider-esxi-ovf ~/.terraform.d/plugins/linux_amd64

echo ""
echo "Finished installation."


