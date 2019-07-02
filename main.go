package main

import (
	"github.com/kuuland/shino/internal"
	"os"
)

func main() {
	if os.Getenv("DRONE_REPO") != "" {
		internal.RunDrone()
	} else {
		internal.RunLocal()
	}
}
