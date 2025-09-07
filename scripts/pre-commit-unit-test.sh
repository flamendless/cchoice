#!/bin/sh

echo "Running unit tests..."

mage testAll

if [ $? -ne 0 ]; then
	echo "You have an error with unit tests. Run 'mage testAll'"
	exit 1
fi
