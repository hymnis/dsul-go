// DSUL - Disturb State USB Light : Settings module.
package settings

import (
	"fmt"
	"log"
	"os"
	"regexp"
	"runtime"
	"strconv"

	"github.com/tucnak/store"
)

var (
	applicationName = "dsul"
	configName      = "dsul.yml"
)

type Color struct {
	Name  string
	Value string
}
type Mode struct {
	Name  string
	Value int
}
type Serial struct {
	Port     string
	Baudrate int
}
type Config struct {
	Colors        []Color
	Modes         []Mode
	BrightnessMin int
	BrightnessMax int
	Serial        Serial
}

type Hardware struct {
	Version            string
	Leds               int
	Brightness_min     int
	Brightness_max     int
	Current_color      string
	Current_brightness int
	Current_mode       int
	Current_dim        int
}

func GetSettings() *Config {
	cfg := getDefaults()

	store.Init(applicationName)
	guaranteeConfigFile()
	if err := store.Load(configName, &cfg); err != nil {
		log.Println("Failed to load the DSUL configuration: ", err)
		return nil
	}

	return &cfg
}

func SaveSettings(cfg *Config) {
	store.Init(applicationName)
	guaranteeConfigFile()
	if err := store.Save(configName, &cfg); err != nil {
		log.Println("Failed to save the DSUL configuration: ", err)
	}
}

func getDefaults() Config {
	config := Config{
		Colors: []Color{
			Color{"black", "0:0:0"},
			Color{"white", "255:255:200"},
			Color{"warmwhite", "255:230:200"},
			Color{"red", "255:0:0"},
			Color{"green", "0:255:0"},
			Color{"blue", "0:0:255"},
			Color{"cyan", "0:255:255"},
			Color{"purple", "255:0:200"},
			Color{"magenta", "255:0:50"},
			Color{"yellow", "255:90:0"},
			Color{"orange", "255:20:0"},
		},
		Modes: []Mode{
			Mode{"solid", 1},
			Mode{"blink", 2},
			Mode{"flash", 3},
			Mode{"pulse", 4},
		},
		BrightnessMin: 0,
		BrightnessMax: 150,
		Serial:        Serial{"/dev/ttyUSB0", 38400},
	}
	return config
}

func guaranteeConfigFile() {
	_, err := os.OpenFile(buildPath("dsul.yml"), os.O_RDONLY|os.O_CREATE, 0644)
	if err != nil {
		log.Fatal(err)
	}
}

func buildPath(path string) string {
	if runtime.GOOS == "windows" {
		return fmt.Sprintf("%s\\%s\\%s", os.Getenv("APPDATA"),
			applicationName,
			path)
	}

	var unixConfigDir string
	if xdg := os.Getenv("XDG_CONFIG_HOME"); xdg != "" {
		unixConfigDir = xdg
	} else {
		unixConfigDir = os.Getenv("HOME") + "/.config"
	}

	return fmt.Sprintf("%s/%s/%s", unixConfigDir,
		applicationName,
		path)
}

// Parse information from harware and return a hardware struct.
func ParseHardwareInformation(info string) *Hardware {
	hardware_info := Hardware{}
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
		hardware_info.Version = fmt.Sprintf("%d.%d.%d", version_major_i, version_minor_i, version_patch_i)
	}
	if len(ll_match) > 0 {
		leds_i, _ := strconv.Atoi(ll_match[1])
		hardware_info.Leds = leds_i
	}
	if len(lb_match) > 0 {
		brightness_min_i, _ := strconv.Atoi(lb_match[1])
		brightness_max_i, _ := strconv.Atoi(lb_match[2])
		hardware_info.Brightness_min = brightness_min_i
		hardware_info.Brightness_max = brightness_max_i
	}
	// TODO: fix current_color regexp matching. always returns empty string
	fmt.Printf("cc_match: %v\n", cc_match) // DEBUG
	if len(cc_match) > 0 {
		current_color_red_i, _ := strconv.Atoi(cc_match[1])
		current_color_green_i, _ := strconv.Atoi(cc_match[2])
		current_color_blue_i, _ := strconv.Atoi(cc_match[3])
		hardware_info.Current_color = fmt.Sprintf("%d:%d:%d", current_color_red_i, current_color_green_i, current_color_blue_i)
	}
	if len(cb_match) > 0 {
		current_brightness_i, _ := strconv.Atoi(cb_match[1])
		hardware_info.Current_brightness = current_brightness_i
	}
	if len(cm_match) > 0 {
		current_mode_i, _ := strconv.Atoi(cm_match[1])
		hardware_info.Current_mode = current_mode_i
	}
	if len(cd_match) > 0 {
		current_dim_i, _ := strconv.Atoi(cd_match[1])
		hardware_info.Current_dim = current_dim_i
	}
	return &hardware_info
}
