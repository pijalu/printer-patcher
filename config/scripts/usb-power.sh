#!/bin/bash

sudo tee /etc/udev/rules.d/50-usb-power.rules <<EOF
    ACTION=="add|change", SUBSYSTEM=="usb", ENV{DEVTYPE}=="usb_interface", \
        ATTR{bInterfaceClass}=="0e", TEST=="../power/control", ATTR{../power/control}="on"

    ACTION=="add", SUBSYSTEM=="usb", KERNEL=="2-1", TEST=="power/control", ATTR{power/control}="on"
EOF

echo "[OK]"