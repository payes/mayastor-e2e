package k8sinstall

import (
	"fmt"
	"io/ioutil"
	"os/exec"
	"sort"
	"strings"
	"time"

	"mayastor-e2e/common"
	"mayastor-e2e/common/custom_resources"
	"mayastor-e2e/common/e2e_config"
	"mayastor-e2e/common/k8stest"
	"mayastor-e2e/common/locations"

	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
)

const InstallSuiteName = "Basic Install Suite"
const MCPInstallSuiteName = "Basic Install Suite (mayastor control plane)"
const UninstallSuiteName = "Basic Teardown Suite"
const MCPUninstallSuiteName = "Basic Teardown Suite (mayastor control plane)"

func GenerateMayastorYamlFiles() error {
	e2eCfg := e2e_config.GetConfig()

	coresDirective := ""
	if e2eCfg.Cores != 0 {
		coresDirective = fmt.Sprintf("%s -c %d", coresDirective, e2eCfg.Cores)
	}

	nodeLocs, err := k8stest.GetNodeLocs()
	if err != nil {
		return fmt.Errorf("GetNodeLocs failed %v", err)
	}
	poolDirectives := ""
	masterNode := ""
	if len(e2eCfg.PoolDevice) != 0 {
		poolDevice := e2eCfg.PoolDevice
		for _, node := range nodeLocs {
			if node.MasterNode {
				masterNode = node.NodeName
			}
			if !node.MayastorNode {
				continue
			}
			if !common.IsControlPlaneMcp() {
				poolDirectives += fmt.Sprintf(" -p '%s,%s'", node.NodeName, poolDevice)
			}
		}
	}

	registryDirective := ""
	if len(e2eCfg.Registry) != 0 {
		registryDirective = fmt.Sprintf(" -r '%s'", e2eCfg.Registry)
	}

	imageTag := e2eCfg.ImageTag

	etcdOptions := "etcd.replicaCount=1,etcd.nodeSelector=kubernetes.io/hostname: " + masterNode + ",etcd.tolerations=- key: node-role.kubernetes.io/master"
	bashCmd := fmt.Sprintf(
		"%s/generate-deploy-yamls.sh -s '%s' -o %s -t '%s' %s %s %s test",
		locations.GetMayastorScriptsDir(),
		etcdOptions,
		locations.GetGeneratedYamlsDir(),
		imageTag, registryDirective, coresDirective, poolDirectives,
	)
	logf.Log.Info("About to execute", "command", bashCmd)
	cmd := exec.Command("bash", "-c", bashCmd)
	out, err := cmd.CombinedOutput()
	if err != nil {
		logf.Log.Info("Error", "output", out)
	}
	return err
}

func WaitForPoolCrd() bool {
	const timoSleepSecs = 5
	const timoSecs = 60
	for ix := 0; ix < timoSecs; ix += timoSleepSecs {
		time.Sleep(time.Second * timoSleepSecs)
		_, err := custom_resources.ListMsPools()
		if err != nil {
			logf.Log.Info("WaitForPoolCrd", "error", err)
			if k8serrors.IsNotFound(err) {
				logf.Log.Info("WaitForPoolCrd, error := IsNotFound")
			} else {
				logf.Log.Info("", "Error", err)
				return false
			}
		} else {
			logf.Log.Info("WaitForPoolCrd, complete", "time", timoSleepSecs*ix)
			return true
		}
	}
	return false
}

func GenerateMCPYamlFiles() error {
	e2eCfg := e2e_config.GetConfig()

	if common.IsControlPlaneMcp() {
		registryDirective := ""
		if len(e2eCfg.Registry) != 0 {
			registryDirective = fmt.Sprintf(" -r '%s'", e2eCfg.Registry)
		}

		imageTag := e2eCfg.ImageTag

		bashCmd := fmt.Sprintf(
			"%s/generate-deploy-yamls.sh -o %s -t '%s' %s test",
			locations.GetMCPScriptsDir(),
			locations.GetControlPlaneGeneratedYamlsDir(),
			imageTag, registryDirective,
		)
		logf.Log.Info("About to execute", "command", bashCmd)
		cmd := exec.Command("bash", "-c", bashCmd)
		out, err := cmd.CombinedOutput()
		if err != nil {
			logf.Log.Info("Error", "output", out)
			return err
		}
	}
	// nothing to do for MOAC
	return nil
}

type cpYamlFiles []string

func (s cpYamlFiles) Len() int {
	return len(s)
}

func (s cpYamlFiles) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

func (s cpYamlFiles) weight(i int) int {
	if strings.HasSuffix(s[i], "rbac.yaml") {
		return 0
	}
	return 1
}

func (s cpYamlFiles) Less(i, j int) bool {
	return s.weight(i) < s.weight(j)
}

