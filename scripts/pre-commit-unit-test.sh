#!/bin/sh

echo "Running unit tests..."

./run.sh testall

if [ $? -ne 0 ]; then
	echo "You have an error with unit tests. Run ./run.sh testall"
	exit 1
fi
