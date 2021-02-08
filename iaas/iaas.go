package iaas

import (
	"context"
	"fmt"
	"log"
	"net/url"
	"os"
	"reflect"
	"strings"

	"github.com/vmware/govmomi/session/cache"
	"github.com/vmware/govmomi/view"
	"github.com/vmware/govmomi/vim25"
	"github.com/vmware/govmomi/vim25/mo"
	"github.com/vmware/govmomi/vim25/types"

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

	// Retrieve Workload Cluster
	var clusters []mo.ClusterComputeResource
	err = containerView.Retrieve(ctx, []string{"ClusterComputeResource"}, []string{"name", "host", "configuration"}, &clusters)
	if err != nil {
		log.Fatalf("Error retrieving vCenter clusters: %v\n", err)
	}

	var workloadCluster mo.ClusterComputeResource
	for _, cluster := range clusters {
		if cluster.Name == vCenterCluster {
			workloadCluster = cluster
			break
		}
	}

	checkClusterExists(workloadCluster)
	checkHostCount(workloadCluster)
	checkVSphereHA(workloadCluster)
	checkVSphereDRS(workloadCluster)
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

func checkClusterExists(workloadCluster mo.ClusterComputeResource) {
	var clusterExists bool

	if reflect.ValueOf(workloadCluster).IsZero() {
		clusterExists = false
	} else {
		clusterExists = true
	}
	check.RegisterResult(
		"Workload cluster exists.",
		clusterExists,
		nil,
	)
}

func checkHostCount(workloadCluster mo.ClusterComputeResource) {
	var sufficientHosts bool
	var err error

	if reflect.ValueOf(workloadCluster).IsZero() {
		sufficientHosts = false
		err = fmt.Errorf("Workload cluster does not exist.")
	} else if len(workloadCluster.Host) < 3 {
		sufficientHosts = false
	} else {
		sufficientHosts = true
	}
	check.RegisterResult(
		"Workload cluster has at least 3 hosts.",
		sufficientHosts,
		err,
	)
}

func checkVSphereHA(workloadCluster mo.ClusterComputeResource) {
	var haEnabled bool
	var err error

	if reflect.ValueOf(workloadCluster).IsZero() {
		haEnabled = false
		err = fmt.Errorf("Workload cluster does not exist.")
	} else {
		haEnabled = *workloadCluster.Configuration.DasConfig.Enabled
	}

	check.RegisterResult(
		"Workload cluster has vSphere HA enabled.",
		haEnabled,
		err,
	)
}

func checkVSphereDRS(workloadCluster mo.ClusterComputeResource) {
	var drsEnabled bool
	var err error

	if reflect.ValueOf(workloadCluster).IsZero() {
		drsEnabled = false
		err = fmt.Errorf("Workload cluster does not exist.")
	} else if workloadCluster.Configuration.DrsConfig.DefaultVmBehavior == types.DrsBehaviorFullyAutomated {
		drsEnabled = true
	} else {
		drsEnabled = false
	}

	check.RegisterResult(
		"Workload cluster has DRS enabled in Fully Automated mode.",
		drsEnabled,
		err,
	)
}
