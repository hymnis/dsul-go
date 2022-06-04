// DSUL - Disturb State USB Light : Serial module.
package serial

import (
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"go.bug.st/serial"
	"github.com/hymnis/dsul-go/internal/settings"
	"github.com/hymnis/dsul-go/internal/watchdog"
)

var verbose bool = false

// Initialize serial port.
func Init(cfg *settings.Config) serial.Port {
	mode := &serial.Mode{
		BaudRate: cfg.Serial.Baudrate,
		Parity:   serial.NoParity,
		DataBits: 8,
		StopBits: serial.OneStopBit,
	}
	port, err := serial.Open(cfg.Serial.Port, mode)
	if err != nil {
		log.Fatalf("[serial] Failed to open port: %v", err.Error())
	}
	port.SetReadTimeout(time.Second * 2)
	if verbose {
		log.Printf("[serial] Port set: %d_N81", cfg.Serial.Baudrate)
	}
	time.Sleep(time.Second * 2) // let device boot properly

	return port
}

// Read serial port and return data.
func Read(port serial.Port) string {
	buff := make([]byte, 64)
	output := ""

	for {
		n, err := port.Read(buff)
		if err != nil {
			log.Fatal(err)
			break
		}
		if n == 0 {
			break
		}

		if verbose {
			log.Printf("[serial] Receiving: '%s' %v\n", buff, buff)
		}
		for _, ch := range buff {
			if ch != 0 { // ignore 0 (only used as filler in buffer)
				output += string(ch)
			}

			if ch == 35 { // 35 = #
				return output
			}
		}
		if strings.Contains(string(buff[:n]), "\n") {
			break
		}

		// Clear buffer before next run
		for j := range buff {
			buff[j] = 0
		}
	}

	return ""
}

// Write data to serial port.
func Write(port serial.Port, data []byte) {
	if verbose {
		log.Printf("[serial] Sending: '%s' %v\n", data, data)
	}
	_, err := port.Write(data)
	if err != nil {
		log.Fatal(err)
	}
}

// Send OK to device.
func SendOK(port serial.Port) bool {
	Write(port, []byte("+!#"))
	return true
}

// Send ping to device.
func SendPing(port serial.Port) bool {
	result := performExchange(port, "-?#")
	return isOK(result)
}

// Send request for information to device.
func SendRequest(port serial.Port) string {
	result := performExchange(port, "-!#")
	return result
}

// Send command to set the color.
// ???:???:??? - red:green:blue values, 0-255
func SendColorCommand(port serial.Port, value string, cfg *settings.Config) bool {
	command, ok := GetColorString(value, cfg)
	if ok {
		if verbose {
			log.Printf("[serial] Setting color: '%v'", value)
		}
		result := performExchange(port, command)
		return isOK(result)
	}
	if verbose {
		log.Printf("[serial] Invalid color argument: '%v'", value)
	}

	return false
}

// Send command to set the brightness.
// ??? - brightness value, 0-255
func SendBrightnessCommand(port serial.Port, value string, cfg *settings.Config) bool {
	command, ok := GetBrightnessString(value, cfg)
	if ok {
		if verbose {
			log.Printf("[serial] Setting brightness: '%v'", value)
		}
		result := performExchange(port, command)
		return isOK(result)
	}
	if verbose {
		log.Printf("[serial] Invalid brightness argument: '%v'", value)
	}

	return false
}

// Send command to set the mode.
// ??? - mode value, 0-4
func SendModeCommand(port serial.Port, value string, cfg *settings.Config) bool {
	command, ok := GetModeString(value, cfg)
	if ok {
		if verbose {
			log.Printf("[serial] Setting mode: '%v'", value)
		}
		result := performExchange(port, command)
		return isOK(result)
	}
	if verbose {
		log.Printf("[serial] Invalid mode argument: '%v'", value)
	}

	return false
}

// Send command to set the dim mode.
// 0 = No dimming (turn off dim mode)
// 1 = Dimming (turn on dim mode)
func SendDimCommand(port serial.Port, value string) bool {
	command, ok := GetDimString(value)
	if ok {
		if verbose {
			log.Printf("[serial] Setting dim mode: '%v'", value)
		}
		result := performExchange(port, command)
		return isOK(result)
	}
	if verbose {
		log.Printf("[serial] Invalid dim argument: '%v'", value)
	}

	return false
}

// Send data and receive data in return.
func performExchange(port serial.Port, data string) string {
	Write(port, []byte(data))
	value := Read(port)
	return value
}

// Return a bool value indicating if data is (a) OK.
func isOK(data string) bool {
	// TODO: fix proper string checks
	status := false
	if len(data) > 0 {
		if data == "+!#" {
			status = true
		}
	}
	return status
}

