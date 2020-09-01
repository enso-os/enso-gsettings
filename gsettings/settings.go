package gsettings

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"os/exec"
	"sync"
)

type Settings struct {
	Xkb string `json:"xkb"`
}

// settings := glib.SettingsNew("org.gnome.desktop.input-sources")
// settings.ListChildren()
// sources := settings.GetString("sources")
// fmt.Print(sources)

func Init() {
	var xfSettings string

	out, err := getSettings("get", "org.gnome.desktop.input-sources", "sources")
	if err != nil {
		log.Printf("Error on get settings: %s", err.Error())
		return
	}

	settings, err := convertVariantToJson(out)
	if err != nil {
		log.Printf("Error on unmarshal of sources: %s", err.Error())
		return
	}

	for _, set := range settings {
		xfSettings += set.Xkb
		xfSettings += ","
	}

	_, err = setXfSettings("-c", "keyboard-layout", "-np", "/Default/XkbDisable", "-s", "false")

	if err != nil {
		log.Printf("Error on get settings: %s", err.Error())
		return
	}

	_, err = setXfSettings("-c", "keyboard-layout", "-np", "/Default/XkbLayout", "-s", xfSettings)

	if err != nil {
		log.Printf("Error on get settings: %s", err.Error())
		return
	}

	return
}

func compareSettings(gsettings string, xfconf string) bool {
	value := false

	return value
}

func PollgSettings(channel chan string, wg *sync.WaitGroup) {
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
			fmt.Printf("taken %d chars %s", n, string(buff[:n]))
		}
	}

	if err := cmd.Wait(); err != nil {
		log.Fatal(err)
	}
}

func PollXfconf(channel chan string, wg *sync.WaitGroup) {
	cmd := exec.Command("xfconf-query", "-c", "keyboard-layout", "-m", "-v")
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
			fmt.Printf("taken %d chars %s", n, string(buff[:n]))
		}
	}

	if err := cmd.Wait(); err != nil {
		log.Fatal(err)
	}
}

func getSettings(cmd ...string) ([]byte, error) {
	c := exec.Command("gsettings", cmd...)
	return c.CombinedOutput()
}

func setXfSettings(cmd ...string) ([]byte, error) {
	c := exec.Command("xfconf-query", cmd...)
	return c.CombinedOutput()
}

func convertJsonToVariant(variant []byte) (string, error) {
	var str string

	// hacky way of converting the a(ss) to a json array
	variant = bytes.ReplaceAll(variant, []byte(`"`), []byte(`'`))
	variant = bytes.ReplaceAll(variant, []byte(`':`), []byte(`',`))
	variant = bytes.ReplaceAll(variant, []byte(`{`), []byte(`(`))
	variant = bytes.ReplaceAll(variant, []byte(`}`), []byte(`)`))

	str = string(variant)

	return str, nil
}

func convertVariantToJson(variant []byte) ([]Settings, error) {
	var settings []Settings

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
