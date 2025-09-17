#!/bin/bash

# Build script for 3D Printer Patcher

echo "Building 3D Printer Patcher..."

# Create bin directory if it doesn't exist
mkdir -p dist/current
SRC=$PWD
# Build for current platform with icon
echo "Building for current platform with icon..."
(cd dist/current && fyne package --source-dir $SRC --icon icon.png --name "PrinterPatcher" --app-id "com.github.pijalu.printer-patcher" --app-build 1 --app-version 1.0.0)

echo "Build completed! Check current directory for the application."
