package tc

import (
	"fmt"
	"os"
	"sync"

	"gopkg.in/yaml.v2"

	"github.com/ilyakaznacheev/cleanenv"
)

const gConfigFile = "/config.yaml"

// E2EConfig is a application configuration structure
type ExtendedTestConfig struct {
	ConfigName string `yaml:"configName" env-default:"default"`

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
	// If set to true - typically during test development - tests with multiple It clauses should defer asserts till after
	// resources have been cleaned up . This behaviour makes it possible to have useful runs for all It clauses.
	// Typically set to false for CI test execution - no cleanup after first failure, as a result subsequent It clauses
	// in the test will fail the BeforeEach check, rendering post-mortem checks on the cluster more useful.
	// It may be set to true for when we want maximum test coverage, and post-mortem analysis is a secondary requirement.
	// NOTE: Only some tests support this feature.
	DefaultReplicaCount int `yaml:"defaultReplicaCount" env-default:"2" env:"e2e_default_replica_count"`
	// Timeout for MOAC CR state reconciliation in seconds, some CR state is not update promptly for example pool usage
	// and finalizers. On hcloud the time lag between synchronisation has been observed to be in the order of
	// a minute.
	MoacSyncTimeoutSeconds int `yaml:"moacSyncTimeoutSeconds" env-default:"600"`

	// Individual Test parameters
	SteadyStateTest struct {
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
}

var once sync.Once
var gConfig ExtendedTestConfig

// This function is called early from junit and various bits have not been initialised yet
// so we cannot use logf or Expect instead we use fmt.Print... and panic.
func GetConfig() (ExtendedTestConfig, error) {
	var err error

	once.Do(func() {
		var info os.FileInfo

		// Initialise the configuration
		_ = cleanenv.ReadEnv(&gConfig)
		gConfig.SteadyStateTest.FioDutyCycles = []struct {
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
		info, err = os.Stat(gConfigFile)
		if os.IsNotExist(err) {
			err = fmt.Errorf("Unable to access configuration file %v", gConfigFile)
			return
		} else {
			if info.IsDir() {
				err = fmt.Errorf("%v is not a file", gConfigFile)
				return
			}
		}
		fmt.Printf("Using configuration file %s\n", gConfigFile)
		err = cleanenv.ReadConfig(gConfigFile, &gConfig)
		if err != nil {
			err = fmt.Errorf("could not read config file, error %v", err)
			return
		}
		cfgBytes, _ := yaml.Marshal(gConfig)
		fmt.Printf("%s\n", string(cfgBytes))
	})

	return gConfig, err
}