func getCPYamlFiles() ([]string, error) {
	yamlsDir := locations.GetControlPlaneGeneratedYamlsDir()
	files, err := ioutil.ReadDir(yamlsDir)
	if err != nil {
		return []string{}, err
	}
	var yamlFiles []string
	for _, file := range files {
		if !file.IsDir() {
			fName := file.Name()
			if strings.HasSuffix(fName, ".yaml") {
				yamlFiles = append(yamlFiles, fName)
			}
		}
	}
	sort.Strings(yamlFiles)
	sort.Sort(cpYamlFiles(yamlFiles))
	return yamlFiles, nil
}

func installControlPlane() error {
	yamlFiles, err := getCPYamlFiles()
	if err != nil {
		return err
	}
	yamlsDir := locations.GetControlPlaneGeneratedYamlsDir()
	for _, yf := range yamlFiles {
		k8stest.KubeCtlApplyYaml(yf, yamlsDir)
	}
	return nil
}

func uninstallControlPlane() error {
	yamlFiles, err := getCPYamlFiles()
	if err != nil {
		return err
	}
	// reverse the list of files for uninstall
	for i, j := 0, len(yamlFiles)-1; i < j; i, j = i+1, j-1 {
		yamlFiles[i], yamlFiles[j] = yamlFiles[j], yamlFiles[i]
	}
	yamlsDir := locations.GetControlPlaneGeneratedYamlsDir()
	for _, yf := range yamlFiles {
		// Deletes can stall indefinitely, try to mitigate this
		// by running the deletes on different threads
		// this may break the reversal of uninstalls
		// for now this is good enough
		go k8stest.KubeCtlDeleteYaml(yf, yamlsDir)
		time.Sleep(time.Second * 1)
	}
	return nil
}

// Install mayastor on the cluster under test.
// We deliberately call out to kubectl, rather than constructing the client-go
// objects, so that we can verify the local deploy yaml files are correct.
func InstallMayastor() error {
	var err error
	e2eCfg := e2e_config.GetConfig()

	if len(e2eCfg.ImageTag) == 0 {
		return fmt.Errorf("mayastor image tag not defined")
	}
	if len(e2eCfg.PoolDevice) == 0 {
		return fmt.Errorf("configuration error pools are not defined.")
	}

	mayastorNodes, err := k8stest.GetMayastorNodeNames()
	if err != nil {
		return err
	}

	numMayastorInstances := len(mayastorNodes)
	if numMayastorInstances == 0 {
		return fmt.Errorf("number of mayastor nodes is 0")
	}

	logf.Log.Info("Install", "tag", e2eCfg.ImageTag, "registry", e2eCfg.Registry, "# of mayastor instances", numMayastorInstances)

	err = GenerateMCPYamlFiles()
	if err != nil {
		return err
	}
	err = GenerateMayastorYamlFiles()
	if err != nil {
		return err
	}
	yamlsDir := locations.GetGeneratedYamlsDir()
	logf.Log.Info("", "yamlsDir", yamlsDir)

	k8stest.EnsureE2EAgent()

	err = k8stest.MkNamespace(common.NSMayastor())
	if err != nil {
		return err
	}

	k8stest.KubeCtlApplyYaml("etcd", yamlsDir)
	k8stest.KubeCtlApplyYaml("nats-deployment.yaml", yamlsDir)
	k8stest.KubeCtlApplyYaml("csi-daemonset.yaml", yamlsDir)
	k8stest.KubeCtlApplyYaml("mayastor-daemonset.yaml", yamlsDir)

	if common.IsControlPlaneMcp() {
		err = installControlPlane()
		if err != nil {
			return err
		}
	} else {
		k8stest.KubeCtlApplyYaml("moac-rbac.yaml", yamlsDir)
		k8stest.KubeCtlApplyYaml("moac-deployment.yaml", yamlsDir)
	}

	ready, err := k8stest.MayastorReady(2, 540)
	if err != nil {
		return err
	}
	if !ready {
		return fmt.Errorf("mayastor installation is not ready")
	}

	ready = k8stest.ControlPlaneReady(10, 180)
	if !ready {
		return fmt.Errorf("mayastor control plane/MOAC installation is not ready")
	}

	crdReady := WaitForPoolCrd()
	if !crdReady {
		return fmt.Errorf("mayastor pool CRD not defined?")
	}

	// Now create configured pools on all nodes.
	k8stest.CreateConfiguredPools()

	// Wait for pools to be online
	const timoSecs = 240
	const timoSleepSecs = 10
	for ix := 0; ix < timoSecs/timoSleepSecs; ix++ {
		time.Sleep(timoSleepSecs * time.Second)
		err = custom_resources.CheckAllMsPoolsAreOnline()
		if err == nil {
			break
		}
	}
	if err != nil {
		return fmt.Errorf("One or more pools are offline %v", err)
	}

	// Mayastor has been installed and is now ready for use.
	return nil
}

