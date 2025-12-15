#!/bin/bash

set -euo pipefail

ACCOUNT_ID="${CLOUDFLARE_ACCOUNT_ID:-}"
API_TOKEN="${CLOUDFLARE_API_TOKEN:-}"
PATTERN="${1:-favicons}"
PER_PAGE=1000

if [[ -z "$ACCOUNT_ID" ]]; then
    echo "Error: CLOUDFLARE_ACCOUNT_ID environment variable is not set"
    exit 1
fi

if [[ -z "$API_TOKEN" ]]; then
    echo "Error: CLOUDFLARE_API_TOKEN environment variable is not set"
    exit 1
fi

if ! command -v jq &> /dev/null; then
    echo "Error: jq is required but not installed"
    exit 1
fi

echo "Cloudflare Images List Script"
echo "=================================="
echo "Account ID: ${ACCOUNT_ID:0:8}..."
echo "Pattern: $PATTERN"
echo ""

BASE_URL="https://api.cloudflare.com/client/v4/accounts/$ACCOUNT_ID/images/v1"

fetch_images() {
    local continuation_token="$1"
    local url="$BASE_URL?per_page=$PER_PAGE"

    if [[ -n "$continuation_token" ]]; then
        url="$url&continuation_token=$continuation_token"
    fi

    curl -s -X GET "$url" \
        -H "Authorization: Bearer $API_TOKEN" \
        -H "Content-Type: application/json"
}

total_found=0
continuation_token=""

while true; do
    echo "Fetching images..."
    response=$(fetch_images "$continuation_token")

    success=$(echo "$response" | jq -r '.success')
    if [[ "$success" != "true" ]]; then
        echo "Error fetching images:"
        echo "$response" | jq '.errors'
        exit 1
    fi

    matching_ids=$(echo "$response" | jq -r --arg pattern "$PATTERN" \
        '.result.images[] | select(.id | test($pattern)) | .id')

    echo "$matching_ids"

    if [[ -n "$matching_ids" ]]; then
        while IFS= read -r image_id; do
            ((total_found++))
        done <<< "$matching_ids"
    fi

    continuation_token=$(echo "$response" | jq -r '.result.continuation_token // empty')
    if [[ -z "$continuation_token" ]]; then
        break
    fi

    echo "Fetching next page..."
done

echo ""
echo "=================================="
echo "Summary:"
echo "  Total images matching pattern: $total_found"
