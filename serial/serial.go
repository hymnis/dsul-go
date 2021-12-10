// DSUL - Disturb State USB Light : Serial module.
package serial

import (
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"go.bug.st/serial"

	"github.com/hymnis/dsul-go/settings"
)

// Initialize serial port.
func Init(cfg *settings.Config) serial.Port {
	mode := &serial.Mode{
		BaudRate: cfg.Serial.Baudrate,
		Parity:   serial.NoParity,
		DataBits: 8,
		StopBits: serial.OneStopBit,
	}
	port, err := serial.Open(cfg.Serial.Port, mode) // TODO: get device from settings and not hardcoded
	if err != nil {
		log.Fatalf("Failed to open serial port: %v", err.Error())
	}
	port.SetReadTimeout(2 * time.Second)
	log.Printf("Serial port set: %d_N81", cfg.Serial.Baudrate)
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

		log.Printf("Serial Receiving: %s\n%v\n", buff, buff) // DEBUG
		for _, ch := range buff {
			output += string(ch)
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
	log.Printf("Serial Sending: %s %v\n", data, data)
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
		log.Printf("Setting color: '%v'", value)
		result := performExchange(port, command)
		return isOK(result)
	}
	log.Printf("Invalid color argument: '%v'", value)
	return false
}

// Send command to set the brightness.
// ??? - brightness value, 0-255
func SendBrightnessCommand(port serial.Port, value string, cfg *settings.Config) bool {
	command, ok := GetBrightnessString(value, cfg)
	if ok {
		log.Printf("Setting brightness: '%v'", value)
		result := performExchange(port, command)
		return isOK(result)
	}
	log.Printf("Invalid brightness argument: '%v'", value)
	return false
}

// Send command to set the mode.
// ??? - mode value, 0-4
func SendModeCommand(port serial.Port, value string, cfg *settings.Config) bool {
	command, ok := GetModeString(value, cfg)
	if ok {
		log.Printf("Setting mode: '%v'", value)
		result := performExchange(port, command)
		return isOK(result)
	}
	log.Printf("Invalid mode argument: '%v'", value)
	return false
}

// Send command to set the dim mode.
// 0 = No dimming (turn off dim mode)
// 1 = Dimming (turn on dim mode)
func SendDimCommand(port serial.Port, value string) bool {
	command, ok := GetDimString(value)
	if ok {
		log.Printf("Setting dim mode: '%v'", value)
		result := performExchange(port, command)
		return isOK(result)
	}
	log.Printf("Invalid dim argument: '%v'", value)
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
	fmt.Printf("isOK input: '%v'\n", data) // DEBUG
	// TODO: fix proper string checks
	status := false
	if len(data) > 0 {
		if data == "+!#" {
			status = true
		}
	}
	fmt.Printf("isOK output: '%v'\n", status) // DEBUG
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
	fmt.Printf("mode_max: %d\n", mode_max) // DEBUG

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

// Runner parts //

// Start runner for serial communication.
// Handles reading and writing in different goroutines.
func Runner(cfg *settings.Config, cmd_channel chan string, rsp_channel chan string) {
	port := Init(cfg)
	time.Sleep(2 * time.Second) // let device boot properly
	_ = SendPing(port)

	// Use either the goroutines below or 'Send*Command' functions, not both at the same time.
	// 'reader' and 'writer' uses the serial port directly.
	//serial_to := make(chan string)
	//go reader(port)
	//go writer(port, serial_to)
	go commandHandler(port, cmd_channel, rsp_channel, cfg)

	// <DEBUG
	// Validate data before sending commands via 'serial_to'.
	// Using 'Get*String' functions the input i validated and the 'ok' return value indicates if value is acceptable.
	//bri, bri_ok := GetBrightnessString("50", 10, 120)
	//if bri_ok {
	//	serial_to <- bri // "+b050#", brightness: 50
	//}
	//time.Sleep(2 * time.Second)
	//led, led_ok := GetColorString("255:000:000")
	//if led_ok {
	//	serial_to <- led // "+l255000000#", led: red
	//}
	//mode, mode_ok := GetModeString("2", 4)
	//if mode_ok {
	//	serial_to <- mode // "+m002#", mode: 2
	//}
	//time.Sleep(3 * time.Second)
	//mode, mode_ok = GetModeString("1", 4)
	//if mode_ok {
	//	serial_to <- mode // "+m001#", mode: 1
	//}
	//led, led_ok = GetColorString("000:255:000")
	//if led_ok {
	//	serial_to <- led // "+l000255000#", led: green
	//}
	// DEBUG>

	select {}
}

// Reads and reacts to data from the serial port.
// DEPRECATED
func reader(port serial.Port) {
	for {
		// Get serial input (blocks until data is received or timeout)
		input := Read(port)
		if len(input) > 0 {
			fmt.Printf("Serial Reader got: %s\n", input) // DEBUG
			if input == "-?#" {
				log.Print("Serial Response: Ping")
				_ = SendPing(port)
			} else if input == "-!#" {
				log.Print("Serial Response: Resend/Request")
			} else if input == "+?#" {
				log.Print("Serial Response: Unknown/Error")
			} else if input == "+!#" {
				log.Print("Serial Response: OK")
			} else {
				fmt.Println("not sure what we got") // DEBUG
			}
		}
	}
}

// Writes data to the serial port when received from channel.
// DEPRECATED
func writer(port serial.Port, serial_to chan string) {
	for {
		select {
		case data := <-serial_to:
			fmt.Printf("Serial Writer got: %s\n", data) // DEBUG
			if len(data) > 0 {
				Write(port, []byte(data))
			}
		}
	}
}

func commandHandler(port serial.Port, cmd_channel chan string, rsp_channel chan string, cfg *settings.Config) {
	for {
		select {
		case data := <-cmd_channel:
			if len(data) > 0 {
				parts := strings.Split(data, ":")

				if parts[0] == "color" {
					_ = SendColorCommand(port, parts[1], cfg)
				} else if parts[0] == "brightness" {
					_ = SendBrightnessCommand(port, parts[1], cfg)
				} else if parts[0] == "mode" {
					mode_str := ""
					for _, cfg_mode := range cfg.Modes {
						if cfg_mode.Name == parts[1] {
							mode_str = strconv.Itoa(cfg_mode.Value)
						}
					}
					_ = SendModeCommand(port, mode_str, cfg)
				} else if parts[0] == "dim" {
					dim_str := "0"
					if parts[1] == "true" {
						dim_str = "1"
					}
					_ = SendDimCommand(port, dim_str)
				} else if parts[0] == "information" {
					if parts[1] == "all" {
						hw_info := SendRequest(port)
						rsp_channel <- hw_info
					}
				}
			}
		}
	}
}
