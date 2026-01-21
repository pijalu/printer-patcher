#!/bin/bash

if [ ! -f /etc/udev/rules.d/50-usb-power.rules ];
then
    sudo tee /etc/udev/rules.d/50-usb-power.rules <<EOF
    ACTION=="add", SUBSYSTEM=="usb", KERNEL=="1-1", TEST=="power/control", ATTR{power/control}="on"
    ACTION=="add", SUBSYSTEM=="usb", KERNEL=="2-1", TEST=="power/control", ATTR{power/control}="on"
EOF
else
    echo "* USB Power already installed"
fi

echo "[OK]"