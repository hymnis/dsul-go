// DSUL - Disturb State USB Light : Serial module.
package serial

import (
    "fmt"
    "log"
    "regexp"
    "strconv"
    "strings"
    "time"

    "go.bug.st/serial"
)

type hardware struct {
    version string
    leds int
    brightness_min int
    brightness_max int
    current_color string
    current_brightness int
    current_mode int
    current_dim int
}

// Initialize serial port.
func Init() serial.Port {
    mode := &serial.Mode{
        BaudRate: 38400,
        Parity:   serial.NoParity,
        DataBits: 8,
        StopBits: serial.OneStopBit,
    }
    port, err := serial.Open("/dev/ttyUSB0", mode) // TODO: get device from settings and not hardcoded
    if err != nil {
        log.Fatal(err)
    }
    port.SetReadTimeout(2 * time.Second)
    log.Print("Serial port set: 38400_N81")
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

        log.Printf("Receiving: %s\n%v\n", buff, buff)
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
    log.Printf("Sending: %s %v\n", data, data)
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
    result := PerformExchange(port, "-?#")
    ok := IsOK(result)
    return ok
}

// Send request for information to device.
func SendRequest(port serial.Port) string {
    result := PerformExchange(port, "-!#")
    return result
}

// Send command to set the color.
// ???:???:??? - red:green:blue values
func SendColorCommand(port serial.Port, value string) bool {
    red, green, blue := func(s []string) (string, string, string) {
        return s[0], s[1], s[2]
    }(strings.Split(value, ":"))
    log.Printf("Setting color: '%s,%s,%s'", red, green, blue)
    red_i, _ := strconv.Atoi(red)
    green_i, _ := strconv.Atoi(green)
    blue_i, _ := strconv.Atoi(blue)
    command := fmt.Sprintf("+l%03d%03d%03d#", red_i, green_i, blue_i)
    result := PerformExchange(port, command)
    ok := IsOK(result)
    return ok
}

// Send command to set the brightness.
// ??? - brightness value
func SendBrightnessCommand(port serial.Port, value string, brightness_min int, brightness_max int) bool {
    value_i, _ := strconv.Atoi(value)
    if value_i >= brightness_min && value_i <= brightness_max {
        log.Printf("Setting brightness: '%v'", value)
        command := fmt.Sprintf("+b%03d#", value_i)
        result := PerformExchange(port, command)
        ok := IsOK(result)
        return ok
    }
    log.Printf("Invalid brightness argument: '%v'", value)
    return false
}

// Send command to set the mode.
// ??? - mode value
func SendModeCommand(port serial.Port, value string) bool {
    value_i, _ := strconv.Atoi(value)
    if value_i >= 1 && value_i <= 4 { // TODO: fix proper check instead of hardcoded values
        log.Printf("Setting mode: '%v'", value)
        command := fmt.Sprintf("+m%03d#", value_i)
        result := PerformExchange(port, command)
        ok := IsOK(result)
        return ok
    }
    log.Printf("Invalid mode argument: '%v'", value)
    return false
}

// Send command to set the dim mode.
// 0 = No dimming (turn off dim mode)
// 1 = Dimming (turn on dim mode)
func SendDimCommand(port serial.Port, value string) bool {
    value_i, _ := strconv.Atoi(value)
    if value_i >= 0 || value_i <= 1 {
        log.Printf("Setting dim mode: '%v'", value)
        command := fmt.Sprintf("+d%1d#", value_i)
        result := PerformExchange(port, command)
        ok := IsOK(result)
        return ok
    }
    log.Printf("Invalid dim argument: '%v'", value)
    return false
}

// Send data and receive data in return.
func PerformExchange(port serial.Port, data string) string {
    Write(port, []byte(data))
    value := Read(port)
    return value
}

// Return a bool value indicating if data is (a) OK.
func IsOK(data string) bool {
    // TODO: fix proper string checks
    status := false
    if len(data) > 0 {
        if data == "+!#" {
            status = true
        }
    }
    return status
}

