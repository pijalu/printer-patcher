#!/bin/bash

# Cross-platform build script for 3D Printer Patcher
echo "Building 3D Printer Patcher (cross)..."

# Call separate build scripts for Linux and Windows
./build-linux.sh
./build-windows.sh

echo "Desktop builds completed. Android build has been moved to build-android.sh"

