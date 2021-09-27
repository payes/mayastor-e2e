package uninstall

import (
	"github.com/onsi/gomega"
	"mayastor-e2e/common"
	"mayastor-e2e/common/custom_resources"
	"mayastor-e2e/common/e2e_config"
	"mayastor-e2e/common/k8stest"
	"mayastor-e2e/common/locations"
	"mayastor-e2e/install"
	"os/exec"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"time"
)

const UninstallSuiteName = "Basic Teardown Suite"
const MCPUninstallSuiteName = "Basic Teardown Suite (mayastor control plane)"

// Helper for deleting mayastor CRDs
func deleteCRD(crdName string) {
	cmd := exec.Command("kubectl", "delete", "crd", crdName)
	_ = cmd.Run()
}

// Create mayastor namespace
func deleteNamespace() {
	cmd := exec.Command("kubectl", "delete", "namespace", common.NSMayastor())
	out, err := cmd.CombinedOutput()
	gomega.Expect(err).ToNot(gomega.HaveOccurred(), "%s", out)
}

// Teardown mayastor on the cluster under test.
// We deliberately call out to kubectl, rather than constructing the client-go
// objects, so that we can verify the local deploy yaml files are correct.
func TeardownMayastor() {
	var cleaned bool
	cleanup := e2e_config.GetConfig().Uninstall.Cleanup != 0

	log.Log.Info("Settings:", "cleanup", cleanup)
	if cleanup {
		cleaned = k8stest.CleanUp()
	} else {

		found, err := k8stest.CheckForTestPods()
		if err != nil {
			log.Log.Info("Failed to checking for test pods.", "error", err)
		} else {
			gomega.Expect(found).To(gomega.BeFalse(), "Application pods were found, none expected.")
		}

		found, err = k8stest.CheckForPVCs()
		if err != nil {
			log.Log.Info("Failed to check for PVCs", "error", err)
		}
		gomega.Expect(found).To(gomega.BeFalse(), "PersistentVolumeClaims were found, none expected.")

		found, err = k8stest.CheckForPVs()
		if err != nil {
			log.Log.Info("Failed to check PVs", "error", err)
		}
		gomega.Expect(found).To(gomega.BeFalse(), "PersistentVolumes were found, none expected.")

		if !k8stest.IsControlPlaneMcp() {
			found, err = custom_resources.CheckForMsVols()
			if err != nil {
				log.Log.Info("Failed to check MSVs", "error", err)
			}
			gomega.Expect(found).To(gomega.BeFalse(), "Mayastor volume CRDs were found, none expected.")
		}

		err = custom_resources.CheckAllMsPoolsAreOnline()
		gomega.Expect(err).ToNot(gomega.HaveOccurred())

	}

	poolsDeleted := k8stest.DeleteAllPools()
	gomega.Expect(poolsDeleted).To(gomega.BeTrue())

	log.Log.Info("Cleanup done, Uninstalling mayastor")
	install.GenerateMayastorYamlFiles()
	if k8stest.IsControlPlaneMcp() {
		install.GenerateMCPYamlFiles()
	}
	yamlsDir := locations.GetGeneratedYamlsDir()

	// Deletes can stall indefinitely, try to mitigate this
	// by running the deletes on different threads
	if k8stest.IsControlPlaneMcp() {
		go k8stest.KubeCtlDeleteYaml("rest-service.yaml", yamlsDir)
		go k8stest.KubeCtlDeleteYaml("rest-deployment.yaml", yamlsDir)
		go k8stest.KubeCtlDeleteYaml("msp-deployment.yaml", yamlsDir)
		go k8stest.KubeCtlDeleteYaml("csi-deployment.yaml", yamlsDir)
		go k8stest.KubeCtlDeleteYaml("core-agents-deployment.yaml", yamlsDir)
	} else {
		go k8stest.KubeCtlDeleteYaml("moac-deployment.yaml", yamlsDir)
		// todo: these should ideally be deleted after MOAC and mayastor have gone
		// because their removal may cause MOAC and mayastor to block
	}

	go k8stest.KubeCtlDeleteYaml("csi-daemonset.yaml", yamlsDir)
	go k8stest.KubeCtlDeleteYaml("mayastor-daemonset.yaml", yamlsDir)
	go k8stest.KubeCtlDeleteYaml("nats-deployment.yaml", yamlsDir)
	go k8stest.KubeCtlDeleteYaml("etcd", yamlsDir)

	{
		const timeOutSecs = 240
		const sleepSecs = 10
		maxIters := (timeOutSecs + sleepSecs - 1) / sleepSecs
		numMayastorPods := k8stest.MayastorUndeletedPodCount()
		if numMayastorPods != 0 {
			log.Log.Info("Waiting for Mayastor pods to be deleted",
				"timeout", timeOutSecs)
		}
		for iter := 0; iter < maxIters && numMayastorPods != 0; iter++ {
			log.Log.Info("\tWaiting ",
				"seconds", sleepSecs,
				"numMayastorPods", numMayastorPods,
				"iter", iter)
			numMayastorPods = k8stest.MayastorUndeletedPodCount()
			time.Sleep(sleepSecs * time.Second)
		}
	}

	// Delete all CRDs in the mayastor namespace.
	// FIXME: should we? For now yes if nothing else this ensures consistency
	// when deploying and redeploying Mayastor with MOAC and Mayastor with control plane
	// on the same cluster.
	if k8stest.IsControlPlaneMcp() {
		k8stest.KubeCtlDeleteYaml("operator-rbac.yaml", yamlsDir)
	} else {
		k8stest.KubeCtlDeleteYaml("moac-rbac.yaml", yamlsDir)
		deleteCRD("mayastornodes.openebs.io")
		deleteCRD("mayastorvolumes.openebs.io")
	}
	deleteCRD("mayastorpools.openebs.io")

	if cleanup {
		// Attempt to forcefully delete mayastor pods
		deleted, podCount, err := k8stest.ForceDeleteMayastorPods()
		gomega.Expect(err).ToNot(gomega.HaveOccurred(), "ForceDeleteMayastorPods failed %v", err)
		gomega.Expect(podCount).To(gomega.BeZero(), "All Mayastor pods have not been deleted")
		// Only delete the namespace if there are no pending resources
		// otherwise this hangs.
		deleteNamespace()
		if deleted {
			log.Log.Info("Mayastor pods were force deleted on cleanup!")
		}
		if cleaned {
			log.Log.Info("Application pods or volume resources were deleted on cleanup!")
		}
	} else {
		gomega.Expect(k8stest.MayastorUndeletedPodCount()).To(gomega.Equal(0), "All Mayastor pods were not removed on uninstall")
		// More verbose here as deleting the namespace is often where this
		// test hangs.
		log.Log.Info("Deleting the mayastor namespace")
		deleteNamespace()
		log.Log.Info("Deleted the mayastor namespace")
	}
}
