#!/bin/bash

set -e

source testfuncs.sh
runinstall=no

echo "Tests have done on ubuntu18"


basedir=`pwd`
cd ../
if [ "$runinstall" == "yes" ]
then
	sudo apt-get install expect -y
	./install.sh
fi

cd $basedir
for f in `ls tests`
do
	cd tests/$f
	cleanup
	runcmd terraform init
	./test.sh
	cd $basedir
done
