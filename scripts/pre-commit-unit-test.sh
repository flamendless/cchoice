#!/bin/sh

echo "Running unit tests..."

mage testAll
mage testInteg

if [ $? -ne 0 ]; then
	echo "You have an error with unit tests. Run `mage testAll` and `mage testInteg`"
	exit 1
fi
