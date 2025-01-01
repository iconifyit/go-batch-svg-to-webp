package main

import (
	"fmt"
	"log"
	"os"
	"time"

	// Correctly import the uuid package

	ip "image-processor/src/image-processor"
)

// main is the entry point for the application.
func main() {
	if len(os.Args) < 2 {
		log.Fatal("Usage: go run main.go <config.yml>")
	}

	startTime := time.Now()

	configFile := os.Args[1]
	log.Printf("Using config file: %s", configFile)

	imageProcessor := ip.NewImageProcessor(configFile)
	fmt.Println(imageProcessor.Config)

	imageProcessor.Run()

	log.Printf("Started image processor at time: %s", startTime.String())
	log.Printf("Finished image processor at time: %s", time.Now().String())
	log.Printf("Elapsed time: %s", time.Since(time.Now()))
}