// Parse information from harware and return a hardware struct.
func ParseInformation(info string) *hardware {
    hardware_info := hardware{}
    ve_match := regexp.MustCompile(`v(\d{3})\.(\d{3}).(\d{3})`).FindStringSubmatch(info)
    ll_match := regexp.MustCompile(`ll(\d{3})`).FindStringSubmatch(info)
    lb_match := regexp.MustCompile(`lb(\d{3}):(\d{3})`).FindStringSubmatch(info)
    cc_match := regexp.MustCompile(`cc(\d{2})(\d{2})(\d{2})`).FindStringSubmatch(info)
    cb_match := regexp.MustCompile(`cb(\d{3})`).FindStringSubmatch(info)
    cm_match := regexp.MustCompile(`cm(\d{3})`).FindStringSubmatch(info)
    cd_match := regexp.MustCompile(`cd(\d{1})`).FindStringSubmatch(info)
    if len(ve_match) > 0 {
        version_major_i, _ := strconv.Atoi(ve_match[1])
        version_minor_i, _ := strconv.Atoi(ve_match[2])
        version_patch_i, _ := strconv.Atoi(ve_match[3])
        hardware_info.version = fmt.Sprintf("%d.%d.%d", version_major_i, version_minor_i, version_patch_i)
    }
    if len(ll_match) > 0 {
        leds_i, _ := strconv.Atoi(ll_match[1])
        hardware_info.leds = leds_i
    }
    if len(lb_match) > 0 {
        brightness_min_i, _ := strconv.Atoi(lb_match[1])
        brightness_max_i, _ := strconv.Atoi(lb_match[2])
        hardware_info.brightness_min = brightness_min_i
        hardware_info.brightness_max = brightness_max_i
    }
    // TODO: fix current_color regexp matching. always returns empty string
    fmt.Printf("cc_match: %v\n", cc_match) // DEBUG
    if len(cc_match) > 0 {
        current_color_red_i, _ := strconv.Atoi(cc_match[1])
        current_color_green_i, _ := strconv.Atoi(cc_match[2])
        current_color_blue_i, _ := strconv.Atoi(cc_match[3])
        hardware_info.current_color = fmt.Sprintf("%d:%d:%d", current_color_red_i, current_color_green_i, current_color_blue_i)
    }
    if len(cb_match) > 0 {
        current_brightness_i, _ := strconv.Atoi(cb_match[1])
        hardware_info.current_brightness = current_brightness_i
    }
    if len(cm_match) > 0 {
        current_mode_i, _ := strconv.Atoi(cm_match[1])
        hardware_info.current_mode = current_mode_i
    }
    if len(cd_match) > 0 {
        current_dim_i, _ := strconv.Atoi(cd_match[1])
        hardware_info.current_dim = current_dim_i
    }
    return &hardware_info
}

// Runner parts //

// Start runner for serial communication.
// Handles reading and writing in different goroutines.
func Runner() {
    port := Init()
    time.Sleep(2 * time.Second) // let device boot properly
    _ = SendPing(port)
    hw_info := SendRequest(port)
    hardware_state := ParseInformation(hw_info)
    fmt.Printf("HW Info: %v\n", hw_info) // DEBUG
    fmt.Printf("Hardware: %v\n", hardware_state) // DEBUG
    serial_to := make(chan string)

    go Reader(port)
    go Writer(port, serial_to)

    // <DEBUG
    time.Sleep(3 * time.Second)
    serial_to <- "b050"
    SendColorCommand(port, "255:000:000")
    SendModeCommand(port, "2")
    time.Sleep(3 * time.Second)
    SendModeCommand(port, "1")
    SendColorCommand(port, "000:255:000")
    // DEBUG>

    select {}
}

// Reads and reacts to data from the serial port.
func Reader(port serial.Port) {
    for {
        // get serial input (blocks until data is received or timeout)
        input := Read(port)
        if len(input) > 0 {
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
func Writer(port serial.Port, serial_to chan string) {
    for {
        select {
        case data := <-serial_to:
            // parse data
            //...
            fmt.Printf("Writer got: %s\n", data)
        }
    }
}
