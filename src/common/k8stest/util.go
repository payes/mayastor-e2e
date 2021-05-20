package k8stest

import (
	"context"
	"errors"
	"fmt"
	"mayastor-e2e/common/crds"
	"mayastor-e2e/common/e2e_config"
	"os/exec"
	"regexp"
	"strconv"
	"time"

	"mayastor-e2e/common"
	"mayastor-e2e/common/mayastorclient"

	. "github.com/onsi/gomega"

	appsV1 "k8s.io/api/apps/v1"
	coreV1 "k8s.io/api/core/v1"
	v1 "k8s.io/api/core/v1"
	storagev1 "k8s.io/api/storage/v1"
	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
)

// Helper for passing yaml from the specified directory to kubectl
func KubeCtlApplyYaml(filename string, dir string) {
	cmd := exec.Command("kubectl", "apply", "-f", filename)
	cmd.Dir = dir
	logf.Log.Info("kubectl apply ", "yaml file", filename, "path", cmd.Dir)
	out, err := cmd.CombinedOutput()
	Expect(err).ToNot(HaveOccurred(), "%s", out)
}

// Helper for passing yaml from the specified directory to kubectl
func KubeCtlDeleteYaml(filename string, dir string) {
	cmd := exec.Command("kubectl", "delete", "-f", filename)
	cmd.Dir = dir
	logf.Log.Info("kubectl delete ", "yaml file", filename, "path", cmd.Dir)
	out, err := cmd.CombinedOutput()
	Expect(err).ToNot(HaveOccurred(), "%s", out)
}

// create a storage class
func MakeStorageClass(scName string, scReplicas int, protocol common.ShareProto, nameSpace string, bindingMode *storagev1.VolumeBindingMode) error {
	logf.Log.Info("Creating storage class", "name", scName, "replicas", scReplicas, "protocol", protocol, "bindingMode", bindingMode)
	createOpts := &storagev1.StorageClass{
		ObjectMeta: metaV1.ObjectMeta{
			Name:      scName,
			Namespace: nameSpace,
		},
		Provisioner: common.CSIProvisioner,
	}
	createOpts.Parameters = make(map[string]string)
	createOpts.Parameters["protocol"] = string(protocol)
	createOpts.Parameters["repl"] = strconv.Itoa(scReplicas)

	if bindingMode != nil {
		createOpts.VolumeBindingMode = bindingMode
	}

	ScApi := gTestEnv.KubeInt.StorageV1().StorageClasses
	_, createErr := ScApi().Create(context.TODO(), createOpts, metaV1.CreateOptions{})
	return createErr
}

// create a storage class with default volume binding mode i.e. not specified
func MkStorageClass(scName string, scReplicas int, protocol common.ShareProto, nameSpace string) error {
	return MakeStorageClass(scName, scReplicas, protocol, nameSpace, nil)
}

// remove a storage class
func RmStorageClass(scName string) error {
	logf.Log.Info("Deleting storage class", "name", scName)
	ScApi := gTestEnv.KubeInt.StorageV1().StorageClasses
	deleteErr := ScApi().Delete(context.TODO(), scName, metaV1.DeleteOptions{})
	return deleteErr
}

func CheckForStorageClasses() (bool, error) {
	found := false
	ScApi := gTestEnv.KubeInt.StorageV1().StorageClasses
	scs, err := ScApi().List(context.TODO(), metaV1.ListOptions{})
	for _, sc := range scs.Items {
		if sc.Provisioner == common.CSIProvisioner {
			found = true
		}
	}
	return found, err
}

func MkNamespace(nameSpace string) error {
	logf.Log.Info("Creating", "namespace", nameSpace)
	nsSpec := coreV1.Namespace{ObjectMeta: metaV1.ObjectMeta{Name: nameSpace}}
	_, err := gTestEnv.KubeInt.CoreV1().Namespaces().Create(context.TODO(), &nsSpec, metaV1.CreateOptions{})
	return err
}

//EnsureNamespace ensure that a namespace exists, creates namespace if not found.
func EnsureNamespace(nameSpace string) error {
	_, err := gTestEnv.KubeInt.CoreV1().Namespaces().Get(context.TODO(), nameSpace, metaV1.GetOptions{})
	if err == nil {
		return nil
	}
	return MkNamespace(nameSpace)
}

func RmNamespace(nameSpace string) error {
	logf.Log.Info("Deleting", "namespace", nameSpace)
	err := gTestEnv.KubeInt.CoreV1().Namespaces().Delete(context.TODO(), nameSpace, metaV1.DeleteOptions{})
	return err
}

