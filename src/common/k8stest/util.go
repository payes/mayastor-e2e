package k8stest

import (
	"context"
	"fmt"
	"mayastor-e2e/common/controlplane"
	"mayastor-e2e/common/custom_resources"
	"mayastor-e2e/common/e2e_config"
	"os/exec"
	"regexp"
	"time"

	"mayastor-e2e/common"
	"mayastor-e2e/common/mayastorclient"

	. "github.com/onsi/gomega"
	errors "github.com/pkg/errors"

	appsV1 "k8s.io/api/apps/v1"
	coreV1 "k8s.io/api/core/v1"
	v1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
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

// create a storage class with default volume binding mode i.e. not specified
func MkStorageClass(scName string, scReplicas int, protocol common.ShareProto, nameSpace string) error {
	return NewScBuilder().
		WithName(scName).
		WithReplicas(scReplicas).
		WithProtocol(protocol).
		WithNamespace(nameSpace).
		BuildAndCreate()
}

// remove a storage class
func RmStorageClass(scName string) error {
	logf.Log.Info("Deleting storage class", "name", scName)
	ScApi := gTestEnv.KubeInt.StorageV1().StorageClasses
	deleteErr := ScApi().Delete(context.TODO(), scName, metaV1.DeleteOptions{})
	if k8serrors.IsNotFound(deleteErr) {
		return nil
	}
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

func SetReplication(appName string, namespace string, replicas *int32) {
	depAPI := gTestEnv.KubeInt.AppsV1().Deployments
	stsAPI := gTestEnv.KubeInt.AppsV1().StatefulSets
	labels := "app=" + appName
	deployments, err := depAPI(namespace).List(context.TODO(), metaV1.ListOptions{LabelSelector: labels})
	Expect(err).ToNot(HaveOccurred())
	sts, err := stsAPI(namespace).List(context.TODO(), metaV1.ListOptions{LabelSelector: labels})
	Expect(err).ToNot(HaveOccurred())
	if len(deployments.Items) == 1 {
		SetDeploymentReplication(deployments.Items[0].Name, namespace, replicas)
	} else if len(sts.Items) == 1 {
		SetStatefulsetReplication(sts.Items[0].Name, namespace, replicas)
	} else {
		Expect(false).To(BeTrue(), "app %s is not deployed as a deployment or sts", appName)
	}
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

// Adjust the number of replicas in the statefulset
func SetStatefulsetReplication(statefulsetName string, namespace string, replicas *int32) {
	stsAPI := gTestEnv.KubeInt.AppsV1().StatefulSets
	var err error

	// this is to cater for a race condition, occasionally seen,
	// when the deployment is changed between Get and Update
	for attempts := 0; attempts < 10; attempts++ {
		sts, err := stsAPI(namespace).Get(context.TODO(), statefulsetName, metaV1.GetOptions{})
		Expect(err).ToNot(HaveOccurred())
		sts.Spec.Replicas = replicas
		_, err = stsAPI(namespace).Update(context.TODO(), sts, metaV1.UpdateOptions{})
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

// FIXME check callers in tests - what should be the state with control plane versions > 0
// FIXME Doh! overloaded semantics :-(
// Checks if MayastorVersion is available and if the requisite number of mayastor instances are
// up and running.
func MayastorInstancesReady(numMayastorInstances int, sleepTime int, duration int) (bool, error) {

	count := (duration + sleepTime - 1) / sleepTime
	ready := false
	for ix := 0; ix < count && !ready; ix++ {
		time.Sleep(time.Duration(sleepTime) * time.Second)
		switch controlplane.MajorVersion() {
		case 0:
			ready = mayastorReadyPodCount() == numMayastorInstances && mayastorCSIReadyPodCount() >= numMayastorInstances && msnOnlineCount() == numMayastorInstances
		case 1:
			ready = mayastorReadyPodCount() == numMayastorInstances && mayastorCSIReadyPodCount() >= numMayastorInstances
		default:
			panic("unexpected control plane version")
		}
	}

	return ready, nil
}

func msnOnlineCount() int {
	msns, err := custom_resources.ListMsNodes()
	if err != nil {
		logf.Log.Info("Failed to List nodes", "error", err)
		return -1
	}
	count := 0
	if len(msns) != 0 {
		for _, msn := range msns {
			if msn.Status == "online" {
				count++
			}
		}
	}
	logf.Log.Info("msn online", "count", count)
	return count
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

func DeploymentReady(deploymentName, namespace string) bool {
	var deployment appsV1.Deployment
	if err := gTestEnv.K8sClient.Get(context.TODO(), types.NamespacedName{Name: deploymentName, Namespace: namespace}, &deployment); err != nil {
		logf.Log.Info("Failed to get deployment", "error", err)
		return false
	}

	for _, condition := range deployment.Status.Conditions {
		if condition.Type == appsV1.DeploymentAvailable {
			if condition.Status == coreV1.ConditionTrue {
				return true
			}
		}
	}
	return false
}

func DaemonSetReady(daemonName string, namespace string) bool {
	var daemon appsV1.DaemonSet
	if err := gTestEnv.K8sClient.Get(context.TODO(), types.NamespacedName{Name: daemonName, Namespace: namespace}, &daemon); err != nil {
		logf.Log.Info("Failed to get daemonset", "error", err)
		return false
	}

	status := daemon.Status
	logf.Log.Info("DaemonSet "+daemonName, "status", status)
	return status.DesiredNumberScheduled == status.CurrentNumberScheduled &&
		status.DesiredNumberScheduled == status.NumberReady &&
		status.DesiredNumberScheduled == status.NumberAvailable
}

func StatefulSetReady(statefulSetName string, namespace string) bool {
	var statefulSet appsV1.StatefulSet
	if err := gTestEnv.K8sClient.Get(context.TODO(), types.NamespacedName{Name: statefulSetName, Namespace: namespace}, &statefulSet); err != nil {
		logf.Log.Info("Failed to get daemonset", "error", err)
		return false
	}
	status := statefulSet.Status
	logf.Log.Info("StatefulSet "+statefulSetName, "status", status)
	return status.Replicas == status.ReadyReplicas &&
		status.ReadyReplicas == status.CurrentReplicas
}

func ControlPlaneReady(sleepTime int, duration int) bool {
	ready := false
	count := (duration + sleepTime - 1) / sleepTime

	if controlplane.MajorVersion() == 1 {
		nonControlPlaneComponents := []string{"mayastor", "mayastor-csi", "nats"}

		for ix := 0; ix < count && !ready; ix++ {
			time.Sleep(time.Duration(sleepTime) * time.Second)
			deployments, err := gTestEnv.KubeInt.AppsV1().Deployments(common.NSMayastor()).List(context.TODO(), metaV1.ListOptions{})
			if err != nil {
				time.Sleep(time.Duration(sleepTime) * time.Second)
				continue
			}
			daemonsets, err := gTestEnv.KubeInt.AppsV1().DaemonSets(common.NSMayastor()).List(context.TODO(), metaV1.ListOptions{})
			if err != nil {
				time.Sleep(time.Duration(sleepTime) * time.Second)
				continue
			}
			statefulsets, err := gTestEnv.KubeInt.AppsV1().StatefulSets(common.NSMayastor()).List(context.TODO(), metaV1.ListOptions{})
			if err != nil {
				time.Sleep(time.Duration(sleepTime) * time.Second)
				continue
			}
			ready = true
			for _, deployment := range deployments.Items {
				if contains(nonControlPlaneComponents, deployment.Name) {
					continue
				}
				tmp := DeploymentReady(deployment.Name, common.NSMayastor())
				logf.Log.Info("mayastor control plane", "deployment", deployment.Name, "ready", tmp)
				ready = ready && tmp
			}
			for _, daemon := range daemonsets.Items {
				if contains(nonControlPlaneComponents, daemon.Name) {
					continue
				}
				tmp := DaemonSetReady(daemon.Name, common.NSMayastor())
				logf.Log.Info("mayastor control plane", "daemonset", daemon.Name, "ready", tmp)
				ready = ready && tmp
			}
			for _, statefulSet := range statefulsets.Items {
				if contains(nonControlPlaneComponents, statefulSet.Name) {
					continue
				}
				tmp := StatefulSetReady(statefulSet.Name, common.NSMayastor())
				logf.Log.Info("mayastor control plane", "statefulset", statefulSet.Name, "ready", tmp)
				ready = ready && tmp
			}
		}
	} else {
		for ix := 0; ix < count && !ready; ix++ {
			time.Sleep(time.Duration(sleepTime) * time.Second)
			ready = moacReady()
		}
	}
	return ready
}

func contains(list []string, str string) bool {
	for _, elem := range list {
		if elem == str {
			return true
		}
	}
	return false
}

// Checks if the requisite number of mayastor instances are up and running.
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
	return MayastorInstancesReady(numMayastorInstances, sleepTime, duration)
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
		return mayastorclient.RmNodeReplicas(nodeAddrs)
	}
	return err
}

// ListNvmeControllersInCluster use mayastorclient to enumerate the set of mayastor nvme controllers present in the cluster
func ListNvmeControllersInCluster() ([]mayastorclient.NvmeController, error) {
	nodeAddrs, err := getClusterMayastorNodeIPAddrs()
	if err == nil {
		return mayastorclient.ListNvmeControllers(nodeAddrs)
	}
	return []mayastorclient.NvmeController{}, err
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
	Expect(len(e2e_config.GetConfig().PoolDevice)).ToNot(BeZero(), "pool device not configured")
	disks := []string{e2e_config.GetConfig().PoolDevice}
	// NO check is made on the status of pools
	nodes, err := GetNodeLocs()
	Expect(err).ToNot(HaveOccurred(), "failed to get list of nodes")
	for _, node := range nodes {
		if node.MayastorNode {
			poolName := fmt.Sprintf("pool-on-%s", node.NodeName)
			pool, err := custom_resources.CreateMsPool(poolName, node.NodeName, disks)
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
	const sleepTime = 5
	pools := []mayastorclient.MayastorPool{}
	for ix := 1; ix < 120/sleepTime; ix++ {
		pools, err = mayastorclient.ListPools(GetMayastorNodeIPAddresses())
		if err != nil {
			logf.Log.Info("ListPools", "error", err)
		}
		if len(pools) == 0 && err == nil {
			break
		}
		time.Sleep(sleepTime * time.Second)
	}
	Expect(err).ToNot(HaveOccurred())

	for ix := 1; ix < 120/sleepTime && len(pools) != 0; ix++ {
		err = mayastorclient.DestroyAllPools(GetMayastorNodeIPAddresses())
		if err != nil {
			logf.Log.Info("DestroyAllPools", "error", err)
		}
		pools, err = mayastorclient.ListPools(GetMayastorNodeIPAddresses())
		if err != nil {
			logf.Log.Info("ListPools", "error", err)
		}
		time.Sleep(sleepTime * time.Second)
	}
	Expect(err).ToNot(HaveOccurred())
	Expect(len(pools)).To(Equal(0))

	CreateConfiguredPools()
	for ix := 1; ix < 120/sleepTime; ix++ {
		time.Sleep(sleepTime * time.Second)
		err := custom_resources.CheckAllMsPoolsAreOnline()
		if err == nil {
			break
		}
	}

	return custom_resources.CheckAllMsPoolsAreOnline()
}

func WaitForPoolsToBeOnline(timeoutSeconds int) error {
	const sleepTime = 5
	for ix := 1; ix < (timeoutSeconds+sleepTime)/sleepTime; ix++ {
		time.Sleep(sleepTime * time.Second)
		err := custom_resources.CheckAllMsPoolsAreOnline()
		if err == nil {
			return nil
		}
	}
	return custom_resources.CheckAllMsPoolsAreOnline()
}

// WaitPodComplete waits until pod is in completed state
func WaitPodComplete(podName string, sleepTimeSecs, timeoutSecs int) error {
	var podPhase coreV1.PodPhase
	var err error

	logf.Log.Info("Waiting for pod to complete", "name", podName)
	for ix := 0; ix < (timeoutSecs+sleepTimeSecs-1)/sleepTimeSecs; ix++ {
		time.Sleep(time.Duration(sleepTimeSecs) * time.Second)
		podPhase, err = CheckPodCompleted(podName, common.NSDefault)
		Expect(err).ToNot(HaveOccurred(), fmt.Sprintf("failed to access pod status %s %v", podName, err))
		if podPhase == coreV1.PodSucceeded {
			return nil
		} else if podPhase == coreV1.PodFailed {
			break
		}
	}
	return errors.Errorf("pod did not complete, phase %v", podPhase)
}

// WorkaroundForMQ1536 work around required for MQ-1536
// TODO revisit.
func WorkaroundForMQ1536() {
	_, err := DeleteAllPoolFinalizers()
	Expect(err).ToNot(HaveOccurred(), "failed to delete all pool finalizers (WorkaroundForMQ1536)")
}

func MakeAccumulatedError(accErr error, err error) error {
	if accErr == nil {
		return err
	}
	return fmt.Errorf("%v; %v", accErr, err)
}

// DeleteVolumeAttachmets deletes volume attachments for a node
func DeleteVolumeAttachments(nodeName string) error {
	volumeAttachments, err := gTestEnv.KubeInt.StorageV1().VolumeAttachments().List(context.TODO(), metaV1.ListOptions{})
	Expect(err).ToNot(HaveOccurred(), "failed to list volume attachments")
	if len(volumeAttachments.Items) == 0 {
		return nil
	}
	for _, volumeAttachment := range volumeAttachments.Items {
		if volumeAttachment.Spec.NodeName != nodeName {
			continue
		}
		logf.Log.Info("DeleteVolumeAttachments: Deleting", "volumeAttachment", volumeAttachment.Name)
		delErr := gTestEnv.KubeInt.StorageV1().VolumeAttachments().Delete(context.TODO(), volumeAttachment.Name, metaV1.DeleteOptions{})
		if delErr != nil {
			logf.Log.Info("DeleteVolumeAttachments: failed to delete the volumeAttachment", "volumeAttachment", volumeAttachment.Name, "error", delErr)
			return err
		}
	}
	return nil
}

// CheckAndSetControlPlane checks which deployments exists and sets config control plane setting
func CheckAndSetControlPlane() error {
	var deployment appsV1.Deployment
	var statefulSet appsV1.StatefulSet
	var err error
	var foundMoac = false
	var foundCoreAgents = false
	var version string

	if err = gTestEnv.K8sClient.Get(context.TODO(), types.NamespacedName{Name: "moac", Namespace: common.NSMayastor()}, &deployment); err == nil {
		foundMoac = true
	}
	// Check for core-agents either as deployment or statefulset to correctly handle older builds of control plane
	// which use core-agents deployment and newer builds which use core-agents statefulset
	if err = gTestEnv.K8sClient.Get(context.TODO(), types.NamespacedName{Name: "core-agents", Namespace: common.NSMayastor()}, &deployment); err == nil {
		foundCoreAgents = true
	}
	if err = gTestEnv.K8sClient.Get(context.TODO(), types.NamespacedName{Name: "core-agents", Namespace: common.NSMayastor()}, &statefulSet); err == nil {
		foundCoreAgents = true
	}

	if foundMoac && foundCoreAgents {
		return fmt.Errorf("MOAC and Restful Control plane components are present")
	}
	if !foundMoac && !foundCoreAgents {
		return fmt.Errorf("MOAC and Restful Control plane components are absent")
	}
	if foundMoac {
		version = "0.8.2"
	}
	if foundCoreAgents {
		version = "1.0.0"
	}

	logf.Log.Info("CheckAndSetControlPlane", "version", version)
	if !e2e_config.SetControlPlane(version) {
		return fmt.Errorf("failed to setup config control plane to %s", version)
	}
	return nil
}

// Checks if the requisite number of mayastor node are online.
func MayastorNodeReady(sleepTime int, duration int) (bool, error) {
	ready := false
	readyCount := 0
	msReadyPodCount := mayastorReadyPodCount()
	count := (duration + sleepTime - 1) / sleepTime
	for ix := 0; ix < count && !ready; ix++ {
		time.Sleep(time.Duration(sleepTime) * time.Second)
		// list mayastor node
		nodeList, err := ListMsns()
		if err != nil {
			logf.Log.Info("MayastorNodeReady: failed to list mayastor node", "error", err)
			return ready, err
		}
		for _, node := range nodeList {
			if node.State.Status == controlplane.NodeStateOnline() {
				readyCount++
			} else {
				logf.Log.Info("Not ready node", "node", node.Name, "status", node.State.Status)
			}
		}
		ready = msReadyPodCount == len(nodeList) && readyCount == msReadyPodCount
		logf.Log.Info("mayastor node status",
			"MayastorReadyPodCount", msReadyPodCount,
			"MayastorNodes", len(nodeList),
			"MaystorNodeReadyCount", readyCount,
		)
	}
	return ready, nil
}
