#!/bin/sh

echo "Running commit msg tests..."

mage checkcommitprefix
if [ $? -ne 0 ]; then
	echo "You have an error with unit tests. Run `mage checkcommitprefix`"
	exit 1
fi
