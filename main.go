package main

import (
	"context"
	"fmt"
	"log"
	"net/url"
	"os"
	"strings"

	"github.com/vmware/govmomi/property"
	"github.com/vmware/govmomi/session/cache"
	"github.com/vmware/govmomi/view"
	"github.com/vmware/govmomi/vim25"
	"github.com/vmware/govmomi/vim25/mo"

	"github.com/adam-power/wcp-prereq-checker/check"
)

const (
	envURL             = "GOVC_URL"
	envUsername        = "GOVC_USERNAME"
	envPassword        = "GOVC_PASSWORD"
	envInsecure        = "GOVC_INSECURE"
	envWorkloadCluster = "GOVC_CLUSTER"
)

func main() {
	vCenterURL := os.Getenv(envURL)
	vCenterUsername := os.Getenv(envUsername)
	vCenterPassword := os.Getenv(envPassword)
	vCenterCluster := os.Getenv(envWorkloadCluster)

	vCenterInsecure := false
	switch env := strings.ToLower(os.Getenv(envInsecure)); env {
	case "1", "true":
		vCenterInsecure = true
	}

	var err error
	ctx := context.Background()

	// Initialize vCenter URL object
	vcURL, err := url.Parse(vCenterURL)
	if err != nil {
		log.Fatalf("Failed to parse vCenter URL: %v", err)
	}
	vcURL.User = url.UserPassword(vCenterUsername, vCenterPassword)
	if vcURL.Path == "/" || vcURL.Path == "" {
		vcURL.Path = "/sdk"
	}

	vCenterSession := &cache.Session{
		URL:      vcURL,
		Insecure: vCenterInsecure,
	}

	// vCenter login
	vc := new(vim25.Client)
	err = vCenterSession.Login(ctx, vc, nil)
	if err != nil {
		log.Fatalf("Failed to log in to vCenter: %v", err)
	}

	m := view.NewManager(vc)
	containerView, err := m.CreateContainerView(ctx, vc.ServiceContent.RootFolder, []string{"ComputeResource"}, true)
	if err != nil {
		log.Fatalf("Error creating container view: %v", err)
	}

	var clusters []mo.ComputeResource
	err = containerView.RetrieveWithFilter(ctx, []string{"ComputeResource"}, []string{"name", "host"}, &clusters, property.Filter{"name": vCenterCluster})
	if err != nil {
		log.Fatalf("Error retrieving compute cluster \"%s\": %v", vCenterCluster, err)
	}

	checkRunner := check.NewRunner()
	checkRunner.RunCheck(
		fmt.Sprintf("Cluster \"%s\" exists.", vCenterCluster),
		func() bool {
			if len(clusters) != 1 {
				return false
			} else {
				return true
			}
		},
	)
	checkRunner.RunCheck(
		"Workload cluster has at least 3 hosts.",
		func() bool {
			if len(clusters[0].Host) < 3 {
				return false
			} else {
				return true
			}
		},
	)

	checkRunner.Summary()
}
