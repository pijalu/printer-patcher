#!/bin/sh

if compgen -G "/dev/video*" > /dev/null; then
    echo "[OK]"
    exit
fi  

echo "Camera not found - unbinding DWC2"
DWC2_DEV="ff400000.usb"
echo "$DWC2_DEV" | sudo tee /sys/bus/platform/drivers/dwc2/unbind > /dev/null
sleep 2
echo "$DWC2_DEV" | sudo tee /sys/bus/platform/drivers/dwc2/bind > /dev/null
sleep 4

if compgen -G "/dev/video*" > /dev/null; then
    echo "[OK]"
else
    echo "[KO]"
fi