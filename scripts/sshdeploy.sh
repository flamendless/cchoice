#!/usr/bin/env bash

set -euf -o pipefail

SSH_ADDR="${SSH_ADDR:-}"
if [ -z "$SSH_ADDR" ]; then
	echo "SSH_ADDR is not set."
	exit 1
fi

ssh "$SSH_ADDR" bash --login -i -s <<EOF
	set -euf -o pipefail

	echo "Navigating to project directory..."
	cd "cchoice"

	echo "Syncing repository..."
	git pull origin main

	echo "Stopping existing process..."
	pkill -f "./tmp/main api" || echo "No existing process found."

	echo "Building..."
	mage prod

	echo "Running API..."
	./tmp/main api > out 2>&1 &

	echo "Done!"
EOF
