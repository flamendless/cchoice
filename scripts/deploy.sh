#!/usr/bin/env bash

set -euf -o pipefail

echo "Navigating to project directory..."
cd "cchoice"

echo "Syncing repository..."
git pull origin main

echo "Fetching tags..."
git fetch --prune --tags origin

echo "Stopping existing process..."
pkill -x "./tmp/main api" || echo "No existing process found."

echo "Running migrations..."
mage dbUp

echo "Building..."
go generate ./...
mage prod

echo "Running API..."
./tmp/main api > out 2>&1 &

echo "Done!"
