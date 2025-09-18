#!/bin/bash

# Build script for 3D Printer Patcher

echo "Building 3D Printer Patcher..."

# Create bin directory if it doesn't exist
mkdir -p dist/current
SRC=$PWD
# Build for current platform with icon
echo "Building for current platform with icon..."
(cd dist/current && fyne package --source-dir $SRC --icon icon.png --name "PrinterPatcher" --app-id "com.github.pijalu.printer-patcher" --app-build 1 --app-version 1.0.0)

# On macOS, also create a DMG package
if [[ "$OSTYPE" == "darwin"* ]]; then
    echo "Creating DMG package for macOS..."
    cd dist/current
    # Create a temporary directory for the DMG
    mkdir -p dmg_tmp
    cp -R "PrinterPatcher.app" dmg_tmp/
    
    # Create symbolic link to Applications
    ln -s /Applications dmg_tmp/Applications
    
    # Create the DMG
    hdiutil create -volname "PrinterPatcher" -srcfolder dmg_tmp -ov -format UDZO "../PrinterPatcher.dmg"
    
    # Clean up
    rm -rf dmg_tmp
    cd ../..
fi

echo "Build completed! Check current directory for the application."