// Helper for deleting mayastor CRDs
func deleteCRD(crdName string) {
	cmd := exec.Command("kubectl", "delete", "crd", crdName)
	_ = cmd.Run()
}

// Create mayastor namespace
func deleteNamespace() error {
	cmd := exec.Command("kubectl", "delete", "namespace", common.NSMayastor())
	out, err := cmd.CombinedOutput()
	if err != nil {
		logf.Log.Info("Error", "output", out)
	}
	return err
}

// TeardownMayastor tear down mayastor installation on the cluster under test.
// We deliberately call out to kubectl, rather than constructing the client-go
// objects, so that we can verify the local deploy yaml files are correct.
func TeardownMayastor() error {
	var cleaned bool
	cleanup := e2e_config.GetConfig().Uninstall.Cleanup != 0

	logf.Log.Info("Settings:", "cleanup", cleanup)
	if cleanup {
		cleaned = k8stest.CleanUp()
	} else {

		found, err := k8stest.CheckForTestPods()
		if err != nil {
			logf.Log.Info("Failed to checking for test pods.", "error", err)
		} else {
			if found {
				return fmt.Errorf("application pods were found, none expected")
			}
		}

		found, err = k8stest.CheckForPVCs()
		if err != nil {
			logf.Log.Info("Failed to check for PVCs", "error", err)
		}
		if found {
			return fmt.Errorf(" PersistentVolumeClaims were found, none expected")
		}

		found, err = k8stest.CheckForPVs()
		if err != nil {
			logf.Log.Info("Failed to check PVs", "error", err)
		}
		if found {
			return fmt.Errorf(" PersistentVolumes were found, none expected")
		}

		if !common.IsControlPlaneMcp() {
			found, err = k8stest.CheckForMsvs()
			if err != nil {
				logf.Log.Info("Failed to check MSVs", "error", err)
			}
			if found {
				return fmt.Errorf(" Mayastor volume CRDs were found, none expected")
			}
		}

		err = custom_resources.CheckAllMsPoolsAreOnline()
		if err != nil {
			return err
		}
	}

	poolsDeleted := k8stest.DeleteAllPools()
	if !poolsDeleted {
		return fmt.Errorf("failed to delete all pools")
	}

	logf.Log.Info("Cleanup done, Uninstalling mayastor")
	err := GenerateMayastorYamlFiles()
	if err != nil {
		return err
	}
	if common.IsControlPlaneMcp() {
		err = GenerateMCPYamlFiles()
	}
	if err != nil {
		return err
	}
	yamlsDir := locations.GetGeneratedYamlsDir()

	// Deletes can stall indefinitely, try to mitigate this
	// by running the deletes on different threads
	if common.IsControlPlaneMcp() {
		err := uninstallControlPlane()
		if err != nil {
			return err
		}
	} else {
		go k8stest.KubeCtlDeleteYaml("moac-deployment.yaml", yamlsDir)
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
			logf.Log.Info("Waiting for Mayastor pods to be deleted",
				"timeout", timeOutSecs)
		}
		for iter := 0; iter < maxIters && numMayastorPods != 0; iter++ {
			logf.Log.Info("\tWaiting ",
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
	if !common.IsControlPlaneMcp() {
		k8stest.KubeCtlDeleteYaml("moac-rbac.yaml", yamlsDir)
		deleteCRD("mayastornodes.openebs.io")
		deleteCRD("mayastorvolumes.openebs.io")
	}
	deleteCRD("mayastorpools.openebs.io")

	if cleanup {
		// Attempt to forcefully delete mayastor pods
		deleted, podCount, err := k8stest.ForceDeleteMayastorPods()
		if err != nil {
			return fmt.Errorf("ForceDeleteMayastorPods failed %v", err)
		}
		if podCount != 0 {
			return fmt.Errorf("all Mayastor pods have not been deleted")
		}
		if deleted {
			logf.Log.Info("Mayastor pods were force deleted on cleanup!")
		}
		if cleaned {
			logf.Log.Info("Application pods or volume resources were deleted on cleanup!")
		}
	} else {
		if k8stest.MayastorUndeletedPodCount() != 0 {
			return fmt.Errorf("all Mayastor pods were not removed on uninstall")
		}
	}

	// More verbose here as deleting the namespace is often where this
	// test hangs.
	logf.Log.Info("Deleting the mayastor namespace")
	// Only delete the namespace if there are no pending resources
	// otherwise this hangs.
	err = deleteNamespace()
	if err != nil {
		return err
	}
	logf.Log.Info("Deleted the mayastor namespace")
	return nil
}
