package e2e_config

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"strings"
	"sync"

	"gopkg.in/yaml.v2"

	"github.com/ilyakaznacheev/cleanenv"
)

const ConfigDir = "/configurations"
const PlatformConfigDir = "/configurations/platforms/"

// E2EConfig is a application configuration structure
type E2EConfig struct {
	ConfigName  string `yaml:"configName" env-default:"default"`
	ConfigPaths struct {
		ConfigFile         string `yaml:"configFile" env:"e2e_config_file" env-default:""`
		PlatformConfigFile string `yaml:"platformConfigFile" env:"e2e_platform_config_file" env-default:""`
	} `yaml:"configPaths"`
	Platform struct {
		// E2ePlatform indicates where the e2e is currently being run from
		Name string `yaml:"name" env-default:"default"`
		// Add HostNetwork: true to the spec of test pods.
		HostNetworkingRequired bool `yaml:"hostNetworkingRequired" env-default:"false"`
		// Some deployments use a different namespace
		MayastorNamespace string `yaml:"mayastorNamespace" env-default:"mayastor"`
		// Some deployments use a different namespace
		FilteredMayastorPodCheck int `yaml:"filteredMayastorPodCheck" env-default:"0"`
	} `yaml:"platform"`

	// gRPC connection to the mayastor is mandatory for the test run
	// With few exceptions, all CI configurations MUST set this to true
	GrpcMandated bool `yaml:"grpcMandated" env-default:"false"`
	// Generic configuration files used for CI and automation should not define MayastorRootDir and E2eRootDir
	MayastorRootDir  string `yaml:"mayastorRootDir" env:"e2e_mayastor_root_dir"`
	E2eRootDir       string `yaml:"e2eRootDir"`
	SessionDir       string `yaml:"sessionDir" env:"e2e_session_dir"`
	MayastorVersion  string `yaml:"mayastorVersion" env:"e2e_mayastor_version"`
	KubectlPluginDir string `yaml:"kubectlPluginDir" env:"e2e_kubectl_plugin_dir"`

	// Operational parameters
	Cores int `yaml:"cores,omitempty"`
	// Registry from where mayastor images are retrieved
	Registry string `yaml:"registry" env:"e2e_docker_registry" env-default:"ci-registry.mayastor-ci.mayadata.io"`
	//	// Registry from where CI testing images are retrieved
	//	CIRegistry string `yaml:"ciRegistry" env:"e2e_ci_docker_registry" env-default:"ci-registry.mayastor-ci.mayadata.io"`
	ImageTag string `yaml:"imageTag" env:"e2e_image_tag"`
	// FIXME: handle empty poolDevice
	PoolDevice  string `yaml:"poolDevice" env:"e2e_pool_device"`
	E2eFioImage string `yaml:"e2eFioImage" env-default:"mayadata/e2e-fio" env:"e2e_fio_image"`
	E2eFsxImage string `yaml:"e2eFsxImage" env-default:"mayadata/e2e-fsx" env:"e2e_fsx_image"`
	// This is an advisory setting for individual tests
	// If set to true - typically during test development - tests with multiple 'It' clauses should defer asserts till after
	// resources have been cleaned up . This behaviour makes it possible to have useful runs for all 'It' clauses.
	// Typically, set to false for CI test execution - no cleanup after first failure, as a result subsequent 'It' clauses
	// in the test will fail the BeforeEach check, rendering post-mortem checks on the cluster more useful.
	// It may be set to true for when we want maximum test coverage, and post-mortem analysis is a secondary requirement.
	// NOTE: Only some tests support this feature.
	DeferredAssert bool `yaml:"deferredAssert" env-default:"false" env:"e2e_defer_asserts"`
	// TODO: for now using a simple boolean for a specific behaviour suffices, a more sophisticated approach using a policy for test runs may be required.

	// Default replica count, used by tests which do not have a config section.
	DefaultReplicaCount int `yaml:"defaultReplicaCount" env-default:"2" env:"e2e_default_replica_count"`
	// Restart Mayastor on failure in a prior AfterEach or ResourceCheck
	BeforeEachCheckAndRestart bool `yaml:"beforeEachCheckAndRestart" env-default:"false"`
	// Fail  quickly after failure of a prior AfterEach, overrides BeforeEachCheckAndRestart
	FailQuick bool `yaml:"failQuick" env-default:"false" env:"e2e_fail_quick"`

	// Run configuration
	ReportsDir string `yaml:"reportsDir" env:"e2e_reports_dir"`
	SelfTest   bool   `yaml:"selfTest" env:"e2e_self_test" env-default:"false"`

	// Individual Test parameters
	PVCStress struct {
		Replicas   int `yaml:"replicas" env-default:"2"`
		CdCycles   int `yaml:"cdCycles" env-default:"100"`
		CrudCycles int `yaml:"crudCycles" env-default:"10"`
	} `yaml:"pvcStress"`
	IOSoakTest struct {
		Replicas int    `yaml:"replicas" env-default:"2"`
		Duration string `yaml:"duration" env-default:"60m"`
		// Number of volumes for each mayastor instance
		// volumes for disruptor pods are allocated from within this "pool"
		LoadFactor int      `yaml:"loadFactor" env-default:"10"`
		Protocols  []string `yaml:"protocols" env-default:"nvmf"`
		// FioStartDelay units are seconds
		FioStartDelay int    `yaml:"fioStartDelay" env-default:"90"`
		ReadyTimeout  string `yaml:"readyTimeout" env-default:"600s"`
		Disrupt       struct {
			// Number of disruptor pods.
			PodCount int `yaml:"podCount" env-default:"3"`
			// FaultAfter units are seconds
			FaultAfter   int    `yaml:"faultAfter" env-default:"51"`
			ReadyTimeout string `yaml:"readyTimeout" env-default:"180s"`
		} `yaml:"disrupt"`
		FioDutyCycles []struct {
			// ThinkTime units are microseconds
			ThinkTime       int `yaml:"thinkTime"`
			ThinkTimeBlocks int `yaml:"thinkTimeBlocks"`
		} `yaml:"fioDutyCycles"`
	} `yaml:"ioSoakTest"`
	CSI struct {
		Replicas       int    `yaml:"replicas" env-default:"2"`
		SmallClaimSize string `yaml:"smallClaimSize" env-default:"50Mi"`
		LargeClaimSize string `yaml:"largeClaimSize" env-default:"500Mi"`
	} `yaml:"csi"`
	Uninstall struct {
		Cleanup int `yaml:"cleanup" env:"e2e_uninstall_cleanup" env-default:"1"`
	} `yaml:"uninstall"`
	BasicVolumeIO struct {
		// FioTimeout is in seconds
		FioTimeout int `yaml:"fioTimeout" env-default:"90"`
		// VolSizeMb Units are MiB
		VolSizeMb int `yaml:"volSizeMb" env-default:"500"`
		// FsVolSizeMb Units are MiB
		FsVolSizeMb int `yaml:"fsVolSizeMb" env-default:"450"`
	} `yaml:"basicVolumeIO"`
	MultipleVolumesPodIO struct {
		VolumeSizeMb         int    `yaml:"volumeSizeMb" env-default:"500"`
		VolumeCount          int    `yaml:"volumeCount" env-default:"6"`
		MultipleReplicaCount int    `yaml:"replicas" env-default:"2"`
		FioLoops             int    `yaml:"fioLoops" env-default:"0"`
		Timeout              string `yaml:"timeout" env-default:"1800s"`
	} `yaml:"multiVolumesPodIO"`
	MsPodDisruption struct {
		VolMb                    int `yaml:"volMb" env-default:"4096"`
		RemoveThinkTime          int `yaml:"removeThinkTime" env-default:"10"`
		RepairThinkTime          int `yaml:"repairThinkTime" env-default:"30"`
		ThinkTimeBlocks          int `yaml:"thinkTimeBlocks" env-default:"10"`
		UnscheduleDelay          int `yaml:"unscheduleDelay" env-default:"10"`
		RescheduleDelay          int `yaml:"rescheduleDelay" env-default:"10"`
		PodUnscheduleTimeoutSecs int `yaml:"podUnscheduleTimeoutSecs" env-default:"100"`
		PodRescheduleTimeoutSecs int `yaml:"podRnscheduleTimeoutSecs" env-default:"180"`
		PodRemovalTest           int `yaml:"podRemovalTest" env-default:"0"`
		DeviceRemovalTest        int `yaml:"deviceRemovalTest" env-default:"1"`
	} `yaml:"msPodDisruption"`
	MaximumVolsIO struct {
		VolMb             int    `yaml:"volMb" env-default:"64"`
		VolumeCountPerPod int    `yaml:"volumeCountPerPod" env-default:"10"`
		PodCount          int    `yaml:"podCount" env-default:"11"`
		Duration          string `yaml:"duration" env-default:"240s"`
		Timeout           string `yaml:"timeout" env-default:"600s"`
		ThinkTime         string `yaml:"thinkTime" env-default:"10ms"`
	} `yaml:"maximumVolsIO"`
	ControlPlaneRescheduling struct {
		// Count of mayastor volume
		MayastorVolumeCount int `yaml:"mayastorVolumeCount" env-default:"3"`
	} `yaml:"controlPlaneRescheduling"`
	ExpandMspDisk struct {
		// DiskPath is the path of the disk
		DiskPath string `yaml:"diskPath" env-default:"/dev/sdb"`
		// PartitionStartSize is the start size of partitioned disk
		PartitionStartSize string `yaml:"partitionStartSize" env-default:"1GiB"`
		// PartitionEndSize is the end size of partitioned disk
		PartitionEndSize string `yaml:"partitionEndSize" env-default:"3GiB"`
		// ResizePartitionDisk is the end size of partiioned disk to resize the disk
		ResizePartitionDisk string `yaml:"resizePartitionDisk" env-default:"5GiB"`
		// Duration is in seconds
		Duration string `yaml:"duration" env-default:"60"`
		// VolSizeMb Units are MiB
		VolSizeMb string `yaml:"volSizeMb" env-default:"50"`
	}
	ValidateIntegrityTest struct {
		Replicas   int    `yaml:"replicas" env-default:"3"`
		FioTimeout int    `yaml:"fioTimeout" env-default:"2000"`
		VolMb      int    `yaml:"volMb" env-default:"9900"`
		Device     string `yaml:"device" env-default:"/dev/sdb"`
	} `yaml:"validateIntegrityTest"`
	PvcReadWriteOnce struct {
		// FioTimeout is in seconds
		FioTimeout int `yaml:"fioTimeout" env-default:"120"`
	} `yaml:"pvcReadWriteOnce"`
	PvcDelete struct {
		// VolSizeMb Units are MiB
		VolSizeMb int `yaml:"volSizeMb" env-default:"1024"`
		// FsVolSizeMb Units are MiB
		FsVolSizeMb              int `yaml:"fsVolSizeMb" env-default:"900"`
		PodUnscheduleTimeoutSecs int `yaml:"podUnscheduleTimeoutSecs" env-default:"100"`
		PodRescheduleTimeoutSecs int `yaml:"podRnscheduleTimeoutSecs" env-default:"180"`
	} `yaml:"pvcDelete"`
	PrimitiveMaxVolsInPool struct {
		VolMb              int `yaml:"volMb" env-default:"64"`
		VolumeCountPerPool int `yaml:"volumeCountPerPool" env-default:"110"`
		Replicas           int `yaml:"replicas" env-default:"2"`
	} `yaml:"primitiveMaxVolsInPool"`
	PrimitiveMspState struct {
		ReplicaSize            int    `yaml:"replicaSize" env-default:"1073741824"`
		PoolDeleteTimeoutSecs  string `yaml:"poolDeleteTimeoutSecs" env-default:"30s"`
		PoolCreateTimeoutSecs  string `yaml:"poolCreateTimeoutSecs" env-default:"20s"`
		PoolUsageTimeoutSecs   string `yaml:"poolUsageTimeoutSecs" env-default:"90s"`
		PoolUsageSleepTimeSecs string `yaml:"poolUsageSleepTimeSecs" env-default:"2s"`
		IterationCount         int    `yaml:"iterationCount" env-default:"100"`
	} `yaml:"primitiveMspState"`
	PrimitiveReplicas struct {
		Iterations  int `yaml:"iterations" env-default:"100"`
		StartSizeMb int `yaml:"startSizeMb" env-default:"128"`
		EndSizeMb   int `yaml:"endSizeMb" env-default:"4096"`
		SizeStepMb  int `yaml:"sizeStepMb" env-default:"64"`
	} `yaml:"primitiveReplicas"`
	PrimitiveMspDelete struct {
		ReplicaSize            int    `yaml:"replicaSize" env-default:"10000000"`
		ReplicasTimeoutSecs    string `yaml:"replicasTimeoutSecs" env-default:"30s"`
		PoolUsageTimeoutSecs   string `yaml:"poolUsageTimeoutSecs" env-default:"30s"`
		PoolDeleteTimeoutSecs  string `yaml:"poolDeleteTimeoutSecs" env-default:"40s"`
		PoolCreateTimeoutSecs  string `yaml:"poolCreateTimeoutSecs" env-default:"20s"`
		MayastorRestartTimeout int    `yaml:"mayastorRestartTimeout" env-default:"240"`
		Iterations             int    `yaml:"iterations" env-default:"30"`
	} `yaml:"primitiveMspDelete"`

	PrimitiveMspStressTest struct {
		PartitionSizeInGiB int `yaml:"partitionSizeInGiB" env-default:"1"`
		PartitionCount     int `yaml:"partitionCount" env-default:"5"`
		Iterations         int `yaml:"iterations" env-default:"10"`
	} `yaml:"PrimitiveMspStressTest"`
	ConcurrentPvcCreate struct {
		Replicas        int `yaml:"replicas" env-default:"1"`
		VolSize         int `yaml:"volMb" env-default:"64"`
		Iterations      int `yaml:"iterations" env-default:"10"`
		VolumeMultipler int `yaml:"volumeMultipler" env-default:"10"`
	} `yaml:"concurrentPvcCreate"`
	PrimitiveFaultInjection struct {
		VolMb     int    `yaml:"volMb" env-default:"512"`
		Replicas  int    `yaml:"replicas" env-default:"3"`
		Duration  string `yaml:"duration" env-default:"240s"`
		Timeout   string `yaml:"timeout" env-default:"420s"`
		ThinkTime string `yaml:"thinkTime" env-default:"10ms"`
	} `yaml:"primitiveFaultInjection"`
	PrimitiveDataIntegrity struct {
		VolMb int `yaml:"volMb" env-default:"1024"`
	} `yaml:"primitiveDataIntegrity"`
	MsvRebuild struct {
		Replicas       int    `yaml:"replicas" env-default:"1"`
		UpdatedReplica int    `yaml:"updatedreplica" env-default:"2"`
		VolSize        int    `yaml:"volSize" env-default:"50"`
		Timeout        string `yaml:"timeout" env-default:"120s"`
		PollPeriod     string `yaml:"pollPeriod" env-default:"1s"`
		DurationSecs   int    `yaml:"durationSecs" env-default:"180"`
		SleepSecs      int    `yaml:"sleepSecs" env-default:"3"`
	} `yaml:"msvRebuild"`
	PrimitiveMsvFuzz struct {
		VolMb               int    `yaml:"volMb" env-default:"64"`
		VolumeCountPerPool  int    `yaml:"volumeCountPerPool" env-default:"2"`
		Iterations          int    `yaml:"iterations" env-default:"2"`
		Replicas            int    `yaml:"replicas" env-default:"1"`
		InvalidReplicaCount int    `yaml:"invalidReplicaCount" env-default:"-1"`
		UnsupportedProtocol string `yaml:"unsupportedProtocol" env-default:"xyz"`
		UnsupportedFsType   string `yaml:"unsupportedFsType" env-default:"xyz"`
		IncorrectScName     string `yaml:"incorrectScName" env-default:"xyz"`
		LargePvcSize        int    `yaml:"largePvcSize" env-default:"11000000000000"`
		VolCount            int    `yaml:"volCount" env-default:"115"`
	} `yaml:"primitiveMsvFuzz"`
	FsxExt4Stress struct {
		VolMb             int    `yaml:"volMb" env-default:"1024"`
		Replicas          int    `yaml:"replicas" env-default:"3"`
		FileSystemType    string `yaml:"fileSystemType" env-default:"ext4"`
		NumberOfOperation int    `yaml:"numberOfOperation" env-default:"9977777"`
	} `yaml:"fsxExt4Stress"`
	PvcCreateDelete struct {
		Replicas         int `yaml:"replicas" env-default:"3"`
		VolSize          int `yaml:"volMb" env-default:"20"`
		Iterations       int `yaml:"iterations" env-default:"1"`
		VolumeMultiplier int `yaml:"volumeMultiplier" env-default:"110"`
		DelayTime        int `yaml:"delayTime" env-default:"10"`
	} `yaml:"pvcCreateDelete"`
	ScIscsiValidation struct {
		VolMb               int    `yaml:"volMb" env-default:"1024"`
		Replicas            int    `yaml:"replicas" env-default:"1"`
		UnsupportedProtocol string `yaml:"unsupportedProtocol" env-default:"iscsi"`
	} `yaml:"scIscsiValidation"`
}