// Add a node selector to the given pod definition
func ApplyNodeSelectorToPodObject(pod *coreV1.Pod, label string, value string) {
	if pod.Spec.NodeSelector == nil {
		pod.Spec.NodeSelector = make(map[string]string)
	}
	pod.Spec.NodeSelector[label] = value
}

// Add a node selector to the deployment spec and apply
func ApplyNodeSelectorToDeployment(deploymentName string, namespace string, label string, value string) {
	depApi := gTestEnv.KubeInt.AppsV1().Deployments
	deployment, err := depApi(namespace).Get(context.TODO(), deploymentName, metaV1.GetOptions{})
	Expect(err).ToNot(HaveOccurred())
	if deployment.Spec.Template.Spec.NodeSelector == nil {
		deployment.Spec.Template.Spec.NodeSelector = make(map[string]string)
	}
	deployment.Spec.Template.Spec.NodeSelector[label] = value
	_, err = depApi(namespace).Update(context.TODO(), deployment, metaV1.UpdateOptions{})
	Expect(err).ToNot(HaveOccurred())
}

// Remove all node selectors from the deployment spec and apply
func RemoveAllNodeSelectorsFromDeployment(deploymentName string, namespace string) {
	depApi := gTestEnv.KubeInt.AppsV1().Deployments
	deployment, err := depApi(namespace).Get(context.TODO(), deploymentName, metaV1.GetOptions{})
	Expect(err).ToNot(HaveOccurred())
	if deployment.Spec.Template.Spec.NodeSelector != nil {
		deployment.Spec.Template.Spec.NodeSelector = nil
		_, err = depApi(namespace).Update(context.TODO(), deployment, metaV1.UpdateOptions{})
	}
	Expect(err).ToNot(HaveOccurred())
}

// Adjust the number of replicas in the deployment
func SetDeploymentReplication(deploymentName string, namespace string, replicas *int32) {
	depAPI := gTestEnv.KubeInt.AppsV1().Deployments
	var err error

	// this is to cater for a race condition, occasionally seen,
	// when the deployment is changed between Get and Update
	for attempts := 0; attempts < 10; attempts++ {
		deployment, err := depAPI(namespace).Get(context.TODO(), deploymentName, metaV1.GetOptions{})
		Expect(err).ToNot(HaveOccurred())
		deployment.Spec.Replicas = replicas
		_, err = depAPI(namespace).Update(context.TODO(), deployment, metaV1.UpdateOptions{})
		if err == nil {
			break
		}
		logf.Log.Info("Re-trying update attempt due to error", "error", err)
		time.Sleep(1 * time.Second)
	}
	Expect(err).ToNot(HaveOccurred())
}

// Wait until all instances of the specified pod are absent from the given node
func WaitForPodAbsentFromNode(podNameRegexp string, namespace string, nodeName string, timeoutSeconds int) error {
	var validID = regexp.MustCompile(podNameRegexp)
	var podAbsent bool = false

	podApi := gTestEnv.KubeInt.CoreV1().Pods

	for i := 0; i < timeoutSeconds && !podAbsent; i++ {
		podAbsent = true
		time.Sleep(time.Second)
		podList, err := podApi(namespace).List(context.TODO(), metaV1.ListOptions{})
		if err != nil {
			return errors.New("failed to list pods")
		}
		for _, pod := range podList.Items {
			if pod.Spec.NodeName == nodeName {
				if validID.MatchString(pod.Name) {
					podAbsent = false
					break
				}
			}
		}
	}
	if !podAbsent {
		return errors.New("timed out waiting for pod")
	}
	return nil
}

// Get the execution status of the given pod, or nil if it does not exist
func getPodStatus(podNameRegexp string, namespace string, nodeName string) *v1.PodPhase {
	var validID = regexp.MustCompile(podNameRegexp)
	podAPI := gTestEnv.KubeInt.CoreV1().Pods
	podList, err := podAPI(namespace).List(context.TODO(), metaV1.ListOptions{})
	Expect(err).ToNot(HaveOccurred())
	for _, pod := range podList.Items {
		if pod.Spec.NodeName == nodeName && validID.MatchString(pod.Name) {
			return &pod.Status.Phase
		}
	}
	return nil // pod not found
}

