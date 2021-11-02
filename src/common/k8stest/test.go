package k8stest

import (
	"context"
	"errors"
	"fmt"
	"mayastor-e2e/common/controlplane"
	"mayastor-e2e/common/custom_resources"
	"mayastor-e2e/common/e2e_config"
	"mayastor-e2e/common/mayastorclient"
	"testing"
	"time"

	"mayastor-e2e/common"
	"mayastor-e2e/common/loki"
	"mayastor-e2e/common/reporter"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"k8s.io/client-go/deprecated/scheme"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
	"sigs.k8s.io/controller-runtime/pkg/envtest"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
)

type TestEnvironment struct {
	Cfg           *rest.Config
	K8sClient     client.Client
	KubeInt       kubernetes.Interface
	K8sManager    *ctrl.Manager
	TestEnv       *envtest.Environment
	DynamicClient dynamic.Interface
}

var gTestEnv TestEnvironment

// InitTesting initialise testing and setup class name + report filename.
func InitTesting(t *testing.T, classname string, reportname string) {
	RegisterFailHandler(Fail)
	fmt.Printf("Mayastor namespace is \"%s\"\n", common.NSMayastor())
	reporters := reporter.GetReporters(reportname)
	if len(reporters) != 0 {
		RunSpecsWithDefaultAndCustomReporters(t, classname, reporter.GetReporters(reportname))
		loki.SendLokiMarker("Start of test " + classname)
	} else {
		RunSpecs(t, reportname)
	}
}

func SetupTestEnvBasic() {
	logf.SetLogger(zap.New(zap.UseDevMode(true), zap.WriteTo(GinkgoWriter)))
	fmt.Printf("Mayastor namespace is \"%s\"\n", common.NSMayastor())

	By("bootstrapping test environment")
	var err error

	useCluster := true
	testEnv := &envtest.Environment{
		UseExistingCluster:       &useCluster,
		AttachControlPlaneOutput: true,
	}

	cfg, err := testEnv.Start()
	Expect(err).ToNot(HaveOccurred())
	Expect(cfg).ToNot(BeNil())

	k8sManager, err := ctrl.NewManager(cfg, ctrl.Options{
		Scheme: scheme.Scheme,
		// We do not consume prometheus metrics.
		MetricsBindAddress: "0",
	})
	Expect(err).ToNot(HaveOccurred())

	go func() {
		err = k8sManager.Start(ctrl.SetupSignalHandler())
		Expect(err).ToNot(HaveOccurred())
	}()

	mgrSyncCtx, mgrSyncCtxCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer mgrSyncCtxCancel()
	if synced := k8sManager.GetCache().WaitForCacheSync(mgrSyncCtx); !synced {
		fmt.Println("Failed to sync")
	}

	k8sClient := k8sManager.GetClient()
	Expect(k8sClient).ToNot(BeNil())

	restConfig := config.GetConfigOrDie()
	Expect(restConfig).ToNot(BeNil())

	kubeInt := kubernetes.NewForConfigOrDie(restConfig)
	Expect(kubeInt).ToNot(BeNil())

	dynamicClient := dynamic.NewForConfigOrDie(restConfig)
	Expect(dynamicClient).ToNot(BeNil())

	gTestEnv = TestEnvironment{
		Cfg:           cfg,
		K8sClient:     k8sClient,
		KubeInt:       kubeInt,
		K8sManager:    &k8sManager,
		TestEnv:       testEnv,
		DynamicClient: dynamicClient,
	}

	// Check if gRPC calls are possible and store the result
	// subsequent calls to mayastorClient.CanConnect retrieves
	// the result.
	mayastorclient.CheckAndSetConnect(GetMayastorNodeIPAddresses())
}

func SetupTestEnv() {
	SetupTestEnvBasic()

	err := CheckAndSetControlPlane()
	Expect(err).To(BeNil())

	// Fail the test setup if gRPC calls are mandated and
	// gRPC calls are not supported.
	if e2e_config.GetConfig().GrpcMandated {
		Expect(mayastorclient.CanConnect()).To(BeTrue(),
			"gRPC calls to mayastor are disabled, but mandated by configuration")
	}
}

func TeardownTestEnvNoCleanup() {
	err := gTestEnv.TestEnv.Stop()
	Expect(err).ToNot(HaveOccurred())
}

func TeardownTestEnv() {
	AfterSuiteCleanup()
	TeardownTestEnvNoCleanup()
}

