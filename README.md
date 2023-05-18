# DSUL - Disturb State USB Light : Go

[![License MIT](https://img.shields.io/badge/license-MIT-blue.svg)](https://opensource.org/licenses/MIT)
![release](https://img.shields.io/github/v/release/hymnis/dsul-go?include_prereleases)
![version](https://img.shields.io/github/go-mod/go-version/hymnis/dsul-go)
![main test](https://github.com/hymnis/dsul-go/actions/workflows/test.yml/badge.svg?branch=main)
![main analysis](https://github.com/hymnis/dsul-go/actions/workflows/codeql-analysis.yml/badge.svg?branch=main)

The goal of the project is to have a USB connected light, that can be be set to different colors, with adjustable brightness and different modes, which can communicate the users current preference regarding being disturbed.

This implementation uses Go for both daemon/server and client. It should work on most platforms as it uses as few and standard libraries as possible.


## Hardware

The hardware used is an Arduino connected to a NeoPixel module. The project was developed using an Arduino Nano, but should work on most models as long as the firmware fit and it has enough RAM for the number of LED's used in the module.

The firmware projects are available at:
- [hymnis/dsul-arduino](https://github.com/hymnis/dsul-arduino)
- [hymnis/dsul-rp2040](https://github.com/hymnis/dsul-rp2040)


## Firmware

As both FW (firmware) and SW (software) needs to talk to each other, not all combinations of versions work. Make sure that the FW and SW versions are compatible with each other. The latest (stable) versions usually has the best support. For more information about compatibility, see the [Arduino firmware](https://github.com/hymnis/dsul-go/wiki/Firmware) wiki page.

Specifications on the serial protocol and messages can be found in both the [Arduino firmware](https://github.com/hymnis/dsul-arduino) and [RP2040 firmware](https://github.com/hymnis/dsul-rp2040) repositories.


## Other implementations

There's also a sister project to this one that uses Python instead, called [dsul-python](https://github.com/hymnis/dsul-python). It uses a different approach to IPC and has support for TCP sockets and can work over network but lacks encryption or security.


## Installation (manual)

DSUL aims to be a proper Go project and can be built into packages.


### Build package(s)

```bash
go build -ldflags "-X main.version=$(head -1 ./cmd/dsuld/VERSION) -X main.sha1=$(git rev-parse HEAD) -X main.buildTime=$(date +'%Y-%m-%d_%T')" -o dsuld ./cmd/dsuld/main.go
go build -ldflags "-X main.version=$(head -1 ./cmd/dsulc/VERSION) -X main.sha1=$(git rev-parse HEAD) -X main.buildTime=$(date +'%Y-%m-%d_%T')" -o dsulc ./cmd/dsulc/main.go
```


## Daemon, dsuld
This part handles communication with the hardware (serial connection) and allows clients to send commands.

If password is set (`-p <password>`) then all clients are required to supply the password when sending commands.
Message that does not supply the correct password will be ignored.

As module: `go run ./cmd/dsuld/main.go [arguments]`  
As binary: `dsuld [arguments]`

### Arguments

    -h, --help                 Show help and usage information.
    -c, --comport <comport>    The COM port to use. [default: /dev/ttyUSB0]
    -b, --baudrate <baudrate>  The baudrate to use with the COM port. [default: 38400]
    -n  --network              Enable network mode.
    -p  --password <password>  Set password.
    -v, --version              Show current version.
    --verbose                  Show more detailed output.
    --debug                    Show debug output.


## CLI client, dsulc
Used to communicate with the daemon through IPC.

As module: `go run ./cmd/dsulc/main.go [arguments]`  
As binary: `dsulc [arguments]`

### Arguments

    -h, --help                     Show help and usage information.
    -l, --list                     List acceptable values for color, brightness and mode.
    -c, --color <color>            Set color to given value (must be one of the predefined colors).
    -b, --brightness <brightness>  Set brightness to given value.
    -m, --mode <mode>              Set mode to given value (must be on of the predefined modes).
    -d, --dim                      Turn on color dimming.
    -u, --undim                    Turn off color dimming.
    -n  --network <server>         Network server to connect to.
    -p  --password <password>      Set password.
    -v, --version                  Show current version.
    --verbose                      Show more detailed output.
    --debug                        Show debug output.


## Development
This is the basic flow for development on the project. Step 1-2 should only have to be run once, while 3-8 is the continuous development cycle.

1. Initialize pre-commit (`pre-commit install`)
0. Create feature branch
0. Develop stuff
0. Format and lint
0. Test
0. Commit changes
0. Push changes

### Requirements
As this repo uses [pre-commit](https://pre-commit.com/) that does linting and format checking. [pre-commit](https://pre-commit.com/) must be installed prior to commit, for it to work.

### Formatting
All Go code should be formatted by `gofmt`. If it's not it will be caught by the pre-commit hook.

### Linting, checks and test
Tests are located in the module directory. They should be named according to format: `<module name>_test.go`

The program (cmd) modules also use Gingko and test suites and there should be a `<module name>_suite_test.go` and a `main_test.go` file for each module.

To check the code itself we use `go vet`.

### pre-commit
Current configuration will lint and format check as well as check files for strings (like "TODO" and "DEBUG") and missed git merge markings.
Look in `.pre-commig-config.yaml` for exact order of tasks and their settings.
