#!/bin/bash

target=$1
cmdname=$2

p=""
matched=0
is_cmd_line=0
cat $target | while read line
do
    p=`echo $line | cut -c1-1`
    if [ "$p" == "%" -a "$matched" -eq 0 ];then
        cmdname2=`echo $line | cut -d"%" -f2`
        if [ "$cmdname" == "$cmdname2" ];then
            matched=1
            is_cmd_line=1
        fi
    fi
    if [ "$matched" -eq 1 -a "$is_cmd_line" -eq 0 -a "$p" == "%" ];then
        exit 0
    fi
    if [ "$matched" -eq 1 ];then
        echo $line
    fi
    p=""
    is_cmd_line=0
done