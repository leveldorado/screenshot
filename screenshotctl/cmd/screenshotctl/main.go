package main

import (
	"fmt"
	"log"
	"os"

	"github.com/leveldorado/screenshot/screenshotctl"
)

func main() {
	opt, err := screenshotctl.ParseFlags(os.Args)
	if err != nil {
		log.Println(err.Error())
		os.Exit(1)
	}
	fmt.Println("URLS", opt.URLs, os.Args)
	cm := screenshotctl.NewCommand(opt.Backend)
	urls, err := opt.ExtractURLs()
	if err != nil {
		log.Println(err.Error())
		os.Exit(1)
	}
	if err := cm.MakeScreenShotsAndPrintResult(urls); err != nil {
		log.Println(err.Error())
		os.Exit(1)
	}
}
