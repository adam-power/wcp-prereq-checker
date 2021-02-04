package main

import (
	"context"
	"log"
	"net/url"
	"os"
	"strings"

	"github.com/vmware/govmomi/property"
	"github.com/vmware/govmomi/session/cache"
	"github.com/vmware/govmomi/view"
	"github.com/vmware/govmomi/vim25"
	"github.com/vmware/govmomi/vim25/mo"
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

	if len(clusters) != 1 {
		log.Fatalf("Cluster \"%s\" not found.", vCenterCluster)
	} else if len(clusters[0].Host) < 3 {
		log.Fatalf("Cluster \"%s\" must have at least 3 hosts, but has %d hosts.\n", vCenterCluster, len(clusters[0].Host))
	} else {
		log.Printf("Cluster \"%s\" has at least 3 hosts.\n", vCenterCluster)
	}
}
