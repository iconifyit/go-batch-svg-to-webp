package main

import (
	"log"
	"time"

	fn "github.com/iconifyit/go-batch-svg-to-webp/src/common"
	ip "github.com/iconifyit/go-batch-svg-to-webp/src/image-processor"

	"github.com/spf13/pflag"
)

func main() {
	// Define named arguments with both short and long names
	var contributor string
	var configFile string

	pflag.StringVarP(&contributor, "contributor", "c", "", "Contributor name")
	pflag.StringVarP(&configFile, "file", "f", "", "Path to the configuration file")

	// Parse the command-line arguments
	pflag.Parse()

	// Validate required arguments
	if configFile == "" || contributor == "" {
		log.Fatal("Usage: go run main.go -f | --file <config.yml> [-c | --contributor <contributor>]")
	}

	startTime := time.Now()

	log.Printf("Using config file: %s", configFile)
	log.Printf("Processing for contributor: %s", contributor)

	// Initialize the ImageProcessor
	imageProcessor := ip.NewImageProcessor(contributor, configFile)

	// Print the loaded configuration
	// log.Printf("Config: %v", fn.ToJSON(imageProcessor.Config))
	log.Printf("ImageProcessor: %v", fn.ToJSON(imageProcessor))

	// Run the image processor
	imageProcessor.Run()

	// Log the timing details
	log.Println("------------------------------------------------------")
	log.Printf("Started image processor at time: %s", startTime.String())
	log.Printf("Finished image processor at time: %s", time.Now().String())
	log.Printf("Elapsed time: %s", time.Since(startTime))
}
