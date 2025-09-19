#!/bin/bash

# Windows build script for 3D Printer Patcher

echo "Building 3D Printer Patcher for Windows..."

# Build for Windows platforms
for arch in {arm64,amd64}
do
    fyne-cross windows -arch $arch -icon icon.png -name "PrinterPatcher" --app-id "com.github.pijalu.printerpatcher" --app-build 1 --app-version 1.0.0  || exit 42
done

echo "Windows builds completed."