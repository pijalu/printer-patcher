#!/bin/bash

# Release types:
# - nightly: Daily builds with the latest changes
# - testing: Pre-release builds for testing new features
# - stable: Production-ready releases
RELEASE_TYPE=stable
ARTDO="sudo -u artillery"

echo "Fetching latest $RELEASE_TYPE release..."

# GitHub API URL for releases
API_URL="https://api.github.com/repos/pijalu/artillery-m1-klipper/releases"

# Get the latest release of the specified type
RESPONSE=$(curl -s "$API_URL")

# Extract tag name for our release type
TAG_NAME=$(echo "$RESPONSE" | awk -v type="$RELEASE_TYPE" '
    /"tag_name":/ && index($0, "\"" type "-") {
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
    /"tag_name":/ && index($0, "\"" tag "\"") { found_tag = 1; next }
    found_tag && /"browser_download_url":/ {
        match($0, /https:\/\/[^"]+/)
        if (RSTART > 0) {
            print substr($0, RSTART, RLENGTH)
            exit
        }
    }
')

# Extract filename for this tag
FILENAME=$(echo "$RESPONSE" | awk -v tag="$TAG_NAME" '
    /"tag_name":/ && index($0, "\"" tag "\"") { found_tag = 1; next }
    found_tag && /"name":/ && /\.zip/ {
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

# Compare with local version
LOCAL_VERSION_FILE="/home/mks/printer_data/config/.version"
REMOTE_VERSION=$(echo "$FILENAME" | sed 's/\.zip$//' | sed 's/release-//')

if [ -f "$LOCAL_VERSION_FILE" ]; then
    LOCAL_VERSION=$(sed 's/release-//g' "$LOCAL_VERSION_FILE")
    echo "Local version: $LOCAL_VERSION"
    echo "Remote version: $REMOTE_VERSION"
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
curl -L -o "/tmp/$FILENAME" "$DOWNLOAD_URL" || { echo "Error: Download failed"; exit $LINENO; }

echo "Extracting $FILENAME..."
mkdir -p "/tmp/klipper-update" || exit $LINENO
unzip -o "/tmp/$FILENAME" -d "/tmp/klipper-update" || exit $LINENO

CHANGES_MADE=0

# Function to merge variables.cfg files
merge_variables_cfg() {
    SOURCE_FILE="$1"
    TARGET_FILE="$2"
    TEMP_MERGED="/tmp/variables_merged_$"
    CHANGES_DETECTED=0
    
    # If target doesn't exist, simply copy source to target
    if [ ! -f "$TARGET_FILE" ]; then
        $ARTDO cp "$SOURCE_FILE" "$TARGET_FILE" || exit $LINENO
        return 1  # Return 1 to indicate changes were made (file created)
    fi
    
    # Copy existing target to temporary file
    cp "$TARGET_FILE" "$TEMP_MERGED"
    
    # Process source file to extract only the variable lines (skipping the [Variables] header)
    while IFS= read -r line || [ -n "$line" ]; do
        # Skip empty lines and comments
        if [[ -z "$line" ]] || [[ "$line" =~ ^[[:space:]]*# ]] || [[ "$line" =~ ^[[:space:]]*\[ ]]; then
            continue
        fi
        
        # Extract variable name (everything before the first '=')
        if [[ "$line" =~ ^[[:space:]]*([^=]+)= ]]; then
            var_name="${BASH_REMATCH[1]}"
            var_name=$(echo "$var_name" | sed 's/^[[:space:]]*//;s/[[:space:]]*$//')  # trim whitespace
            
            # Check if this variable already exists in target file
            if ! grep -q "^[[:space:]]*$var_name[[:space:]]*=" "$TARGET_FILE"; then
                # Add this variable to the merged file
                echo "$line" >> "$TEMP_MERGED"
                CHANGES_DETECTED=1
            fi
        fi
    done < "$SOURCE_FILE"
    
    # Only copy the merged file if changes were detected
    if [ $CHANGES_DETECTED -eq 1 ]; then
        $ARTDO cp "$TEMP_MERGED" "$TARGET_FILE"
    fi
    
    # Clean up temporary file
    rm -f "$TEMP_MERGED"
    
    # Return whether changes were detected (0 = no changes, 1 = changes made)
    return $CHANGES_DETECTED
}

# ---------------------------------------------------------------------
# Deploy config files
CONFIG_DIR="/tmp/klipper-update/release/config"
TARGET_DIR="/home/mks/printer_data/config"
BACKUP_SUFFIX="pre-$REMOTE_VERSION-$(date +%s)"

echo "Deploying files from $CONFIG_DIR to $TARGET_DIR..."
if [ ! -d "$CONFIG_DIR" ]; then
    echo "Error: Config directory not found in package"
    exit $LINENO
fi

while read -r FILE; do
    REL_PATH="${FILE#$CONFIG_DIR/}"
    TARGET_FILE="$TARGET_DIR/$REL_PATH"
    TARGET_FILE_DIR=$(dirname "$TARGET_FILE")
    mkdir -p "$TARGET_FILE_DIR" || exit $LINENO

    if [ "$(basename "$TARGET_FILE")" = "moonraker-obico.cfg" ]; then
        if ! grep -q '#auth_token = here.be.dragons' "$TARGET_FILE" 2>/dev/null; then
            echo "User modified file: $REL_PATH"
            continue
        fi
    fi

    if [ -f "$TARGET_FILE" ]; then
        if ! cmp -s "$FILE" "$TARGET_FILE"; then
            # Special case: Merge variables.cfg to preserve user settings while adding new ones
            if [ "$(basename "$TARGET_FILE")" = "variables.cfg" ]; then
                echo "Merging $REL_PATH"
                merge_variables_cfg "$FILE" "$TARGET_FILE"
                if [ $? -eq 1 ]; then
                    # Function returned 1, meaning changes were made
                    CHANGES_MADE=1
                fi
                continue
            fi

            echo "Updating $REL_PATH"
            $ARTDO cp "$TARGET_FILE" "${TARGET_FILE}-$BACKUP_SUFFIX" || exit $LINENO
            $ARTDO cp "$FILE" "$TARGET_FILE" || exit $LINENO
            CHANGES_MADE=1
        fi
    else
        echo "Adding new file $REL_PATH"
        $ARTDO cp "$FILE" "$TARGET_FILE" || exit $LINENO
        CHANGES_MADE=1
    fi
done < <(find "$CONFIG_DIR" -type f)

# ---------------------------------------------------------------------
# Deploy extras files
CONFIG_DIR="/tmp/klipper-update/release/extras"
TARGET_DIR="/home/mks/klipper/klippy/extras"
BACKUP_SUFFIX="pre-$REMOTE_VERSION-$(date +%s)"

echo "Deploying files from $CONFIG_DIR to $TARGET_DIR..."
if [ ! -d "$CONFIG_DIR" ]; then
    echo "Error: Extras directory not found in package"
    exit $LINENO
fi

while read -r FILE; do
    REL_PATH="${FILE#$CONFIG_DIR/}"
    TARGET_FILE="$TARGET_DIR/$REL_PATH"
    TARGET_FILE_DIR=$(dirname "$TARGET_FILE")
    mkdir -p "$TARGET_FILE_DIR" || exit $LINENO

    if [ -f "$TARGET_FILE" ]; then
        if ! cmp -s "$FILE" "$TARGET_FILE"; then
            echo "Updating $REL_PATH"
            sudo cp "$TARGET_FILE" "${TARGET_FILE}-$BACKUP_SUFFIX" || exit $LINENO
            sudo cp "$FILE" "$TARGET_FILE" || exit $LINENO
            sudo chown linaro:linaro "$TARGET_FILE" || exit $LINENO
            PYCACHE_FILE=$(basename "$TARGET_FILE" .py).cpython-*.pyc
            sudo rm "$TARGET_DIR/__pycache__/$PYCACHE_FILE" 2>/dev/null || true
            CHANGES_MADE=1
        fi
    else
        echo "Adding new file $REL_PATH"
        sudo cp "$FILE" "$TARGET_FILE" || exit $LINENO
        sudo chown linaro:linaro "$TARGET_FILE" || exit $LINENO
        CHANGES_MADE=1
    fi
done < <(find "$CONFIG_DIR" -type f)

# ---------------------------------------------------------------------
echo "$CHANGES_MADE changes made during deployment"

if [ "$CHANGES_MADE" -eq 1 ]; then
    echo "Updating local version file..."
    $ARTDO cp /tmp/klipper-update/release/version /home/mks/printer_data/config/.version || exit $LINENO

    echo "Restarting Klipper and Moonraker services..."
    sudo systemctl restart crowsnest || exit $LINENO
    sudo systemctl restart klipper || exit $LINENO
    sudo systemctl restart moonraker || exit $LINENO
    echo "Services restarted successfully"
else
    echo "No changes made, services not restarted"
fi

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
