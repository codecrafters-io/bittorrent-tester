#!/bin/sh
string=$2           # 5:hello
string=${string#*:} # hello
echo "\"$string\""
exit 0
