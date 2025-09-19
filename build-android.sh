#!/bin/bash

# Build script for 3D Printer Patcher Android APK

echo "Building 3D Printer Patcher for Android ARM64..."

# Check if fyne-cross is installed
if ! command -v fyne-cross &> /dev/null
then
    echo "Error: fyne-cross is not installed. Please install it with:"
    echo "  go install github.com/fyne-io/fyne-cross@latest"
    exit 1
fi

# Build for Android ARM64 only
fyne-cross android -arch arm64 -icon icon.png -name "PrinterPatcher" \
    --app-id "com.github.pijalu.printerpatcher" \
    --app-build 1 \
    --app-version 1.0.0

if [ $? -eq 0 ]; then
    echo "Android ARM64 APK build completed successfully!"
    echo "APK file can be found in the fyne-cross/dist/ directory"
    
    # List the generated APK file
    echo "Generated APK file:"
    find fyne-cross/dist -name "*.apk" -type f | while read -r apk; do
        echo "  - $apk"
    done
else
    echo "Error: Android ARM64 APK build failed"
    exit 1
fi
