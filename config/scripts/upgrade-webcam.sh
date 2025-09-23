#!/bin/bash
ARTDO="sudo -u artillery "

# Update crowsnest resolution if needed
if grep -q 'resolution: 640x480' /home/mks/printer_data/config/crowsnest.conf; then
    $ARTDO sed -i 's/resolution: 640x480/resolution: 1920x1080/' /home/mks/printer_data/config/crowsnest.conf || exit $LINENO
    sudo systemctl restart crowsnest || exit $LINENO
else
    echo "* No change required on crowsnest resolution"
fi

# Update moonraker webcam service if needed
if grep -q 'service: mjpegstreamer-adaptive' /home/mks/printer_data/config/moonraker.conf; then
    $ARTDO sed -i 's/service: mjpegstreamer-adaptive/service: mjpegstreamer/' /home/mks/printer_data/config/moonraker.conf || exit $LINENO
    sudo systemctl restart moonraker || exit $LINENO
else
    echo "* No change required on moonraker service"
fi

echo "[OK] - webcam"
