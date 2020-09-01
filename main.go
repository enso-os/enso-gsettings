package main

import (
	"sync"

	"github.com/nick92/settings/gsettings"
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
