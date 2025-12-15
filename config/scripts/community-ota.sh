#!/bin/bash

VERSION="v0.0.1"

# Define file paths
OTA_SERVER_PATH="/usr/local/bin/ota-server-linux-arm64"
RC_LOCAL="/etc/rc.local"
HOSTS_FILE="/etc/hosts"
HOST_ENTRY="127.0.0.1       studio.ota.artillery3d.com"

# Use wget -N to only download if newer
cd /usr/local/bin
if sudo wget -N "https://github.com/pijalu/artillery-ota-server/releases/download/${VERSION}/ota-server-linux-arm64"; then
    # If wget returns success (0), it means the file was downloaded (newer version available)
    echo "Newer version downloaded."

    # Make sure the file is executable
    sudo chmod +x "$OTA_SERVER_PATH"

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
fi

echo "[OK]"