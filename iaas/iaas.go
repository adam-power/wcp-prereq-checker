package iaas

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

func RunIaaSChecks() {
	var err error
	ctx := context.Background()

	vc, err := vlogin(ctx)
	if err != nil {
		log.Fatalf("%v\n", err)
	}

	vCenterCluster := os.Getenv(envWorkloadCluster)

	m := view.NewManager(vc)
	containerView, err := m.CreateContainerView(ctx, vc.ServiceContent.RootFolder, []string{"ComputeResource"}, true)
	if err != nil {
		log.Fatalf("Error creating vCenter ContainerView: %v\n", err)
	}

	// Verify whether Workload Cluster exists
	var clusterExists bool
	var clusters []mo.ComputeResource
	err = containerView.RetrieveWithFilter(ctx, []string{"ComputeResource"}, []string{"name", "host"}, &clusters, property.Filter{"name": vCenterCluster})
	if err != nil {
		clusterExists = false
	} else if len(clusters) != 1 {
		clusterExists = false
	} else {
		clusterExists = true
	}
	check.RegisterResult(
		fmt.Sprintf("Cluster \"%s\" exists.", vCenterCluster),
		clusterExists,
		err,
	)

	// Verify that Workload Cluster has sufficient hosts
	var sufficientHosts bool
	if len(clusters[0].Host) < 3 {
		sufficientHosts = false
	} else {
		sufficientHosts = true
	}
	check.RegisterResult(
		"Workload cluster has at least 3 hosts.",
		sufficientHosts,
		nil,
	)
}

func vlogin(ctx context.Context) (*vim25.Client, error) {
	vc := new(vim25.Client)

	vCenterURL := os.Getenv(envURL)
	vCenterUsername := os.Getenv(envUsername)
	vCenterPassword := os.Getenv(envPassword)

	vCenterInsecure := false
	switch env := strings.ToLower(os.Getenv(envInsecure)); env {
	case "1", "true":
		vCenterInsecure = true
	}

	vcURL, err := url.Parse(vCenterURL)
	if err != nil {
		return nil, fmt.Errorf("Failed to parse vCenter URL: %v", err)
	}
	vcURL.User = url.UserPassword(vCenterUsername, vCenterPassword)
	if vcURL.Path == "/" || vcURL.Path == "" {
		vcURL.Path = "/sdk"
	}

	vCenterSession := &cache.Session{
		URL:      vcURL,
		Insecure: vCenterInsecure,
	}

	err = vCenterSession.Login(ctx, vc, nil)
	if err != nil {
		return nil, fmt.Errorf("Failed to log in to vCenter: %v", err)
	}

	return vc, nil
}
