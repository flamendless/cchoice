#!/usr/bin/env bash

set -euf -o pipefail

usage() {
	cat <<'USAGE'
Usage: SSH_ADDR=user@host ./scripts/sshdeploy.sh --dev|--prod

Environment variables:
  SSH_ADDR    Remote SSH address (required), e.g. user@example.com

Flags:
  --dev, -d   Deploy to ~/cchoice-dev (mage dev)
  --prod, -p  Deploy to ~/cchoice (mage prod)
USAGE
}

SSH_ADDR="${SSH_ADDR:-}"
if [ -z "$SSH_ADDR" ]; then
	echo "Error: SSH_ADDR environment variable is required." >&2
	usage >&2
	exit 1
fi

ENV=""
while [ $# -gt 0 ]; do
	case "$1" in
	--dev | -d)
		ENV="dev"
		;;
	--prod | -p)
		ENV="prod"
		;;
	-h | --help)
		usage
		exit 0
		;;
	*)
		echo "Error: unknown argument '$1'." >&2
		usage >&2
		exit 1
		;;
	esac
	shift
done

if [ -z "$ENV" ]; then
	echo "Error: --dev or --prod flag is required." >&2
	usage >&2
	exit 1
fi

read -rsp "SSH password: " SSH_PASSWORD
echo

ssh_cmd() {
	if command -v sshpass >/dev/null 2>&1; then
		SSHPASS="$SSH_PASSWORD" sshpass -e ssh "$@"
	else
		echo "sshpass not found; ssh will prompt for password." >&2
		ssh "$@"
	fi
}

echo "Deploying $ENV to $SSH_ADDR..."

ssh_cmd "$SSH_ADDR" bash --login -s "$ENV" <<'EOF'
set -euf -o pipefail

ENV="$1"

if [ "$ENV" = "dev" ]; then
	PROJECT_DIR=~/cchoice-dev
	MAGE_TARGET="dev"
	PROCESS_PATTERN="./tmp/cchoicedev api"
else
	PROJECT_DIR=~/cchoice
	MAGE_TARGET="prod"
	PROCESS_PATTERN="./tmp/cchoiceprod api"
fi

echo "Navigating to project directory..."
cd "$PROJECT_DIR"

echo "Syncing repository..."
if ! git pull origin main; then
	echo "Pull failed, stashing local changes and retrying..."
	git stash push -u -m "deploy stash $(date -u +%Y-%m-%dT%H:%M:%SZ)"
	git pull origin main
fi

echo "Fetching tags..."
git fetch --prune --tags origin

echo "Stopping existing process..."
if pgrep -af "$PROCESS_PATTERN" >/dev/null 2>&1; then
	pkill -f "$PROCESS_PATTERN" || true
	sleep 1
else
	echo "No existing process found."
fi

echo "Building..."
BUILD_OUTPUT=$(mage "$MAGE_TARGET" 2>&1)
echo "$BUILD_OUTPUT"

RUN_CMD=$(echo "$BUILD_OUTPUT" | grep '^Run: ' | sed 's/^Run: //' | sed 's/ > out.*//' | sed 's/ &//' | xargs)
if [ -z "$RUN_CMD" ]; then
	echo "Error: could not parse run command from mage $MAGE_TARGET output." >&2
	exit 1
fi

echo "Starting API with: $RUN_CMD"
eval "$RUN_CMD > out 2>&1 &"

echo "Done!"
EOF
