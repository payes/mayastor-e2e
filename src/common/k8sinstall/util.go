package k8sinstall

import (
	"fmt"
	"io/ioutil"
	"os/exec"
	"sort"
	"strings"
	"sync"
	"time"

	"mayastor-e2e/common"
	"mayastor-e2e/common/controlplane"
	"mayastor-e2e/common/custom_resources"
	"mayastor-e2e/common/e2e_config"
	"mayastor-e2e/common/k8stest"
	"mayastor-e2e/common/locations"

	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
)

const InstallSuiteName = "Basic Install Suite"
const InstallSuiteNameV1 = "Basic Install Suite (mayastor control plane)"
const UninstallSuiteName = "Basic Teardown Suite"
const UninstallSuiteNameV1 = "Basic Teardown Suite (mayastor control plane)"

const MCPLogLevel = "debug"

// mayastor install yaml files names
var mayastorYamlFiles = []string{
	"etcd",
	"nats-deployment.yaml",
	"csi-daemonset.yaml",
	"mayastor-daemonset.yaml",
}

// postFixInstallation post fix installation yaml files for coverage and debug if required
func postFixInstallation(yamlsDir string) error {

	buildInfoFile, err := locations.GetBuildInfoFile()
	if err != nil {
		logf.Log.Info("no postfix: failed to find build information")
		// didn't find the build flags files then there is nothing to do
		return nil
	}

	// Post process installation yaml files for coverage and debug builds.
	bashCmd := fmt.Sprintf("%s/postfix-install.py %s -b %s",
		locations.GetE2EScriptsPath(),
		yamlsDir,
		buildInfoFile,
	)

	logf.Log.Info("About to execute", "command", bashCmd)
	cmd := exec.Command("bash", "-c", bashCmd)
	out, err := cmd.CombinedOutput()
	if err != nil {
		logf.Log.Info("Error", "output", string(out))
		return err
	} else {
		logf.Log.Info(string(out))
	}
	return nil
}

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
		for _, node := range nodeLocs {
			if node.MasterNode {
				masterNode = node.NodeName
			}
			if !node.MayastorNode {
				continue
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
		logf.Log.Info("Error", "output", string(out))
		return err
	}

	err = postFixInstallation(locations.GetGeneratedYamlsDir())
	return err
}