// Return a string ready to send to serial device, for changing LED color.
func GetColorString(value string, cfg *settings.Config) (string, bool) {
	command := ""
	ok := false
	if !strings.Contains(value, ":") {
		// Convert text into rgb string
		for _, cfg_color := range cfg.Colors {
			if cfg_color.Name == value {
				value = cfg_color.Value
			}
		}
	}
	red, green, blue := func(s []string) (string, string, string) {
		return s[0], s[1], s[2]
	}(strings.Split(value, ":"))
	red_i, _ := strconv.Atoi(red)
	green_i, _ := strconv.Atoi(green)
	blue_i, _ := strconv.Atoi(blue)

	if red_i >= 0 && red_i <= 255 && green_i >= 0 && green_i <= 255 && blue_i >= 0 && blue_i <= 255 {
		command = fmt.Sprintf("+l%03d%03d%03d#", red_i, green_i, blue_i)
		ok = true
	}

	return command, ok
}

// Return a string ready to send to serial device, for setting LED brightness.
func GetBrightnessString(value string, cfg *settings.Config) (string, bool) {
	command := ""
	ok := false
	value_i, _ := strconv.Atoi(value)

	if value_i >= cfg.BrightnessMin && value_i <= cfg.BrightnessMax {
		command = fmt.Sprintf("+b%03d#", value_i)
		ok = true
	}

	return command, ok
}

// Return a string ready to send to serial device, for setting display mode.
func GetModeString(value string, cfg *settings.Config) (string, bool) {
	command := ""
	ok := false
	value_i, _ := strconv.Atoi(value)
	mode_max := len(cfg.Modes)

	if value_i >= 1 && value_i <= mode_max {
		command = fmt.Sprintf("+m%03d#", value_i)
		ok = true
	}

	return command, ok
}

// Return a string ready to send to serial device, for dimming LED.
func GetDimString(value string) (string, bool) {
	command := ""
	ok := false
	value_i, _ := strconv.Atoi(value)

	if value_i >= 0 || value_i <= 1 {
		command = fmt.Sprintf("+d%1d#", value_i)
		ok = true
	}

	return command, ok
}

// Get and parse hardware information, updating settings if needed and returning information.
func updateHardwareInformation(port serial.Port, cfg *settings.Config) string {
	hardware_info := SendRequest(port)
	hardware_state := *settings.ParseHardwareInformation(hardware_info)

	if hardware_state.Brightness_min >= 0 {
		cfg.BrightnessMin = hardware_state.Brightness_min
	}
	if hardware_state.Brightness_max > 0 {
		cfg.BrightnessMax = hardware_state.Brightness_max
	}

	return hardware_info
}

// Runner parts //

// Start runner for serial communication.
// Handles reading and writing in different goroutines.
func Runner(cfg *settings.Config, verbosity bool, cmd_channel chan string, rsp_channel chan string) {
	verbose = verbosity
	port := Init(cfg)
	_ = SendPing(port)
	_ = updateHardwareInformation(port, cfg)

	go commandHandler(port, cmd_channel, rsp_channel, cfg)

	select {}
}

func commandHandler(port serial.Port, cmd_channel chan string, rsp_channel chan string, cfg *settings.Config) {
	pinger := watchdog.NewChannelTimer(time.Second * 30) // make sure watchdog send ping every 58 if no other commands have been sent

	for {
		select {
		case <-pinger.Channel():
			_ = SendPing(port)
			pinger.Kick()
		case data := <-cmd_channel:
			if len(data) > 0 {
				parts := strings.Split(data, ":")
				status := false
				rsp_msg := "nok"

				if parts[0] == "color" {
					status = SendColorCommand(port, parts[1], cfg)
				} else if parts[0] == "brightness" {
					status = SendBrightnessCommand(port, parts[1], cfg)
				} else if parts[0] == "mode" {
					mode_str := ""
					for _, cfg_mode := range cfg.Modes {
						if cfg_mode.Name == parts[1] {
							mode_str = strconv.Itoa(cfg_mode.Value)
						}
					}
					status = SendModeCommand(port, mode_str, cfg)
				} else if parts[0] == "dim" {
					dim_str := "0"
					if parts[1] == "true" {
						dim_str = "1"
					}
					status = SendDimCommand(port, dim_str)
				} else if parts[0] == "information" {
					if parts[1] == "all" {
						hw_info := updateHardwareInformation(port, cfg)
						rsp_channel <- hw_info
						pinger.Kick()
						break // skip kicking and sending reply later on
					}
				}

				if status {
					rsp_msg = "ok"
				}
				rsp_channel <- rsp_msg
				pinger.Kick()
			}
		}
	}
}
