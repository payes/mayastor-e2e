package k8stest

// Utility functions for Mayastor CRDs
import (
	"fmt"
	"mayastor-e2e/common/custom_resources"
	v1alpha1Api "mayastor-e2e/common/custom_resources/api/types/v1alpha1"

	. "github.com/onsi/gomega"
)

// GetMSV Get pointer to a mayastor volume custom resource
// returns nil and no error if the msv is in pending state.
func GetMSV(uuid string) (*v1alpha1Api.MayastorVolume, error) {
	msv, err := custom_resources.GetMsVol(uuid)
	if err != nil {
		return nil, fmt.Errorf("GetMSV: %v", err)
	}

	// pending means still being created
	if msv.Status.State == "pending" {
		return nil, nil
	}

	//logf.Log.Info("GetMSV", "msv", msv)
	// Note: msVol.Node can be unassigned here if the volume is not mounted
	if msv.Status.State == "" {
		return nil, fmt.Errorf("GetMSV: state not defined, got msv.Status=\"%v\"", msv.Status)
	}

	if len(msv.Status.Replicas) < 1 {
		return nil, fmt.Errorf("GetMSV: msv.Status.Replicas=\"%v\"", msv.Status.Replicas)
	}
	return msv, nil
}

// GetMsvNodes Retrieve the nexus node hosting the Mayastor Volume,
// and the names of the replica nodes
// function asserts if the volume CR is not found.
func GetMsvNodes(uuid string) (string, []string) {
	msv, err := custom_resources.GetMsVol(uuid)
	Expect(err).ToNot(HaveOccurred())
	node := msv.Status.Nexus.Node
	replicas := make([]string, len(msv.Status.Replicas))
	for ix, r := range msv.Status.Replicas {
		replicas[ix] = r.Node
	}
	return node, replicas
}