// AfterSuiteCleanup  placeholder function for now
// To aid postmortem analysis for the most common CI use case
// namely cluster is retained aon failure, we do nothing
// For other situations behaviour should be configurable
func AfterSuiteCleanup() {
	logf.Log.Info("AfterSuiteCleanup")
}

// CheckMsPoolFinalizers check
//	1) that finalizers exist for pools with replicas (used size != 0)
//  2) that finalizers DO NOT EXIST for pools with no replicas (used size == 0)
//  with timeout to allow MOAC state sync.
func CheckMsPoolFinalizers() error {
	if controlplane.MajorVersion() != 0 {
		// Finalizers do not need to be checked with deployments of control plane versions
		// > 0 as finalizers are not added and removed when volumes/replicas are created or removed
		return nil
	}
	err := custom_resources.CheckAllMsPoolFinalizers()
	logf.Log.Info("Checking pool finalizers", "timeout seconds", e2e_config.GetConfig().MoacSyncTimeoutSeconds)
	const sleepTime = 5
	t0 := time.Now()
	for ix := 0; ix < e2e_config.GetConfig().MoacSyncTimeoutSeconds && err != nil; ix += sleepTime {
		time.Sleep(sleepTime * time.Second)
		err = custom_resources.CheckAllMsPoolFinalizers()
	}
	if err != nil {
		logf.Log.Info("Checking pool finalizers", "error", err)
	} else {
		logf.Log.Info("Checking pool finalizers, done.", "waiting time", time.Since(t0))
	}
	return err
}

func getMspUsage() (int64, error) {
	var mspUsage int64
	msPools, err := custom_resources.ListMsPools()
	if err != nil {
		logf.Log.Info("unable to list mayastor pools")
	} else {
		mspUsage = 0
		for _, pool := range msPools {
			mspUsage += pool.Status.Used
		}
	}
	return mspUsage, err
}

