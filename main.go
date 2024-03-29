package main

import (
	"sync"

	gsettings "github.com/enso-os/enso-gsettings/common"
)

func main() {
	gset := make(chan string, 1)
	xset := make(chan string, 1)
	var wg sync.WaitGroup

	wg.Add(1)
	go gsettings.PollXfconf(xset, &wg)
	wg.Add(1)
	go gsettings.PollgSettings(gset, &wg)

	wg.Wait()
}
