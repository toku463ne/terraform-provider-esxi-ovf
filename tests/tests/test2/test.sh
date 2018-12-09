#!/bin/bash

set -e

source ../../testfuncs.sh

echo $TF_VAR_hostip1

for f in `ls ../../ovf`
do
	echo ""
	echo ""
	echo "======================="
	export TF_VAR_ovfpath="../../ovf/$f/$f.ovf"
	ls $TF_VAR_ovfpath
	cleanup
	terraform_plan
	terraform_cmd apply
	terraform_plan
	terraform_cmd destroy
	exit
done