// ResourceCheck  Fit for purpose checks
// - No pods
// - No PVCs
// - No PVs
// - No MSVs
// - Mayastor pods are all healthy
// - All mayastor pools are online
// and if e2e-agent is available
// - mayastor pools usage is 0
// - No nexuses
// - No replicas
func ResourceCheck() error {
	var errorMsg = ""

	pods, err := CheckForTestPods()
	if err != nil {
		errorMsg += fmt.Sprintf("%s %v", errorMsg, err)
	}
	if pods {
		errorMsg += " found Pods"
	}

	pvcs, err := CheckForPVCs()
	if err != nil {
		errorMsg += fmt.Sprintf("%s %v", errorMsg, err)
	}
	if pvcs {
		errorMsg += " found PersistentVolumeClaims"
	}

	pvs, err := CheckForPVs()
	if err != nil {
		errorMsg += fmt.Sprintf("%s %v", errorMsg, err)
	}
	if pvs {
		errorMsg += " found PersistentVolumes"
	}

	//FIXME: control plane 1 temporary do not check MSVs
	if controlplane.MajorVersion() == 0 {
		// Mayastor volumes
		msvs, err := ListMsvs()
		if err != nil {
			errorMsg += fmt.Sprintf("%s %v", errorMsg, err)
		} else {
			if msvs != nil {
				if len(msvs) != 0 {
					errorMsg += " found MayastorVolumes"
				}
			} else {
				logf.Log.Info("Listing MSVs returned nil array")
			}
		}
	}

	// Check that Mayastor pods are healthy no restarts or fails.
	err = CheckTestPodsHealth(common.NSMayastor())
	if err != nil {
		if e2e_config.GetConfig().SelfTest {
			logf.Log.Info("SelfTesting, ignoring:", "", err)
		} else {
			errorMsg += fmt.Sprintf("%s %v", errorMsg, err)
		}
	}

	scs, err := CheckForStorageClasses()
	if err != nil {
		errorMsg += fmt.Sprintf("%s %v", errorMsg, err)
	}
	if scs {
		errorMsg += " found storage classes using mayastor "
	}

	err = custom_resources.CheckAllMsPoolsAreOnline()
	if err != nil {
		errorMsg += fmt.Sprintf("%s %v", errorMsg, err)
		logf.Log.Info("ResourceCheck: not all pools are online")
	}

	mspUsage, err := getMspUsage()
	if err != nil || mspUsage != 0 {
		logf.Log.Info("Waiting for pool usage to be 0")
		const sleepTime = 10
		t0 := time.Now()
		// Wait for pool usage reported by CRS to drop to 0
		for ix := 0; ix < (60*sleepTime) && mspUsage != 0; ix += sleepTime {
			time.Sleep(sleepTime * time.Second)
			mspUsage, err = getMspUsage()
			if err != nil {
				errorMsg += fmt.Sprintf("%s %v", errorMsg, err)
				logf.Log.Info("ResourceCheck: unable to list msps")
			}
		}
		logf.Log.Info("ResourceCheck:", "mspool Usage", mspUsage, "waiting time", time.Since(t0))
		Expect(mspUsage).To(BeZero(), "pool usage reported via custom resources %d", mspUsage)
	}

	// gRPC calls can only be executed successfully is the e2e-agent daemonSet has been deployed successfully.
	if mayastorclient.CanConnect() {
		// check pools
		{
			var poolUsage uint64 = 1
			const sleepTime = 2
			t0 := time.Now()
			// Wait for pool usage to drop to 0
			for ix := 0; ix < 120 && poolUsage != 0; ix += sleepTime {
				time.Sleep(sleepTime * time.Second)
				poolUsage, err = GetPoolUsageInCluster()
				if err != nil {
					errorMsg += fmt.Sprintf("%s %v", errorMsg, err)
					logf.Log.Info("ResourceEachCheck: failed to retrieve pools usage")
				}
			}
			logf.Log.Info("ResourceCheck:", "poolUsage", poolUsage, "waiting time", time.Since(t0))
			Expect(poolUsage).To(BeZero(), "pool usage reported via mayastor client is %d", poolUsage)
		}
		// check nexuses
		{
			nexuses, err := ListNexusesInCluster()
			if err != nil {
				errorMsg += fmt.Sprintf("%s %v", errorMsg, err)
				logf.Log.Info("ResourceEachCheck: failed to retrieve list of nexuses")
			}
			logf.Log.Info("ResourceCheck:", "num nexuses", len(nexuses))
			Expect(len(nexuses)).To(BeZero(), "count of nexuses reported via mayastor client is %d", len(nexuses))
		}
		// check replicas
		{
			replicas, err := ListReplicasInCluster()
			if err != nil {
				errorMsg += fmt.Sprintf("%s %v", errorMsg, err)
				logf.Log.Info("ResourceEachCheck: failed to retrieve list of replicas")
			}
			logf.Log.Info("ResourceCheck:", "num replicas", len(replicas))
			Expect(len(replicas)).To(BeZero(), "count of replicas reported via mayastor client is %d", len(replicas))
		}
		// check nvmeControllers
		{
			nvmeControllers, err := ListNvmeControllersInCluster()
			if err != nil {
				errorMsg += fmt.Sprintf("%s %v", errorMsg, err)
				logf.Log.Info("ResourceEachCheck: failed to retrieve list of nvme controllers")
			}
			logf.Log.Info("ResourceCheck:", "num nvme controllers", len(nvmeControllers))
			Expect(len(nvmeControllers)).To(BeZero(), "count of replicas reported via mayastor client is %d", len(nvmeControllers))
		}
	} else {
		logf.Log.Info("WARNING: gRPC calls to mayastor are not enabled, all checks cannot be run")
	}

	if len(errorMsg) != 0 {
		return errors.New(errorMsg)
	}
	return nil
}

//BeforeEachCheck asserts that the state of mayastor resources is fit for the test to run
func BeforeEachCheck() error {
	logf.Log.Info("BeforeEachCheck")

	err := ResourceCheck()
	if err != nil {
		logf.Log.Info("ResourceCheck failed", "CleanupOnBeforeEach", e2e_config.GetConfig().CleanupOnBeforeEach)
		if e2e_config.GetConfig().CleanupOnBeforeEach {
			_ = CleanUp()
			_ = RestoreConfiguredPools()
			err = ResourceCheck()
		}
		if err != nil {
			logf.Log.Info("BeforeEachCheck failed", "error", err)
		}
	}
	if err != nil {
		err = fmt.Errorf("not running test case, k8s cluster is not \"clean\"!!!\n%v", err)
	}

	if err != nil {
		logf.Log.Info("AfterEachCheck failed", "error", err)
	}
	return err
}

// AfterEachCheck asserts that the state of mayastor resources has been restored.
func AfterEachCheck() error {
	logf.Log.Info("AfterEachCheck")
	err := ResourceCheck()
	if err == nil {
		err = CheckMsPoolFinalizers()
		if e2e_config.GetConfig().SelfTest {
			logf.Log.Info("SelfTesting, ignoring:", "", err)
			err = nil
		}
	}
	return err
}
