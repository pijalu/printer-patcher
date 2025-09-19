#!/bin/bash

# Cross-platform build script for 3D Printer Patcher
echo "Building 3D Printer Patcher (cross)..."

# Build for desktop platforms
for type in {linux,windows}
do
    for arch in {arm64,amd64}
    do
        fyne-cross $type -arch $arch -icon icon.png -name "PrinterPatcher" --app-id "com.github.pijalu.printerpatcher" --app-build 1 --app-version 1.0.0  || exit 42
    done
done

# Build for Android ARM64 only
echo "Building for Android ARM64..."
fyne-cross android -arch arm64 -icon icon.png -name "PrinterPatcher" 
    --app-id "com.github.pijalu.printerpatcher" 
    --app-build 1 
    --app-version 1.0.0 || exit 42

