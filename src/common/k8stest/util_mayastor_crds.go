package k8stest

// Utility functions for Mayastor CRDs
import (
	. "github.com/onsi/gomega"
	"mayastor-e2e/common/custom_resources"
	v1alpha1Api "mayastor-e2e/common/custom_resources/api/types/v1alpha1"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
)

// GetMSV Get pointer to a mayastor volume custom resource
// function asserts if the volume CR is not found.
func GetMSV(uuid string) *v1alpha1Api.MayastorVolume {
	msv, err := custom_resources.GetMsVol(uuid)
	if err != nil {
		logf.Log.Info("GetMSV", "error", err)
		return nil
	}

	// pending means still being created
	if msv.Status.State == "pending" {
		return nil
	}

	// Note: msVol.Node can be unassigned here if the volume is not mounted
	Expect(msv.Status.State).NotTo(Equal(""))
	Expect(len(msv.Status.Replicas)).To(BeNumerically(">", 0))
	return msv
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
