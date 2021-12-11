// DSUL - Disturb State USB Light : Client application.
package main

import (
	"errors"
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/akamensky/argparse"
	"github.com/hymnis/dsul-go/ipc"
	"github.com/hymnis/dsul-go/settings"
)

var verbose bool = false
var pkg_version string = "0.0.1-alpha"
var hardware_info string = ""

func main() {
	cfg := settings.GetSettings()
	ipc_command := make(chan ipc.Command)
	ipc_response := make(chan ipc.Command)
	done := make(chan bool)
	go ipc.ClientRunner(ipc_command, ipc_response, done) // act on ipc_command's given and send 'done' signal all are sent
	go handleResponse(cfg, ipc_response)                 // handle responses from IPC daemon
	handleArguments(cfg, ipc_command)                    // read arguments and send ipc_command's
	close(ipc_command)                                   // close channel once we are done sending commands

	<-done // run until 'done' signal is received
}

// Parse command line arguments and send IPC commands.
func handleArguments(cfg *settings.Config, ipc_command chan ipc.Command) {
	parser := argparse.NewParser("dsulc", "Disturb State USB Light - CLI")

	arg_color := parser.String("c", "color", &argparse.Options{
		Required: false,
		Validate: func(args []string) error {
			for _, color := range args {
				for _, cfg_color := range cfg.Colors {
					if cfg_color.Name == color {
						return nil
					}
				}
			}
			return errors.New("Color given is not supported.")
		},
		Help: "Set given color"})
	arg_list := parser.Flag("l", "list", &argparse.Options{
		Required: false,
		Help:     "List settings and values"})
	arg_mode := parser.String("m", "mode", &argparse.Options{
		Required: false,
		Validate: func(args []string) error {
			for _, mode := range args {
				for _, cfg_mode := range cfg.Modes {
					if cfg_mode.Name == mode {
						return nil
					}
				}
			}
			return errors.New("Mode given is not supported.")
		},
		Help: "Set given mode"})
	arg_brightness := parser.Int("b", "brightness", &argparse.Options{
		Required: false,
		Validate: func(args []string) error {
			for _, mode := range args {
				if n, err := strconv.ParseInt(mode, 10, 64); err != nil || int(n) < cfg.BrightnessMin || int(n) > cfg.BrightnessMax {
					msg := fmt.Sprintf("Brightness must be between %d and %d.", cfg.BrightnessMin, cfg.BrightnessMax)
					return errors.New(msg)
				}
			}
			return nil
		},
		Help: "Set given brightness"})
	arg_dim := parser.Flag("d", "dim", &argparse.Options{
		Required: false,
		Help:     "Dim colors"})
	arg_undim := parser.Flag("u", "undim", &argparse.Options{
		Required: false,
		Help:     "Un-dim colors"})
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
	actions := 0
	if *arg_verbose {
		verbose = true
		log.Println("Verbose mode is on")
	}
	if *arg_version {
		fmt.Printf("dsulc v%s\n", pkg_version)
		os.Exit(0)
	}
	if *arg_list {
		ipc_command <- ipc.Command{"get", "information", "all"} // request information from daemon
		time.Sleep(time.Second * 1)                             // wait to cmd to be sent and response received and parsed
		listInformation(cfg)
		os.Exit(0)
	}
	if *arg_mode != "" {
		if verbose {
			log.Printf("Set mode: %v\n", *arg_mode)
		}
		ipc_command <- ipc.Command{"set", "mode", *arg_mode}
		actions += 1
	}
	if *arg_brightness > 0 {
		if verbose {
			log.Printf("Set brightness: %d\n", *arg_brightness)
		}
		ipc_command <- ipc.Command{"set", "brightness", fmt.Sprint(*arg_brightness)}
		actions += 1
	}
	if *arg_dim {
		if verbose {
			log.Print("Set dim\n")
		}
		ipc_command <- ipc.Command{"set", "dim", "true"}
		actions += 1
	}
	if *arg_undim {
		if verbose {
			log.Print("Set un-dim\n")
		}
		ipc_command <- ipc.Command{"set", "undim", "true"}
		actions += 1
	}
	if *arg_color != "" {
		if verbose {
			log.Printf("Set color: %v\n", *arg_color)
		}
		ipc_command <- ipc.Command{"set", "color", *arg_color}
		actions += 1
	}

	// Handle actions
	if actions == 0 {
		fmt.Print(parser.Usage(nil))
		close(ipc_command)
		os.Exit(1)
	}
}

// Handle responses from IPC daemon.
func handleResponse(cfg *settings.Config, ipc_response chan ipc.Command) {
	for {
		select {
		case response := <-ipc_response:
			if len(response.Value) > 4 {
				// Update settings values from hardware limits
				hardware_info = response.Value
				hardware_state := *settings.ParseHardwareInformation(hardware_info)
				if hardware_state.Brightness_min >= 0 {
					cfg.BrightnessMin = hardware_state.Brightness_min
				}
				if hardware_state.Brightness_max > 0 {
					cfg.BrightnessMax = hardware_state.Brightness_max
				}
			}
		}
	}
}

// Return information about configuration settings and current hardware values.
func listInformation(cfg *settings.Config) {
	hardware_state := *settings.ParseHardwareInformation(hardware_info)

	fmt.Println("[modes]")
	for _, cfg_mode := range cfg.Modes {
		fmt.Printf("- %s\n", cfg_mode.Name)
	}

	fmt.Println("\n[colors]")
	for _, cfg_color := range cfg.Colors {
		fmt.Printf("- %s\n", cfg_color.Name)
	}

	fmt.Println("\n[brightness]")
	fmt.Printf("- min = %v\n", cfg.BrightnessMin)
	fmt.Printf("- max = %v\n", cfg.BrightnessMax)

	if hardware_state.Version != "" {
		fmt.Println("\n[hardware values]")
		fmt.Printf("- version = %v\n", hardware_state.Version)
		fmt.Printf("- color = %v\n", hardware_state.Current_color)
		fmt.Printf("- mode = %v\n", hardware_state.Current_mode)
		fmt.Printf("- brightness = %v\n", hardware_state.Current_brightness)
		fmt.Printf("- dim = %v\n", hardware_state.Current_dim)
	}
}