// Wait until the instance of the specified pod is present and in the running
// state on the given node
func WaitForPodRunningOnNode(podNameRegexp string, namespace string, nodeName string, timeoutSeconds int) error {
	for i := 0; i < timeoutSeconds; i++ {
		stat := getPodStatus(podNameRegexp, namespace, nodeName)

		if stat != nil && *stat == v1.PodRunning {
			return nil
		}
		time.Sleep(1 * time.Second)
	}
	return errors.New("timed out waiting for pod to be running")
}

// Wait until the instance of the specified pod is absent or not in the running
// state on the given node
func WaitForPodNotRunningOnNode(podNameRegexp string, namespace string, nodeName string, timeoutSeconds int) error {
	for i := 0; i < timeoutSeconds; i++ {
		stat := getPodStatus(podNameRegexp, namespace, nodeName)

		if stat == nil || *stat != v1.PodRunning {
			return nil
		}
		time.Sleep(1 * time.Second)
	}
	return errors.New("timed out waiting for pod to stop running")
}

// returns true if the pod is present on the given node
func PodPresentOnNode(podNameRegexp string, namespace string, nodeName string) bool {
	var validID = regexp.MustCompile(podNameRegexp)
	podApi := gTestEnv.KubeInt.CoreV1().Pods
	podList, err := podApi(namespace).List(context.TODO(), metaV1.ListOptions{})
	Expect(err).ToNot(HaveOccurred())

	for _, pod := range podList.Items {
		if pod.Spec.NodeName == nodeName {
			if validID.MatchString(pod.Name) {
				return true
			}
		}
	}
	return false
}

func mayastorReadyPodCount() int {
	var mayastorDaemonSet appsV1.DaemonSet
	if gTestEnv.K8sClient.Get(context.TODO(), types.NamespacedName{Name: "mayastor", Namespace: common.NSMayastor()}, &mayastorDaemonSet) != nil {
		logf.Log.Info("Failed to get mayastor DaemonSet")
		return -1
	}
	logf.Log.Info("mayastor daemonset", "available instances", mayastorDaemonSet.Status.NumberAvailable)
	return int(mayastorDaemonSet.Status.NumberAvailable)
}

func mayastorCSIReadyPodCount() int {
	var mayastorCsiDaemonSet appsV1.DaemonSet
	if gTestEnv.K8sClient.Get(context.TODO(), types.NamespacedName{Name: "mayastor-csi", Namespace: common.NSMayastor()}, &mayastorCsiDaemonSet) != nil {
		logf.Log.Info("Failed to get mayastor-csi DaemonSet")
		return -1
	}
	logf.Log.Info("mayastor CSI daemonset", "available instances", mayastorCsiDaemonSet.Status.NumberAvailable)
	return int(mayastorCsiDaemonSet.Status.NumberAvailable)
}

func moacReady() bool {
	var moacDeployment appsV1.Deployment
	if gTestEnv.K8sClient.Get(context.TODO(), types.NamespacedName{Name: "moac", Namespace: common.NSMayastor()}, &moacDeployment) != nil {
		logf.Log.Info("Failed to get MOAC deployment")
		return false
	}

	logf.Log.Info("moacDeployment.Status",
		"ObservedGeneration", moacDeployment.Status.ObservedGeneration,
		"Replicas", moacDeployment.Status.Replicas,
		"UpdatedReplicas", moacDeployment.Status.UpdatedReplicas,
		"ReadyReplicas", moacDeployment.Status.ReadyReplicas,
		"AvailableReplicas", moacDeployment.Status.AvailableReplicas,
		"UnavailableReplicas", moacDeployment.Status.UnavailableReplicas,
		"CollisionCount", moacDeployment.Status.CollisionCount)
	for ix, condition := range moacDeployment.Status.Conditions {
		logf.Log.Info("Condition", "ix", ix,
			"Status", condition.Status,
			"Type", condition.Type,
			"Message", condition.Message,
			"Reason", condition.Reason)
	}

	for _, condition := range moacDeployment.Status.Conditions {
		if condition.Type == appsV1.DeploymentAvailable {
			if condition.Status == coreV1.ConditionTrue {
				logf.Log.Info("MOAC is Available")
				return true
			}
		}
	}
	logf.Log.Info("MOAC is Not Available")
	return false
}

// Checks if MOAC is available and if the requisite number of mayastor instances are
// up and running.
func MayastorReady(sleepTime int, duration int) (bool, error) {
	nodes, err := GetNodeLocs()
	if err != nil {
		return false, err
	}

	numMayastorInstances := 0
	for _, node := range nodes {
		if node.MayastorNode && !node.MasterNode {
			numMayastorInstances += 1
		}
	}

	count := (duration + sleepTime - 1) / sleepTime
	ready := false
	for ix := 0; ix < count && !ready; ix++ {
		time.Sleep(time.Duration(sleepTime) * time.Second)
		ready = mayastorReadyPodCount() == numMayastorInstances && moacReady() && mayastorCSIReadyPodCount() == numMayastorInstances
	}

	return ready, nil
}

