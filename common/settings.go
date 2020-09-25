package gsettings

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"os/exec"
	"reflect"
	"strings"
	"sync"
)

type Settings struct {
	Xkb string `json:"xkb"`
}

// settings := glib.SettingsNew("org.gnome.desktop.input-sources")
// settings.ListChildren()
// sources := settings.GetString("sources")
// fmt.Print(sources)

func PollgSettings(channel chan string, wg *sync.WaitGroup) {
	log.Println("Polling gsettings")
	cmd := exec.Command("gsettings", "monitor", "org.gnome.desktop.input-sources", "sources")

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		log.Fatal(err)
	}

	defer wg.Done()

	if err := cmd.Start(); err != nil {
		log.Fatal(err)
	}

	// stdout, err := cmd.StdoutPipe()
	buff := make([]byte, 100)
	var n int

	for err == nil {
		n, err = stdout.Read(buff)

		if n > 0 {
			log.Println(string(buff))
			xconf, err := getXfSettings()

			if err != nil {
				log.Fatal(err)
			}

			gsets, err := getGSettings()

			if err != nil {
				log.Fatal(err)
			}

			if err == nil && !reflect.DeepEqual(xconf, gsets) {
				log.Println("Not equal so setting xfconf ..")
				setXfSettings(gsets)
				log.Println("Xfconf set ..")
			}
		}
	}

	if err := cmd.Wait(); err != nil {
		log.Fatal(err)
	}
}

func PollXfconf(channel chan string, wg *sync.WaitGroup) {
	log.Println("Polling xfconf setting")
	cmd := exec.Command("xfconf-query", "-c", "keyboard-layout", "-m", "-v")

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		log.Fatal(err)
	}

	defer wg.Done()

	if err := cmd.Start(); err != nil {
		log.Fatal(err)
	}

	buff := make([]byte, 100)
	var n int

	for err == nil {
		n, err = stdout.Read(buff)

		if n > 0 {
			log.Println(string(buff))
			xconf, err := getXfSettings()

			if err != nil {
				log.Fatal(err)
			}

			gsets, err := getGSettings()

			if err != nil {
				log.Fatal(err)
			}

			if err == nil && !reflect.DeepEqual(xconf, gsets) {
				log.Println("Not equal so setting gsettings ..")
				setGSettings(xconf)
				log.Println("gettings set ..")
			}
		}
	}

	if err := cmd.Wait(); err != nil {
		log.Fatal(err)
	}
}

func getGSettings() ([]Settings, error) {
	c := exec.Command("gsettings", "get", "org.gnome.desktop.input-sources", "sources")
	out, err := c.CombinedOutput()

	if err != nil {
		log.Println(err)
	}

	return convertVariantToJson(out)
}

func getXfSettings() ([]Settings, error) {
	c := exec.Command("xfconf-query", "-c", "keyboard-layout", "-p", "/Default/XkbLayout")
	out, err := c.CombinedOutput()

	if err != nil {
		// if the setting doesn't exist, create it
		cmd := exec.Command("xfconf-query", "-c", "keyboard-layout", "-p", "/Default/XkbLayout", "-n", "-t", "string", "-s", "")
		cmd.CombinedOutput()
		log.Println(err)

		settings := []Settings{}

		return settings, nil
	}

	fmt.Println(string(out))

	return convertXfConfToSettings(out)
}

func setXfSettings(settings []Settings) ([]byte, error) {
	var xfSettings string

	for _, set := range settings {
		xfSettings += set.Xkb
		xfSettings += ","
	}

	c := exec.Command("xfconf-query", "-c", "keyboard-layout", "-np", "/Default/XkbLayout", "-s", xfSettings)
	return c.CombinedOutput()
}

func setGSettings(settings []Settings) ([]byte, error) {
	var variant string

	for _, set := range settings {
		if set.Xkb != "" {
			variant += fmt.Sprintf("('xkb', '%s'),", set.Xkb)
		}
	}
	if variant != "" {
		variant = fmt.Sprintf("[%s]", variant[0:len(variant)-1])
	}

	c := exec.Command("gsettings", "set", "org.gnome.desktop.input-sources", "sources", variant)
	return c.CombinedOutput()
}

func convertVariantToJson(variant []byte) ([]Settings, error) {
	var settings []Settings

	if bytes.Contains(variant, []byte(`@a(ss)`)) {
		return settings, nil
	}

	// hacky way of converting the a(ss) to a json array
	variant = bytes.ReplaceAll(variant, []byte(`',`), []byte(`':`))
	variant = bytes.ReplaceAll(variant, []byte(`'`), []byte(`"`))
	variant = bytes.ReplaceAll(variant, []byte(`(`), []byte(`{`))
	variant = bytes.ReplaceAll(variant, []byte(`)`), []byte(`}`))

	err := json.Unmarshal(variant, &settings)

	if err != nil {
		return settings, err
	}

	return settings, nil
}

func convertXfConfToSettings(out []byte) ([]Settings, error) {
	var settings []Settings

	str := string(out)

	str = strings.Trim(str, "\n")

	xkbs := strings.Split(str, ",")

	for _, xkb := range xkbs {
		if xkb != "" {
			setting := Settings{
				Xkb: xkb,
			}

			settings = append(settings, setting)
		}
	}

	return settings, nil
}
