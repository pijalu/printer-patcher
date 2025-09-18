#!/bin/bash

# Release types:
# - nightly: Daily builds with the latest changes
# - testing: Pre-release builds for testing new features
# - stable: Production-ready releases
RELEASE_TYPE=testing

# Validate release type
if [[ "$RELEASE_TYPE" != "nightly" && "$RELEASE_TYPE" != "testing" && "$RELEASE_TYPE" != "stable" ]]; then
    echo "Error: Invalid release type. Must be one of: nightly, testing, stable"
    exit $LINENO
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
    exit $LINENO
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
    exit $LINENO
fi

# Check if the filename (releasetype-timestamp.zip) is different that the /home/mks/printer_data/config/.version (contains the release-timestamp)
LOCAL_VERSION_FILE="/home/mks/printer_data/config/.version"
REMOTE_VERSION=$(echo "$FILENAME" | sed 's/\.zip$//' | sed 's/release-//')

# Check if local version file exists and read it
if [ -f "$LOCAL_VERSION_FILE" ]; then
    LOCAL_VERSION=$(cat "$LOCAL_VERSION_FILE" | sed 's/release-//g')
    echo "Local version: $LOCAL_VERSION"
    echo "Remote version: $REMOTE_VERSION"
    
    # Compare versions
    if [ "$LOCAL_VERSION" = "$REMOTE_VERSION" ]; then
        echo "Klipper is already up to date ($LOCAL_VERSION)"
        echo "[OK] - Klipper update" 
        exit 0
    else
        echo "Updating Klipper from $LOCAL_VERSION to $REMOTE_VERSION"
    fi
else
    echo "No local version found, installing $REMOTE_VERSION"
fi

echo "Downloading $FILENAME..."
curl -L -o "/tmp/$FILENAME" "$DOWNLOAD_URL"

if [ $? -eq 0 ]; then
    echo "Download complete: $FILENAME"
else
    echo "Error: Download failed"
    exit $LINENO
fi

# Extract the zip in TMP, deploy files, and cleanup
echo "Extracting $FILENAME..."
mkdir -p "/tmp/klipper-update" || exit $LINENO
unzip -o "/tmp/$FILENAME" -d "/tmp/klipper-update" || exit $LINENO

CHANGES_MADE=0

# Deploy files manually, only copying files that are different
# Files are in the config/ subdirectory of the release zip
CONFIG_DIR="/tmp/klipper-update/release/config"
TARGET_DIR="/home/mks/printer_data/config"
BACKUP_SUFFIX="pre-$REMOTE_VERSION-$(date +%s)"

echo "Deploying files from $CONFIG_DIR to $TARGET_DIR..."

# Check if config directory exists in the package
if [ ! -d "$CONFIG_DIR" ]; then
    echo "Error: Config directory not found in package"
    exit $LINENO
fi

# Copy files that are different, backing up originals
ARTDO="sudo -u artillery"
find "$CONFIG_DIR" -type f | while read FILE; do
    # Get relative path from config directory
    REL_PATH="${FILE#$CONFIG_DIR/}"
    TARGET_FILE="$TARGET_DIR/$REL_PATH"
    
    # Create target directory if it doesn't exist
    TARGET_FILE_DIR=$(dirname "$TARGET_FILE")
    mkdir -p "$TARGET_FILE_DIR" || exit $LINENO
    
    # Check if file exists in target and is different
    if [ -f "$TARGET_FILE" ]; then
        if ! cmp -s "$FILE" "$TARGET_FILE"; then
            echo "Updating $REL_PATH"
            # Backup original file with timestamp
            $ARTDO cp "$TARGET_FILE" "${TARGET_FILE}-$BACKUP_SUFFIX" || exit $LINENO
            # Copy new file
            $ARTDO cp "$FILE" "$TARGET_FILE" || exit $LINENO
            CHANGES_MADE=1
        fi
    else
        echo "Adding new file $REL_PATH"
        # Copy new file
        $ARTDO cp "$FILE" "$TARGET_FILE" || exit $LINENO
        CHANGES_MADE=1
    fi
done

# Restart services if changes were made
if [ "$CHANGES_MADE" -eq 1 ]; then
    echo "Restarting Klipper and Moonraker services..."
    sudo systemctl restart klipper || exit $LINENO
    sudo systemctl restart moonraker || exit $LINENO
    echo "Services restarted successfully"
else
    echo "No changes made, services not restarted"
fi

# Cleanup downloaded files and extracted content
echo "Cleaning up temporary files..."
rm -f "/tmp/$FILENAME" || exit $LINENO
rm -rf "/tmp/klipper-update" || exit $LINENO

if [ "$CHANGES_MADE" -eq 1 ]; then
    echo "Klipper update completed and services restarted successfully"
    echo "Original files backed up with suffix: $BACKUP_SUFFIX"
else
    echo "Klipper is up to date, no changes made"
fi

echo "[OK] - Klipper update" 