func getClusterMayastorNodeIPAddrs() ([]string, error) {
	var nodeAddrs []string
	nodes, err := GetNodeLocs()
	if err != nil {
		return nodeAddrs, err
	}

	for _, node := range nodes {
		if node.MayastorNode {
			nodeAddrs = append(nodeAddrs, node.IPAddress)
		}
	}
	return nodeAddrs, err
}

// ListPoolsInCluster use mayastorclient to enumerate the set of mayastor pools present in the cluster
func ListPoolsInCluster() ([]mayastorclient.MayastorPool, error) {
	nodeAddrs, err := getClusterMayastorNodeIPAddrs()
	if err == nil {
		return mayastorclient.ListPools(nodeAddrs)
	}
	return []mayastorclient.MayastorPool{}, err
}

// ListNexusesInCluster use mayastorclient to enumerate the set of mayastor nexuses present in the cluster
func ListNexusesInCluster() ([]mayastorclient.MayastorNexus, error) {
	nodeAddrs, err := getClusterMayastorNodeIPAddrs()
	if err == nil {
		return mayastorclient.ListNexuses(nodeAddrs)
	}
	return []mayastorclient.MayastorNexus{}, err
}

// ListReplicasInCluster use mayastorclient to enumerate the set of mayastor replicas present in the cluster
func ListReplicasInCluster() ([]mayastorclient.MayastorReplica, error) {
	nodeAddrs, err := getClusterMayastorNodeIPAddrs()
	if err == nil {
		return mayastorclient.ListReplicas(nodeAddrs)
	}
	return []mayastorclient.MayastorReplica{}, err
}

// RmReplicasInCluster use mayastorclient to remove mayastor replicas present in the cluster
func RmReplicasInCluster() error {
	nodeAddrs, err := getClusterMayastorNodeIPAddrs()
	if err == nil {
		return mayastorclient.RmReplicas(nodeAddrs)
	}
	return err
}

// GetPoolUsageInCluster use mayastorclient to enumerate the set of pools and sum up the pool usage in the cluster
func GetPoolUsageInCluster() (uint64, error) {
	var poolUsage uint64
	pools, err := ListPoolsInCluster()
	if err == nil {
		for _, pool := range pools {
			poolUsage += pool.Used
		}
	}
	return poolUsage, err
}

// CreateConfiguredPools (re)create pools as defined by the configuration.
// No check is made on the status of pools
func CreateConfiguredPools() {
	disks := []string{e2e_config.GetConfig().PoolDevice}
	// NO check is made on the status of pools
	nodes, err := GetNodeLocs()
	Expect(err).ToNot(HaveOccurred(), "failed to get list of nodes")
	for _, node := range nodes {
		if node.MayastorNode {
			poolName := fmt.Sprintf("pool-on-%s", node.NodeName)
			pool, err := crds.CreatePool(poolName, node.NodeName, disks)
			Expect(err).ToNot(HaveOccurred(), "failed to create pool on %v, disks %v", node, disks)
			logf.Log.Info("Created", "pool", pool)
		}
	}
}

// RestoreConfiguredPools (re)create pools as defined by the configuration.
// As part of the tests we may modify the pools, in such test cases
// the test should delete all pools and recreate the configured set of pools.
func RestoreConfiguredPools() error {
	var err error
	_, err = DeleteAllPoolFinalizers()
	Expect(err).ToNot(HaveOccurred())
	deletedAllPools := DeleteAllPools()
	Expect(deletedAllPools).To(BeTrue())
	CreateConfiguredPools()
	const sleepTime = 5
	for ix := 1; ix < 120/sleepTime; ix++ {
		time.Sleep(sleepTime * time.Second)
		err := CheckAllPoolsAreOnline()
		if err == nil {
			break
		}
	}

	return CheckAllPoolsAreOnline()
}

func WaitForPoolsToBeOnline(timeoutSeconds int) error {
	const sleepTime = 5
	for ix := 1; ix < (timeoutSeconds+sleepTime)/sleepTime; ix++ {
		time.Sleep(sleepTime * time.Second)
		err := CheckAllPoolsAreOnline()
		if err == nil {
			return nil
		}
	}
	return CheckAllPoolsAreOnline()
}
