package main

import (
	"github.com/charmbracelet/log"

	"github.com/itspacchu/node-exporter-pusher/cmd"
	"github.com/itspacchu/node-exporter-pusher/constants"
)

func main() {
	constants.PrintTitle()
	if err := cmd.Run(); err != nil {
		log.Error(err.Error())
	}
}
