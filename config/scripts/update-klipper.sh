#!/bin/bash

# Script to download the latest release of a specific type from GitHub
# Usage: ./download_release.sh [nightly|testing|stable]
#
# This script will:
# 1. Check GitHub releases for the latest release of the specified type
# 2. Download the release ZIP file to the current directory
#
# Release types:
# - nightly: Daily builds with the latest changes
# - testing: Pre-release builds for testing new features
# - stable: Production-ready releases

# Check if release type is provided
if [ $# -eq 0 ]; then
    echo "Error: Release type not provided"
    echo "Usage: $0 [nightly|testing|stable]"
    exit 1
fi

RELEASE_TYPE=$1

# Validate release type
if [[ "$RELEASE_TYPE" != "nightly" && "$RELEASE_TYPE" != "testing" && "$RELEASE_TYPE" != "stable" ]]; then
    echo "Error: Invalid release type. Must be one of: nightly, testing, stable"
    exit 1
fi

echo "Fetching latest $RELEASE_TYPE release..."

# GitHub API URL for releases
API_URL="https://api.github.com/repos/pijalu/artillery-m1-klipper/releases"

# Get the latest release of the specified type
# Using awk to parse JSON response instead of jq
RESPONSE=$(curl -s "$API_URL")

# Extract tag name for our release type
TAG_NAME=$(echo "$RESPONSE" | awk -v type="$RELEASE_TYPE" '
    /"tag_name":/ && index($0, "\"" type "-") {
        # Extract the tag name
        gsub(/[[:space:]]*"tag_name":[[:space:]]*"/, "")
        gsub(/".*/, "")
        print
        exit
    }
')

if [ -z "$TAG_NAME" ]; then
    echo "Error: No $RELEASE_TYPE release found"
    exit 1
fi

# Extract download URL for this tag
DOWNLOAD_URL=$(echo "$RESPONSE" | awk -v tag="$TAG_NAME" '
    # State: looking for our tag
    /"tag_name":/ && index($0, "\"" tag "\"") {
        found_tag = 1
        next
    }
    
    # If we found our tag, look for assets
    found_tag && /"browser_download_url":/ {
        # Extract the URL
        match($0, /https:\/\/[^"]+/)
        if (RSTART > 0) {
            print substr($0, RSTART, RLENGTH)
            exit
        }
    }
')

# Extract filename for this tag
FILENAME=$(echo "$RESPONSE" | awk -v tag="$TAG_NAME" '
    # State: looking for our tag
    /"tag_name":/ && index($0, "\"" tag "\"") {
        found_tag = 1
        next
    }
    
    # If we found our tag, look for the zip filename
    found_tag && /"name":/ && /\.zip/ {
        # Extract the filename
        match($0, /[^"]+\.zip/)
        if (RSTART > 0) {
            print substr($0, RSTART, RLENGTH)
            exit
        }
    }
')

if [ -z "$DOWNLOAD_URL" ] || [ -z "$FILENAME" ]; then
    echo "Error: Could not extract download URL or filename"
    exit 1
fi

echo "Downloading $FILENAME..."
curl -L -o "$FILENAME" "$DOWNLOAD_URL"

if [ $? -eq 0 ]; then
    echo "Download complete: $FILENAME"
else
    echo "Error: Download failed"
    exit 1
fi