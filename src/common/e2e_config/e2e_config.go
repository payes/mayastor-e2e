package e2e_config

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"sync"

	"gopkg.in/yaml.v2"

	"github.com/ilyakaznacheev/cleanenv"
)

const ConfigDir = "/configurations"
const DefaultConfigFileRelPath = ConfigDir + "/mayastor_ci_hcloud_e2e_config.yaml"
const PlatformConfigDir = "/configurations/platforms/"

// E2EConfig is a application configuration structure
type E2EConfig struct {
	ConfigName string `yaml:"configName"`
	Platform   struct {
		// E2ePlatform indicates where the e2e is currently being run from
		Name string `yaml:"name"`
		// Add HostNetwork: true to the spec of test pods.
		HostNetworkingRequired bool `yaml:"hostNetworkingRequired" env-default:"false"`
		// Some deployments use a different namespace
		MayastorNamespace string `yaml:"mayastorNamespace" env-default:"mayastor"`
		// Some deployments use a different namespace
		FilteredMayastorPodCheck int `yaml:"filteredMayastorPodCheck" env-default:"0"`
		// FIXME: temporary for volterra Do not use e2e-agent
		DisableE2EAgent bool `yaml:"disableE2EAgent" env-default:"false"`
	} `yaml:"platform"`

	// Generic configuration files used for CI and automation should not define MayastorRootDir and E2eRootDir
	MayastorRootDir string `yaml:"mayastorRootDir" env:"e2e_mayastor_root_dir"`
	E2eRootDir      string `yaml:"e2eRootDir" env:"e2e_root_dir"`
	// Operational parameters
	Cores int `yaml:"cores,omitempty"`
	// Registry from where mayastor images are retrieved
	Registry string `yaml:"registry" env:"e2e_docker_registry" env-default:"ci-registry.mayastor-ci.mayadata.io"`
	// Registry from where CI testing images are retrieved
	CIRegistry  string `yaml:"ciRegistry" env:"e2e_ci_docker_registry" env-default:"ci-registry.mayastor-ci.mayadata.io"`
	ImageTag    string `yaml:"imageTag" env:"e2e_image_tag" env-default:"ci"`
	PoolDevice  string `yaml:"poolDevice" env:"e2e_pool_device"`
	E2eFioImage string `yaml:"e2eFioImage" env-default:"mayadata/e2e-fio" env:"e2e_fio_image"`
	// This is an advisory setting for individual tests
	// If set to true - typically during test development - tests with multiple It clauses should defer asserts till after
	// resources have been cleaned up . This behaviour makes it possible to have useful runs for all It clauses.
	// Typically set to false for CI test execution - no cleanup after first failure, as a result subsequent It clauses
	// in the test will fail the BeforeEach check, rendering post-mortem checks on the cluster more useful.
	// It may be set to true for when we want maximum test coverage, and post-mortem analysis is a secondary requirement.
	// NOTE: Only some tests support this feature.
	DeferredAssert bool `yaml:"deferredAssert" env-default:"false" env:"e2e_defer_asserts"`
	// TODO: for now using a simple boolean for a specific behaviour suffices, a more sophisticated approach using a policy for test runs may be required.
	CleanupOnBeforeEach bool `yaml:"cleanupOnBeforeEach" env-default:"false" env:"e2e_policy_cleanup_before"`
	// Default replica count, used by tests which do not have a config section.
	DefaultReplicaCount int `yaml:"defaultReplicaCount" env-default:"2" env:"e2e_default_replica_count"`

	// Run configuration
	ReportsDir string `yaml:"reportsDir" env:"e2e_reports_dir"`

	// Individual Test parameters
	PVCStress struct {
		Replicas   int `yaml:"replicas" env-default:"1"`
		CdCycles   int `yaml:"cdCycles" env-default:"100"`
		CrudCycles int `yaml:"crudCycles" env-default:"20"`
	} `yaml:"pvcStress"`
	IOSoakTest struct {
		Replicas int    `yaml:"replicas" env-default:"2"`
		Duration string `yaml:"duration" env-default:"10m"`
		// Number of volumes for each mayastor instance
		// volumes for disruptor pods are allocated from within this "pool"
		LoadFactor int      `yaml:"loadFactor" env-default:"10"`
		Protocols  []string `yaml:"protocols" env-default:"nvmf"`
		// FioStartDelay units are seconds
		FioStartDelay int    `yaml:"fioStartDelay" env-default:"60"`
		ReadyTimeout  string `yaml:"readyTimeout" env-default:"300s"`
		Disrupt       struct {
			// Number of disruptor pods.
			PodCount int `yaml:"podCount" env-default:"3"`
			// FaultAfter units are seconds
			FaultAfter   int    `yaml:"faultAfter" env-default:"45"`
			ReadyTimeout string `yaml:"readyTimeout" env-default:"60s"`
		} `yaml:"disrupt"`
		FioDutyCycles []struct {
			// ThinkTime units are microseconds
			ThinkTime       int `yaml:"thinkTime"`
			ThinkTimeBlocks int `yaml:"thinkTimeBlocks"`
		} `yaml:"fioDutyCycles"`
	} `yaml:"ioSoakTest"`
	CSI struct {
		Replicas       int    `yaml:"replicas" env-default:"1"`
		SmallClaimSize string `yaml:"smallClaimSize" env-default:"50Mi"`
		LargeClaimSize string `yaml:"largeClaimSize" env-default:"500Mi"`
	} `yaml:"csi"`
	Uninstall struct {
		Cleanup int `yaml:"cleanup" env:"e2e_uninstall_cleanup"`
	} `yaml:"uninstall"`
	BasicVolumeIO struct {
		// FioTimeout is in seconds
		FioTimeout int `yaml:"fioTimeout" env-default:"120"`
		// VolSizeMb Units are MiB
		VolSizeMb int `yaml:"volSizeMb" env-default:"1024"`
		// FsVolSizeMb Units are MiB
		FsVolSizeMb int `yaml:"fsVolSizeMb" env-default:"900"`
	} `yaml:"basicVolumeIO"`
	MultipleVolumesPodIO struct {
		VolumeCount          int    `yaml:"volumeCount" env-default:"2"`
		MultipleReplicaCount int    `yaml:"replicas" env-default:"2"`
		Duration             string `yaml:"duration" env-default:"30s"`
		Timeout              string `yaml:"timeout" env-default:"60s"`
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
	} `yaml:"msPodDisruption"`
	MaximumVolsIO struct {
		VolMb             int    `yaml:"volMb" env-default:"64"`
		VolumeCountPerPod int    `yaml:"volumeCountPerPod" env-default:"10"`
		PodCount          int    `yaml:"podCount" env-default:"11"`
		Duration          string `yaml:"duration" env-default:"240s"`
		Timeout           string `yaml:"timeout" env-default:"360s"`
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
		Iterations             int    `yaml:"iterations" env-default:"100"`
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
}

var once sync.Once
var e2eConfig E2EConfig

// This function is called early from junit and various bits have not been initialised yet
// so we cannot use logf or Expect instead we use fmt.Print... and panic.
func GetConfig() E2EConfig {
	var err error
	e2eRootDir, okE2eRootDir := os.LookupEnv("e2e_root_dir")
	// The configuration overrides the e2eRootDir setting,
	// this makes it possible to use a configuration file written out
	// previously to replicate a test run configuration.
	once.Do(func() {
		var configFile string
		// We absorb the complexity of locating the configuration file here
		// so that scripts invoking the tests can be simpler
		// - if OS envvar e2e_config is defined
		//		- if it is a path to a file then that file is used as the config file
		//		- else try to use a file of the same name in the configuration directory
		// - Otherwise the config file is defaulted to ci_e2e_config
		// A configuration file *MUST* be specified.
		value, ok := os.LookupEnv("e2e_config_file")
		if !ok {
			panic("configuration file not specified, use env var e2e_config_file")
		}
		configFile = path.Clean(e2eRootDir + ConfigDir + "/" + value)
		fmt.Printf("Using configuration file %s\n", configFile)
		err = cleanenv.ReadConfig(configFile, &e2eConfig)
		if err != nil {
			panic(fmt.Sprintf("%v", err))
		}

		value, ok = os.LookupEnv("e2e_platform_config_file")
		if !ok {
			panic("Platform configuration file not specified, use env var e2e_platform_config_file")
		}
		platformCfg := path.Clean(e2eRootDir + PlatformConfigDir + value)
		fmt.Printf("Using platform configuration file %s\n", configFile)
		err = cleanenv.ReadConfig(platformCfg, &e2eConfig)
		if err != nil {
			panic(fmt.Sprintf("%v", err))
		}

		// There are complications because there are 2 possible sources for truth for the e2e root directory
		// 1. the environment variable
		// 2. the configuration file
		// If only one is defined, we use the defined value,
		// We need to resolve in a well defined manner when
		// a. neither are defined (panic)
		// b. both are defined, (environment variable overrides configuration setting)
		if !okE2eRootDir {
			if e2eConfig.E2eRootDir == "" {
				panic("E2E root directory is not specified.")
			}
		} else {
			if e2eRootDir != e2eConfig.E2eRootDir {
				fmt.Printf("overriding configuration e2e root dir from %s to %s", e2eConfig.E2eRootDir, e2eRootDir)
			}
			e2eConfig.E2eRootDir = e2eRootDir
		}

		// MayastorRootDir is either set from the environment var mayastor_root_dir
		// or is pre-configured in the configuration file.
		// It *cannot* be empty
		if e2eConfig.MayastorRootDir == "" {
			panic("Configuration error unspecified mayastor directory")
		}

		cfgBytes, _ := yaml.Marshal(e2eConfig)
		cfgUsedFile := path.Clean(e2eConfig.E2eRootDir + "/artifacts/used-" + e2eConfig.ConfigName + "-" + e2eConfig.Platform.Name + ".yaml")
		err = ioutil.WriteFile(cfgUsedFile, cfgBytes, 0644)
		if err == nil {
			fmt.Printf("Resolved config written to %s\n", cfgUsedFile)
		}
	})

	return e2eConfig
}
