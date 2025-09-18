# 3D Printer Patcher
A cross-platform GUI tool for patching 3D printers via SSH.
The use case is to provide an easier way to install/patch typical 3D printer via SSH.

## Features

- Cross-platform GUI (Windows, macOS, Linux)
- SSH connection to 3D printers with configurable credentials
- Configuration-based actions with step-by-step execution
- Script validation with expected output checking
- Built-in default scripts that work out of the box
- Visual feedback for connection status and execution progress
- headless modes for automated operations
- Real-time script execution output display
- Modal dialogs for execution
- Simple, user-friendly interface

## Prerequisites

1. Install Go (version 1.16 or later)
2. Install Git
3. For GUI support on macOS, you might need to install Xcode command line tools
4. For GUI support on Linux, you might need to install development packages:
   ```
   sudo apt-get install gcc libgl1 libxcursor-dev libxrandr-dev libxinerama-dev libxi-dev libxxf86vm-dev
   ```
5. Docker/podman for cross build

## Dependencies

This project uses the following Go modules:
- fyne.io/fyne/v2 - Cross-platform GUI framework
- gopkg.in/yaml.v2 - YAML parsing
- golang.org/x/crypto/ssh - SSH client

## Building

### Using the build script (recommended):
It will build the executabe for the current platform
```bash
./build.sh
```

The executable will be created in the directory dist/current.

### Building for all platforms (except darwin):
```bash
./build-cross.sh
```

This will create a packaged version for windows and linux (arm64 / amd64). Output will be located in fyne-cross/dist

### Manual build:

To build the application manually in dist:
```bash
go build -o dist/printer-patcher
```

To package with icon:
```bash
fyne package -icon icon.png -name "PrinterPatcher" --app-id "com.github.pijalu.printer-patcher" --app-build 1 --app-version 1.0.0
```

## Running

### GUI mode (default):

```bash
./PrinterPatcher
```

Or on macOS, you can simply double-click the PrinterPatcher.app bundle.

### Console mode:

```bash
./PrinterPatcher -console
```

### Headless mode:

```bash
./PrinterPatcher -headless -ip 192.168.0.106 -action "Test Connection 106" -username linaro -password linaro
```

### Command line options:

- `-headless` - Run in headless mode (no GUI)
- `-ip` - Printer IP address (for headless mode)
- `-username` - SSH username (default: linaro)
- `-password` - SSH password (default: linaro)
- `-action` - Action to execute (for headless mode)

## Configuration

The application uses YAML configuration files to define actions - it will use `config/actions.yaml`.
The content of the yaml and sub scripts directory will be included in the executable

- `username`: username to use to connect
- `password`: password to use for SSH

Each action has:
- `title`: A descriptive title for the action
- `description`: A short explanation of what the action does
- `steps`: A list of steps to execute

Each step has:
- `title`: A descriptive title for the step
- `description`: A short explanation of what the step does
- `script`: The shell commands to execute on the printer (can be a file path to a script)
- `expected` (optional): The expected output to validate the step success

Default SSH credentials are linaro/linaro.

## MacOS install
If you downloaded and installed the app but are greeted with a corrupted name: 
`xattr -cr /Applications/PrinterPatcher.app`

## Script Directory
The application includes scripts in the `config/scripts/` directory.

## License
This project is licensed under the GPLv3 License.