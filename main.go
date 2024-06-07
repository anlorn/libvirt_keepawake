package main

import (
	log "github.com/sirupsen/logrus"
	"libvirt_keepawake/cmd"
	"os"
)

func main() {
	// output to stdout instead of the default stderr
	log.SetOutput(os.Stdout)
	cmd.Execute()
}
