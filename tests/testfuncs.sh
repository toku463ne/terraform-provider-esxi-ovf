#!/bin/bash

source /var/tmp/esxi-ovf-test_vars.sh
statusfile=/tmp/tfteststatus

function runcmd {
	cmd=$@
	echo $cmd
	$cmd
}

function cleanup {
	rm -f terraform.tfstate*
	rm -rf tfwork
}

function terraform_plan {
	runcmd terraform plan
}

function terraform_cmd {
	echo 0 > $statusfile
	expect -c "
set timeout 300
spawn terraform $@
expect \"Enter a value:\"
send \"yes\n\"
expect \"$PROMPT\"
" | while read line
	do
		echo $line
		if [[ "$line" =~ .*"Error:".* ]]
		then
			echo 1 > $statusfile
		fi
	done
	status=`cat $statusfile`
	rm -f $statusfile
	return $status
}


