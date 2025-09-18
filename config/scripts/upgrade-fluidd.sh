#!/bin/bash
ARTDO="sudo -u artillery"

if [ ! -f /home/mks/fluidd/.version ]; then
    echo "Fluidd not found !"
    exit $LINENO
fi

if grep -q 'v1.34.4' /home/mks/fluidd/.version; then
    echo "* Fluidd version is up to date"
else
    echo "* Downloading FLUIDD"
    curl -L https://github.com/fluidd-core/fluidd/releases/download/v1.34.4/fluidd.zip -o /tmp/fluidd.zip || exit $LINENO

    echo "* Creating backup"
    $ARTDO rm -rf /home/mks/fluidd.bak
    $ARTDO mv /home/mks/fluidd /home/mks/fluidd.bak
    $ARTDO mkdir -p /home/mks/fluidd

    echo "* Extracting FLUIDD"
    $ARTDO unzip /tmp/fluidd.zip -d /home/mks/fluidd || exit $LINENO
    rm /tmp/fluidd.zip

    echo "* Restarting NGINX (server)"
    sudo systemctl restart nginx || exit $LINENO
    echo "* Restarting moonraker (server)"
    sudo systemctl restart moonraker || exit $LINENO
fi

VERSION=`cat /home/mks/fluidd/.version`
echo "[OK] - Using Fluidd version $VERSION"