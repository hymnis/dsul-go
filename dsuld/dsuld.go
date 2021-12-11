// DSUL - Disturb State USB Light : Daemon application.
package main

import (
	"errors"
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/akamensky/argparse"
	"github.com/hymnis/dsul-go/ipc"
	"github.com/hymnis/dsul-go/serial"
	"github.com/hymnis/dsul-go/settings"
)

var verbose bool = false
var pkg_version string = "0.0.1-alpha"

func main() {
	// Get settings and arguments
	cfg := settings.GetSettings()
	handleArguments(cfg)

	// Start runners
	cmd_channel := make(chan string) // commands to serial device
	rsp_channel := make(chan string) // response from serial device
	go serial.Runner(cfg, cmd_channel, rsp_channel)
	go ipc.ServerRunner(cmd_channel, rsp_channel)

	select {} // run until user exits
}

// Parse command line arguments .
func handleArguments(cfg *settings.Config) {
	parser := argparse.NewParser("dsuld", "Disturb State USB Light - Daemon")

	arg_comport := parser.String("c", "comport", &argparse.Options{
		Required: false,
		Validate: func(args []string) error {
			for _, comport := range args {
				if _, err := os.Stat(comport); !os.IsNotExist(err) {
					return nil
				}
			}
			return errors.New("COM port path incorrect or not readable.")
		},
		Help: "Set COM port (path)"})
	arg_baudrate := parser.Int("b", "baudrate", &argparse.Options{
		Required: false,
		Validate: func(args []string) error {
			for _, baudrate := range args {
				if n, err := strconv.ParseInt(baudrate, 10, 64); err != nil || int(n) < 9600 || int(n) > 38400 {
					return errors.New("Baudrate is outside allowed range (9600-38400).")
				}
			}
			return nil
		},
		Help: "Set COM port baudrate"})
	arg_version := parser.Flag("v", "version", &argparse.Options{
		Required: false,
		Help:     "Show version"})
	arg_verbose := parser.Flag("", "verbose", &argparse.Options{
		Required: false,
		Help:     "Show verbose output"})

	err := parser.Parse(os.Args)
	if err != nil {
		// This can also be done by passing -h or --help
		fmt.Print(parser.Usage(err))
		os.Exit(1)
	}

	// Handle arguments
	if *arg_verbose {
		verbose = true
		log.Println("Verbose mode is on")
	}
	if *arg_version {
		fmt.Printf("dsuld v%s\n", pkg_version)
		os.Exit(0)
	}
	if *arg_comport != "" {
		if verbose {
			log.Printf("Set COM port: %v\n", *arg_comport)
		}
		cfg.Serial.Port = *arg_comport
	}
	if *arg_baudrate > 0 {
		if verbose {
			log.Printf("Set COM port baudrate: %d\n", *arg_baudrate)
		}
		cfg.Serial.Baudrate = *arg_baudrate
	}
}
