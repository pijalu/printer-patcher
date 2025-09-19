#!/bin/bash

# Linux build script for 3D Printer Patcher

echo "Building 3D Printer Patcher for Linux..."

# Build for Linux platforms
for arch in {arm64,amd64}
do
    fyne-cross linux -arch $arch -icon icon.png -name "PrinterPatcher" --app-id "com.github.pijalu.printerpatcher" --app-build 1 --app-version 1.0.0  || exit 42
done

echo "Linux builds completed."