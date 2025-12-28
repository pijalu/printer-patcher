#!/bin/bash

VERSION="v0.0.3"
SHA256="0df55c456f51526496821befc54ef0910898b2ce2a09439a7ef34e09ded75f33"

# Define file paths
OTA_SERVER_PATH="/usr/local/bin/ota-server-linux-arm64"
RC_LOCAL="/etc/rc.local"
HOSTS_FILE="/etc/hosts"
HOST_ENTRY="127.0.0.1       studio.ota.artillery3d.com"

cd /usr/local/bin

# Check if the file exists and its checksum matches
DOWNLOAD_REQUIRED=false

if [ -f "$OTA_SERVER_PATH" ]; then
    CURRENT_SHA256=$(sha256sum "$OTA_SERVER_PATH" | cut -d' ' -f1)
    if [ "$CURRENT_SHA256" != "$SHA256" ]; then
        echo "Checksum mismatch. Deleting existing file."
        sudo rm "$OTA_SERVER_PATH"
        DOWNLOAD_REQUIRED=true
    else
        echo "File exists and checksum matches. Nothing to do."
    fi
else
    echo "File does not exist."
    DOWNLOAD_REQUIRED=true
fi

# Download the file if required
if [ "$DOWNLOAD_REQUIRED" = true ]; then
    echo "Downloading new version..."
    if sudo wget -O "$OTA_SERVER_PATH" "https://github.com/pijalu/artillery-ota-server/releases/download/${VERSION}/ota-server-linux-arm64"; then
        # Verify the checksum of the downloaded file
        DOWNLOADED_SHA256=$(sha256sum "$OTA_SERVER_PATH" | cut -d' ' -f1)
        if [ "$DOWNLOADED_SHA256" != "$SHA256" ]; then
            echo "ERROR: Downloaded file checksum does not match expected value!" >&2
            exit ${LINENO}
        fi
        echo "New version downloaded and verified."
    else
        echo "ERROR: Failed to download the file!" >&2
        exit ${LINENO}
    fi

    # Make sure the file is executable
    sudo chmod +x "$OTA_SERVER_PATH"
fi

# Add the hosts entry if it doesn't already exist
if ! sudo grep -q "studio.ota.artillery3d.com" "$HOSTS_FILE"; then
    echo "$HOST_ENTRY" | sudo tee -a "$HOSTS_FILE"
    echo "Added hosts entry for studio.ota.artillery3d.com"
fi

# Check if the service is already in rc.local
if [ -f "$RC_LOCAL" ]; then
    # Check if ota-server is already in rc.local
    if ! sudo grep -q "ota-server-linux-arm64" "$RC_LOCAL"; then
        # Insert before the exit line
        sudo sed -i '/^exit 0/i\\sudo -u nobody /usr/local/bin/ota-server-linux-arm64 \&' "$RC_LOCAL"
        echo "Added OTA server to rc.local"
    fi
else
    # Create rc.local with the service
    echo "#!/bin/bash" | sudo tee "$RC_LOCAL"
    echo "sudo -u nobody /usr/local/bin/ota-server-linux-arm64 &" | sudo tee -a "$RC_LOCAL"
    echo "exit 0" | sudo tee -a "$RC_LOCAL"
    sudo chmod +x "$RC_LOCAL"
    echo "Created new rc.local with OTA server"
fi

# Restart rc.local service to apply changes since file was updated
if systemctl is-active --quiet rc-local; then
    echo "Restarting rc-local service..."
    sudo systemctl restart rc-local
elif systemctl is-active --quiet rclocal; then
    echo "Restarting rclocal service..."
    sudo systemctl restart rclocal
else
    echo "rc.local service not found or not active, attempting to start..."
    sudo systemctl start rc-local 2>/dev/null || sudo systemctl start rclocal 2>/dev/null
fi

echo "[OK]"