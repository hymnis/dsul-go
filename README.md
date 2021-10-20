# DSUL - Disturb State USB Light : Go

[![Build Status](https://travis-ci.com/hymnis/dsul-go.svg?branch=master)](https://travis-ci.com/github/hymnis/dsul-go)
[![Maintainability](https://api.codeclimate.com/v1/badges/<id>/maintainability)](https://codeclimate.com/github/hymnis/dsul-go/maintainability)
[![Test Coverage](https://api.codeclimate.com/v1/badges/<id>/test_coverage)](https://codeclimate.com/github/hymnis/dsul-go/test_coverage)
[![License MIT](https://img.shields.io/badge/license-MIT-blue.svg)](https://opensource.org/licenses/MIT)

The goal of the project is to have a USB connected light, that can be be set to different colors, with adjustable brightness and different modes, which can communicate the users current preference regarding being disturbed.

This implementation used Go for both daemon/server and client. It should work on most platforms as it uses as few and standard libraries as possible.


## Hardware

The hardware used is an Arduino connected to a NeoPixel module. The project was developed using an Arduino Nano, but should work on most models as long as the firmware fit and it has enough RAM for the number of LED's used in the module.

The firmware project is available at [hymnis/dsul-arduino](https://github.com/hymnis/dsul-arduino).


## Firmware

As both FW (firmware) and SW (software) needs to talk to each other, not all combinations of versions work. Make sure that the FW and SW versions are compatible with each other. The latest (stable) versions usually has the best support. For more information about compatibility, see the [Firmware](https://github.com/hymnis/dsul-go/wiki/Firmware) wiki page.


## Installation (manual)

DSUL is a proper Go project and can be built into packages.


### Build package(s)

```
go build
```

### Install package(s)

```
go install
```


## Configuration

...


## Daemon, dsuld
This part handles communication with the hardware (serial connection) and allows clients to send commands.

As module: `go run dsuld.go [arguments]`  
As package: `dsuld [arguments]`

### Options

...


## CLI client, dsulc
Used to communicate with the daemon through IPC.

...


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

To check the code itself we use `go vet`, which both checks the code and runs the tests.

### pre-commit
Current configuration will lint and format check as well as check files for strings (like "TODO" and "DEBUG") and missed git merge markings.
Look in `.pre-commig-config.yaml` for exact order of tasks and their settings.


## Acknowledgements

- ...
