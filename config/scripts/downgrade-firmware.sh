#!/bin/bash

sudo wget -O /tmp/yuntu_m1.deb https://github.com/pijalu/artillery-m1-debs/raw/refs/heads/main/Yuntu_m1-Yuntu_m1_client_deb-129.deb && sudo dpkg -i --force-downgrade /tmp/yuntu_m1.deb && echo "[OK] - Downgrade"

