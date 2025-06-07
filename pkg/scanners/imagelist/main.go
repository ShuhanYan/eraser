package main

import (
	"context"
	"flag"
	"fmt"
	"os"

	"github.com/eraser-dev/eraser/api/unversioned"
	"github.com/eraser-dev/eraser/pkg/logger"
	template "github.com/eraser-dev/eraser/pkg/scanners/template"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
)

const (
	generalErr = 1
)

var (
	config = flag.String("config", "", "path to the configuration file")

	log = logf.Log.WithName("scanner").WithValues("provider", "imageListScanner")
)

func main() {
	flag.Parse()

	err := logger.Configure()
	if err != nil {
		fmt.Fprintf(os.Stderr, "error setting up logger: %s", err)
		os.Exit(generalErr)
	}

	log.Info("config", "config", *config)

	userConfig := *DefaultConfig()
	if *config != "" {
		var err error
		userConfig, err = loadConfig(*config)
		if err != nil {
			log.Error(err, "unable to read config")
			os.Exit(generalErr)
		}
	}

	log.V(1).Info("userConfig",
		"json", userConfig,
		"struct", fmt.Sprintf("%#v\n", userConfig),
	)

	// create image provider with custom values
	imageProvider := template.NewImageProvider(
		template.WithContext(context.Background()),
		template.WithMetrics(true),
		template.WithDeleteScanFailedImages(true),
		template.WithLogger(log),
	)

	// retrieve list of all non-running, non-excluded images from collector container
	allImages, err := imageProvider.ReceiveImages()
	if err != nil {
		log.Error(err, "unable to retrieve list of images from collector container")
		return
	}

	// scan images with custom scanner
	nonCompliant, failedImages := scan(allImages, userConfig.uncompliantImageList)

	// send images to eraser container
	if err := imageProvider.SendImages(nonCompliant, failedImages); err != nil {
		log.Error(err, "unable to send non-compliant images to eraser container")
		return
	}

	// complete scan
	if err := imageProvider.Finish(); err != nil {
		log.Error(err, "unable to complete scanner")
		return
	}
}

func scan(allImages []unversioned.Image, uncompliantImageList []string) ([]unversioned.Image, []unversioned.Image) {
	log := logf.Log.WithName("scanner").WithValues("provider", "imageListScanner")
	var nonCompliant []unversioned.Image
	imageSet := make(map[string]bool)
	for _, v := range uncompliantImageList {
		imageSet[v] = true
	}

	for _, img := range allImages {
		for _, name := range img.Names {
			if _, ok := imageSet[name]; ok {
				log.Info("Image is not compliant", "image", name)
				nonCompliant = append(nonCompliant, img)
				break
			}
		}
	}
	return nonCompliant, nil
}
