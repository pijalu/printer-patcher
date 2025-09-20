#!/bin/bash
sudo chown -R artillery:netdev /home/mks/printer_data/config/ || exit $LINENO

echo "[OK] - Fix rights"