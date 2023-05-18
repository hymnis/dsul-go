/*
DSUL - Disturb State USB Light : Daemon application

dsuld is the daemon/server part of the DSUL project.
It handles communication with the serial device and client via IPC.

Usage:

    dsuld [arguments]

*/
package main

import (
	"errors"
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/akamensky/argparse"
	"github.com/hymnis/dsul-go/internal/ipc"
	"github.com/hymnis/dsul-go/internal/serial"
	"github.com/hymnis/dsul-go/internal/settings"
)

var (
	version   string = "0.0.0"
	sha1      string //lint:ignore U1000 supplied at build time
	buildTime string //lint:ignore U1000 supplied at build time
	verbose   bool   = false
	debug     bool   = false
)

// main runs the main loop and runners for serial handling and IPC.
func main() {
	// Get settings and arguments
	cfg := settings.GetSettings()
	handleArguments(cfg)

	output_handling := struct {
		Verbose bool
		Debug   bool
	}{
		Verbose: verbose,
		Debug:   debug,
	}

	// Start runners
	cmd_channel := make(chan string) // commands to serial device
	rsp_channel := make(chan string) // response from serial device
	go serial.Runner(cfg, output_handling, cmd_channel, rsp_channel)
	go ipc.ServerRunner(cfg, output_handling, cmd_channel, rsp_channel)

	select {} // run until user exits
}

// handleArguments parses command line arguments and performs actions based on them.
func handleArguments(cfg *settings.Config) {
	// Parse arguments
	parser := argparse.NewParser("dsuld", "Disturb State USB Light - Daemon")

	arg_comport := parser.String("c", "comport", &argparse.Options{
		Required: false,
		Help:     "Set COM port (path)"})
	arg_baudrate := parser.Int("b", "baudrate", &argparse.Options{
		Required: false,
		Validate: func(args []string) error {
			for _, baudrate := range args {
				if n, err := strconv.Atoi(baudrate); err != nil || int(n) < 9600 || int(n) > 115200 {
					return errors.New("baudrate is outside allowed range (9600-115200)")
				}
			}
			return nil
		},
		Help: "Set COM port baudrate"})
	//arg_network := parser.String("n", "network", &argparse.Options{
	arg_network := parser.Flag("n", "network", &argparse.Options{
		Required: false,
		Help:     "Enable network mode"})
	arg_password := parser.String("p", "password", &argparse.Options{
		Required: false,
		Validate: func(args []string) error {
			for _, password := range args {
				if len(password) > 0 {
					return nil
				}
			}
			return errors.New("password can't be empty")
		},
		Help: "Set password"})
	arg_version := parser.Flag("v", "version", &argparse.Options{
		Required: false,
		Help:     "Show version"})
	arg_verbose := parser.Flag("", "verbose", &argparse.Options{
		Required: false,
		Help:     "Show verbose output"})
	arg_debug := parser.Flag("", "debug", &argparse.Options{
		Required: false,
		Help:     "Show debug output"})

	err := parser.Parse(os.Args)
	if err != nil {
		// This can also be done by passing -h or --help
		fmt.Print(parser.Usage(err))
		os.Exit(1)
	}

	// Handle arguments
	if *arg_debug {
		debug = true
		verbose = true
		log.Println("[dsuld] Debug mode is on")
	}
	if *arg_verbose {
		verbose = true
		log.Println("[dsuld] Verbose mode is on")
	}
	if *arg_version {
		fmt.Printf("dsuld v%s\n", version)
		os.Exit(0)
	}
	if *arg_comport != "" {
		if verbose {
			log.Printf("[dsuld] Set COM port: %v\n", *arg_comport)
		}
		cfg.Serial.Port = *arg_comport
	}
	if *arg_baudrate > 0 {
		if verbose {
			log.Printf("[dsuld] Set COM port baudrate: %d\n", *arg_baudrate)
		}
		cfg.Serial.Baudrate = *arg_baudrate
	}
	if *arg_network {
		if verbose {
			log.Printf("[dsuld] Using network mode. Listening on port: %d\n", cfg.Network.Port)
		}
		cfg.Network.Listen = *arg_network
	}
	if *arg_password != "" {
		if verbose {
			log.Print("[dsuld] Using password authentication.\n")
		}
		cfg.Password = *arg_password
	}
}
