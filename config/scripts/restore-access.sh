#!/bin/bash
CHANGES=0

# Revert any possible change of directory on fluidd
if grep -q 'root /home/mks/ws/wyd' '/etc/nginx/sites-available/fluidd'; then
    echo "* Reverting workaround nginx changes"
    sudo sed -i 's|root /home/mks/ws/wyd|root /home/mks/fluidd|g' /etc/nginx/sites-available/fluidd || exit $LINENO
    CHANGES=1
else
    echo "* No workaround detected"
fi

if [ ! -f /home/mks/printer_data/gcodes/sda1/factory_mode ]; then
    echo "* Enabling factory mode"
    sudo touch /home/mks/printer_data/gcodes/sda1/factory_mode || exit $LINENO
    CHANGES=1
else
    echo "* Already running in factory mode"
fi

if [ "$CHANGES" == "1" ]; then
    # Force rename of fluidd now
    sudo mv /home/mks/ws/wyd /home/mks/fluidd
    # Do restart mksclient
    echo "* Restarting makerbase client (LCD)"
    sudo systemctl restart makerbase-client || exit $LINENO
    # Restart NGINX
    echo "* Restarting NGINX (server)"
    sudo systemctl restart nginx || exit $LINENO
else
    echo "* No changes required"
fi

echo "[OK] - Restore Access"
