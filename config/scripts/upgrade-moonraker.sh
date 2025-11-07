#!/bin/bash
CHANGES=0

ARTDO="sudo -u artillery "

REMOTE=`cd /home/mks/moonraker && $ARTDO git config --get remote.origin.url`
BRANCH=`cd /home/mks/moonraker && $ARTDO git rev-parse --abbrev-ref HEAD`

if [ "$REMOTE" != "https://github.com/pijalu/artillery-m1-moonraker.git" ]; then
	echo "* Creating backup"
	$ARTDO rm -rf /home/mks/moonraker.bak
	$ARTDO mv /home/mks/moonraker /home/mks/moonraker.bak

	echo "* Checkout moonraker"
	$ARTDO git clone https://github.com/pijalu/artillery-m1-moonraker.git /home/mks/moonraker || exit $LINENO
	cd /home/mks/moonraker
	$ARTDO git checkout feature/diy-macros || exit $LINENO
	
	# Prepare restart
	CHANGES=1
else
	echo "* Checking version"
	cd /home/mks/moonraker
	
	# Check if there are any uncommitted changes
	if ! $ARTDO git diff-index --quiet HEAD --; then
		echo "* Repository is stale - cleaning up"
		$ARTDO git clean -fd || exit $LINENO
		$ARTDO git reset --hard HEAD || exit $LINENO
		CHANGES=1
	fi
	
	if [ "$BRANCH" != "feature/diy-macros" ]; then
		$ARTDO git checkout feature/diy-macros || exit $LINENO
		CHANGES=1
	else
		behind=`$ARTDO git rev-list --count HEAD..@{u}`
		if [ "$behind" -gt 0 ]; then
			$ARTDO git pull || exit $LINENO
			CHANGES=1
		fi
	fi
fi

if [ "$CHANGES" == "1" ]; then 
	sudo systemctl stop moonraker || exit $LINENO
	echo "* Installing moonraker dependencies"
	source /home/mks/moonraker-env/bin/activate
	$ARTDO pip install --upgrade -r /home/mks/moonraker/scripts/moonraker-requirements.txt || exit $LINENO
	#sudo -u artillery python3 -m pip install --no-cache-dir "dbus-fast<=2.28.0" || exit $LINENO
	echo "* Restarting moonraker (server)"
    sudo systemctl start moonraker || exit $LINENO
fi

REMOTE=`cd /home/mks/moonraker && $ARTDO git config --get remote.origin.url`
BRANCH=`cd /home/mks/moonraker && $ARTDO git rev-parse --abbrev-ref HEAD`
echo "[OK] moonraker: $REMOTE - $BRANCH"