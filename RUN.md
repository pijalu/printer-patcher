# Printer Patcher - What This Tool Does

This tool is designed to update and fix common issues with your Artillery M1 Pro 3D printer running firmware version 1.0.11.0. It connects to your printer over your local network and applies several important updates to improve performance and reliability.

## What the Tool Does

When you run this tool, it performs the following actions on your printer:

1. **Enables FLUIDD Web Interface** - Restores access to the FLUIDD web interface that may have been disabled, allowing you to control your printer from a web browser.

2. **Updates FLUIDD Software** - Installs the latest version of FLUIDD (v1.34.4), which provides a better user interface for controlling your printer with new features and bug fixes.

3. **Updates Moonraker** - Installs the latest version of Moonraker, which is the communication layer between the web interface and your printer's control system. This update includes special fixes for Artillery printers.

4. **Updates Klipper macros** - Installs community updated macros that addresses common issues with the printer's operation.

5. **Applies Artillery-Specific Fixes**:
   - Fixes macro commands that help prevent clogging of the extruder
   - Ensures bed leveling is performed at the correct temperature for more accurate results
   - Adjusts system settings for better overall performance

## Important Warnings

### ⚠️ Potential Risk to Your Printer
This tool modifies software on your printer. While it has been tested, there is always a risk that something could go wrong:
- The printer may be left in an incorrect state and can require reflashing
- If your printer stops responding after running this tool, you may need to reinstall the firmware using the method described at https://wiki.artillery3d.com/m1/manual/flashing-guide

### ⏱️ Updates Continue After Tool Exits
After the tool finishes running, the actual updating process continues on the printer itself:
- After you finish update process *do not turn off the printer*. http://printerip:8078 will show bad gateway error *for ~10 mins*. When update process will be finished you'll be able to use Fluidd.
- Do not turn off your printer during this time
- You can monitor the progress through the FLUIDD web interface at: http://printerip:8078 (replace "printerip" with your printer's actual IP address)

### 🔄 Calibration Required After Installation
After the updates are complete, you must perform a complete calibration of your printer:
1. Run bed leveling
2. Calibrate extruder steps/mm
3. Run PID tuning for the hotend and bed
4. Perform a test print to verify everything is working correctly

## How to Monitor Progress

After running the tool:
1. Open a web browser on your computer or phone
2. Navigate to: http://printerip:8078 (replace "printerip" with your printer's IP address)
3. Watch the interface to see when updates are complete
4. The system may restart several times during the process - this is normal

## If Something Goes Wrong

If your printer becomes unresponsive or doesn't work properly after running this tool:
1. Try restarting the printer
2. If that doesn't work, you may need to reflash the firmware using [USB OTG](https://wiki.artillery3d.com/m1/manual/flashing-guide)
3. Contact support with details about what happened