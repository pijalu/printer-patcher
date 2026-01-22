#!/bin/bash

sudo tee /etc/udev/rules.d/99-usb-power.rules <<EOF
ACTION=="add", SUBSYSTEM=="usb", ENV{DEVTYPE}=="usb_device", ATTR{idVendor}=="090c", ATTR{idProduct}=="337b",ATTR{power/control}="on"
ACTION=="add", SUBSYSTEM=="usb", KERNEL=="usb2", ATTR{power/control}="on"
EOF

echo "[OK]"