func WaitForPoolCrd() bool {
	const timoSleepSecs = 5
	const timoSecs = 60
	for ix := 0; ix < timoSecs; ix += timoSleepSecs {
		time.Sleep(time.Second * timoSleepSecs)
		_, err := k8stest.ListMsPools()
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

func GenerateControlPlaneYamlFiles() error {
	e2eCfg := e2e_config.GetConfig()

	if controlplane.MajorVersion() == 1 {
		registryDirective := ""
		if len(e2eCfg.Registry) != 0 {
			registryDirective = fmt.Sprintf(" -r '%s'", e2eCfg.Registry)
		}

		imageTag := e2eCfg.ImageTag

		bashCmd := fmt.Sprintf(
			"%s/generate-deploy-yamls.sh -o %s -t '%s' -s mayastorCP.logLevel=%s %s test",
			locations.GetControlPlaneScriptsDir(),
			locations.GetControlPlaneGeneratedYamlsDir(),
			imageTag, MCPLogLevel, registryDirective,
		)
		logf.Log.Info("About to execute", "command", bashCmd)
		cmd := exec.Command("bash", "-c", bashCmd)
		out, err := cmd.CombinedOutput()
		if err != nil {
			logf.Log.Info("Error", "output", string(out))
			return err
		}

		err = postFixInstallation(locations.GetControlPlaneGeneratedYamlsDir())
		return err
	}

	return fmt.Errorf("unsupported control plane version %d\n", controlplane.MajorVersion())
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
	var errs common.ErrorAccumulator
	yamlFiles, err := getCPYamlFiles()
	if err != nil {
		return err
	}
	yamlsDir := locations.GetControlPlaneGeneratedYamlsDir()
	for _, yf := range yamlFiles {
		err := k8stest.KubeCtlApplyYaml(yf, yamlsDir)
		if err != nil {
			errs.Accumulate(fmt.Errorf("failed to apply yaml file %s , yaml path: %s, error: %v", yf, yamlsDir, err))
		}
	}
	return errs.GetError()
}

func uninstallControlPlane() error {
	errors := make(chan error, 1)
	var wg sync.WaitGroup
	var errs common.ErrorAccumulator
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
		wg.Add(1)
		// Deletes can stall indefinitely, try to mitigate this
		// by running the deletes on different threads
		// this may break the reversal of uninstalls
		// for now this is good enough
		go func(yf string) {
			time.Sleep(time.Second * 1)
			defer wg.Done()
			err = k8stest.KubeCtlDeleteYaml(yf, yamlsDir)
			if err != nil {
				errors <- err
			}
		}(yf)
	}
	wg.Wait()
	close(errors)
	for err := range errors {
		errs.Accumulate(err)
	}
	return errs.GetError()
}

// Install mayastor on the cluster under test.
// We deliberately call out to kubectl, rather than constructing the client-go
// objects, so that we can verify the local deploy yaml files are correct.
func InstallMayastor() error {
	var err error
	var errs common.ErrorAccumulator
	e2eCfg := e2e_config.GetConfig()

	if len(e2eCfg.ImageTag) == 0 {
		return fmt.Errorf("mayastor image tag not defined")
	}
	if len(e2eCfg.PoolDevice) == 0 {
		return fmt.Errorf("configuration error pools are not defined.")
	}

	cpVersion := controlplane.Version()
	logf.Log.Info("Control Plane", "version", cpVersion)

	mayastorNodes, err := k8stest.GetMayastorNodeNames()
	if err != nil {
		return err
	}

	numMayastorInstances := len(mayastorNodes)
	if numMayastorInstances == 0 {
		return fmt.Errorf("number of mayastor nodes is 0")
	}

	logf.Log.Info("Install", "tag", e2eCfg.ImageTag, "registry", e2eCfg.Registry, "# of mayastor instances", numMayastorInstances)

	err = GenerateControlPlaneYamlFiles()
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

	for _, yaml := range mayastorYamlFiles {
		err = k8stest.KubeCtlApplyYaml(yaml, yamlsDir)
		if err != nil {
			errs.Accumulate(err)
		}
	}
	if errs.GetError() != nil {
		return errs.GetError()
	}

	if controlplane.MajorVersion() == 1 {
		err = installControlPlane()
		if err != nil {
			return err
		}
	} else {
		return fmt.Errorf("unsupported control plane version %d/n", controlplane.MajorVersion())
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
		return fmt.Errorf("mayastor control plane installation is not ready")
	}

	// wait for mayastor node to be ready
	nodeReady, err := k8stest.MayastorNodeReady(5, 180)
	if err != nil {
		return err
	}
	if !nodeReady {
		return fmt.Errorf("all mayastor node are not ready")
	}

	crdReady := WaitForPoolCrd()
	if !crdReady {
		return fmt.Errorf("mayastor pool CRD not defined?")
	}

	// Now create configured pools on all nodes.
	err = k8stest.CreateConfiguredPools()
	if err != nil {
		return err
	}

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
	errors := make(chan error, 1)
	var wg sync.WaitGroup
	var errs common.ErrorAccumulator
	cleanup := e2e_config.GetConfig().Uninstall.Cleanup != 0
	err := k8stest.CheckAndSetControlPlane()
	if err != nil {
		return err
	}
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
	err = GenerateMayastorYamlFiles()
	if err != nil {
		return err
	}
	if controlplane.MajorVersion() == 1 {
		err = GenerateControlPlaneYamlFiles()
	} else {
		return fmt.Errorf("unsupported control plane version %d/n", controlplane.MajorVersion())
	}
	if err != nil {
		return err
	}
	yamlsDir := locations.GetGeneratedYamlsDir()

	// Deletes can stall indefinitely, try to mitigate this
	// by running the deletes on different threads
	if controlplane.MajorVersion() == 1 {
		err := uninstallControlPlane()
		if err != nil {
			return err
		}
	} else {
		return fmt.Errorf("unsupported control plane version %d/n", controlplane.MajorVersion())
	}

	// delete in reverse order
	for i := len(mayastorYamlFiles) - 1; i >= 0; i-- {
		yf := mayastorYamlFiles[i]
		wg.Add(1)
		go func() {
			time.Sleep(time.Second * 1)
			defer wg.Done()
			err = k8stest.KubeCtlDeleteYaml(yf, yamlsDir)
			if err != nil {
				errors <- err
			}
		}()
	}
	wg.Wait()
	close(errors)
	for err := range errors {
		errs.Accumulate(err)
	}
	if errs.GetError() != nil {
		return errs.GetError()
	}

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