var once sync.Once
var e2eConfig E2EConfig

// This function is called early from junit and various bits have not been initialised yet
// so we cannot use logf or Expect instead we use fmt.Print... and panic.
func GetConfig() E2EConfig {

	once.Do(func() {
		var err error
		var info os.FileInfo
		e2eRootDir, haveE2ERootDir := os.LookupEnv("e2e_root_dir")
		if !haveE2ERootDir {
			// try to work out the root directory of the mayastor-e2e repo so
			// that configuration file loading will work
			cwd, err := os.Getwd()
			if err == nil {
				comps := strings.Split(cwd, "/")
				for ix, comp := range comps {
					// Expect mayastor-e2e/src/...
					if comp == "mayastor-e2e" && len(comps[ix:]) > 2 && comps[ix+1] == "src" {
						candidate := path.Clean("/" + strings.Join(comps[:ix+1], "/"))
						info, err = os.Stat(candidate)
						if err == nil && info.IsDir() {
							fmt.Printf("Setting e2eRootDir to %v\n", candidate)
							e2eRootDir = candidate
							haveE2ERootDir = true
							break
						} else {
							fmt.Printf("Unable to stat path %v error:%v\n", candidate, err)
						}
					}
				}
			}
		}

		// Initialise the configuration
		_ = cleanenv.ReadEnv(&e2eConfig)
		e2eConfig.IOSoakTest.FioDutyCycles = []struct {
			ThinkTime       int `yaml:"thinkTime"`
			ThinkTimeBlocks int `yaml:"thinkTimeBlocks"`
		}{
			{500000, 1000},
			{750000, 1000},
			{1250000, 2000},
			{1500000, 3000},
			{1750000, 3000},
			{2000000, 4000},
		}

		// We absorb the complexity of locating the configuration file here
		// so that scripts invoking the tests can be simpler
		// - if OS envvar e2e_config is defined
		//		- if it is a path to a file then that file is used as the config file
		//		- else try to use a file of the same name in the configuration directory
		if e2eConfig.ConfigPaths.ConfigFile == "" {
			fmt.Println("Configuration file not specified, using defaults.")
			fmt.Println("	Use environment variable \"e2e_config_file\" to specify configuration file.")
		} else {
			var configFile string = path.Clean(e2eConfig.ConfigPaths.ConfigFile)
			info, err = os.Stat(configFile)
			if os.IsNotExist(err) && haveE2ERootDir {
				configFile = path.Clean(e2eRootDir + ConfigDir + "/" + e2eConfig.ConfigPaths.ConfigFile)
				info, err = os.Stat(configFile)
				if err != nil {
					panic(fmt.Sprintf("Unable to access configuration file %v", configFile))
				}
				e2eConfig.ConfigPaths.ConfigFile = configFile
			}
			if info.IsDir() {
				panic(fmt.Sprintf("%v is not a file", configFile))
			}
			fmt.Printf("Using configuration file %s\n", configFile)
			err = cleanenv.ReadConfig(configFile, &e2eConfig)
			if err != nil {
				panic(fmt.Sprintf("%v", err))
			}
		}

		if e2eConfig.ConfigPaths.PlatformConfigFile == "" {
			fmt.Println("Platform configuration file not specified, using defaults.")
			fmt.Println("	Use environment variable \"e2e_platform_config_file\" to specify platform configuration.")
		} else {
			var platformCfg string = path.Clean(e2eConfig.ConfigPaths.PlatformConfigFile)
			info, err = os.Stat(platformCfg)
			if os.IsNotExist(err) && haveE2ERootDir {
				platformCfg = path.Clean(e2eRootDir + PlatformConfigDir + e2eConfig.ConfigPaths.PlatformConfigFile)
				info, err = os.Stat(platformCfg)
				if err != nil {
					panic(fmt.Sprintf("Unable to access platform configuration file %v", platformCfg))
				}
				e2eConfig.ConfigPaths.PlatformConfigFile = platformCfg
			}
			if info.IsDir() {
				panic(fmt.Sprintf("%v is not a file", platformCfg))
			}
			fmt.Printf("Using platform configuration file %s\n", platformCfg)
			err = cleanenv.ReadConfig(platformCfg, &e2eConfig)
			if err != nil {
				panic(fmt.Sprintf("%v", err))
			}
		}

		// MayastorRootDir is either set from the environment variable
		// e2e_mayastor_root_dir or is set in the configuration file.
		if e2eConfig.MayastorRootDir == "" {
			fmt.Println("WARNING: mayastor directory not specified, install and uninstall tests will fail!")
		}

		artifactsDir := ""
		// if e2e root dir was specified record this in the configuration
		if haveE2ERootDir {
			e2eConfig.E2eRootDir = e2eRootDir
			// and setup the artifacts directory
			artifactsDir = path.Clean(e2eRootDir + "/artifacts")
		} else {
			// use the tmp directory for artifacts
			artifactsDir = path.Clean("/tmp/mayastor-e2e")
		}
		fmt.Printf("artifacts directory is %s\n", artifactsDir)

		if e2eConfig.SessionDir == "" {
			// The session directory is required for install and uninstall tests
			// create and use the default one.
			e2eConfig.SessionDir = artifactsDir + "/sessions/default"
			err = os.MkdirAll(e2eConfig.SessionDir, os.ModeDir|os.ModePerm)
			if err != nil {
				panic(err)
			}
		}
		fmt.Printf("session directory is %s\n", e2eConfig.SessionDir)

		if e2eConfig.ReportsDir == "" {
			fmt.Println("junit report files will not be generated.")
			fmt.Println("		Use environment variable \"e2e_reports_dir\" to specify a path for the report directory")
		} else {
			fmt.Printf("reports directory is %s\n", e2eConfig.ReportsDir)
		}
		saveConfig()
	})

	return e2eConfig
}

func saveConfig() {
	cfgBytes, _ := yaml.Marshal(e2eConfig)
	cfgUsedFile := path.Clean(e2eConfig.SessionDir + "/resolved-configuration-" + e2eConfig.ConfigName + "-" + e2eConfig.Platform.Name + ".yaml")
	err := ioutil.WriteFile(cfgUsedFile, cfgBytes, 0644)
	if err == nil {
		fmt.Printf("Resolved config written to %s\n", cfgUsedFile)
	} else {
		fmt.Printf("Resolved config not written to %s\n%v\n", cfgUsedFile, err)
	}
}

// SetControlPlane sets the control plane configuration if it is unset (i.e. empty) and writes it out if changed.
// If config setting matches  the existing value no action.
// Returns true it the config control plane value matches the input value
func SetControlPlane(controlPlane string) bool {
	_ = GetConfig()
	if e2eConfig.MayastorVersion == "" || e2eConfig.MayastorVersion == controlPlane {
		e2eConfig.MayastorVersion = controlPlane
		saveConfig()
		return true
	} else {
		fmt.Printf("Unable to override config control plane from '%s' to '%s'",
			e2eConfig.MayastorVersion, controlPlane)
	}
	return false
}
