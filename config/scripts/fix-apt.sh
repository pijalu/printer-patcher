#!/bin/bash
if grep -q 'mirrors.ustc.edu.cn' '/etc/apt/sources.list'; then
    echo "* Fixing apt source"
    sudo mv /etc/apt/sources.list /etc/apt/sources.list.$$ || exit $LINENO
    sudo tee /etc/apt/sources.list <<EOF
deb http://deb.debian.org/debian bullseye main contrib
deb-src http://deb.debian.org/debian bullseye main contrib
deb http://security.debian.org/debian-security bullseye-security main contrib
deb-src http://security.debian.org/debian-security bullseye-security main contrib
deb http://deb.debian.org/debian bullseye-updates main contrib
deb-src http://deb.debian.org/debian bullseye-updates main contrib
EOF
#    echo "* Updating apt"
#    sudo apt-get update
else
    echo "* not change on APT source"
fi

echo "[OK] - apt source list"