package openstack

import (
	"fmt"
	"github.com/mitchellh/multistep"
	"github.com/mitchellh/packer/packer"
	"github.com/rackspace/gophercloud"
	"log"
	"time"
)

type stepCreateImage struct{}

func (s *stepCreateImage) Run(state map[string]interface{}) multistep.StepAction {
	csp := state["csp"].(gophercloud.CloudServersProvider)
	config := state["config"].(config)
	server := state["server"].(*gophercloud.Server)
	ui := state["ui"].(packer.Ui)

	// Create the image
	ui.Say(fmt.Sprintf("Creating the image: %s", config.ImageName))
	createOpts := gophercloud.CreateImage{
		Name: config.ImageName,
	}
	imageId, err := csp.CreateImage(server.Id, createOpts)
	if err != nil {
		err := fmt.Errorf("Error creating image: %s", err)
		state["error"] = err
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	// Set the Image ID in the state
	ui.Say(fmt.Sprintf("Image: %s", imageId))
	state["image"] = imageId

	// Wait for the image to become ready
	ui.Say("Waiting for image to become ready...")
	if err := WaitForImage(csp, imageId); err != nil {
		err := fmt.Errorf("Error waiting for image: %s", err)
		state["error"] = err
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	return multistep.ActionContinue
}

func (s *stepCreateImage) Cleanup(map[string]interface{}) {
	// No cleanup...
}

// WaitForImage waits for the given Image ID to become ready.
func WaitForImage(csp gophercloud.CloudServersProvider, imageId string) error {
	for {
		image, err := csp.ImageById(imageId)
		if err != nil {
			return err
		}

		if image.Status == "ACTIVE" {
			return nil
		}

		log.Printf("Waiting for image creation status: %s (%d%%)", image.Status, image.Progress)
		time.Sleep(2 * time.Second)
	}
}
