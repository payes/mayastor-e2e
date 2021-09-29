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
	Test        string `yaml:"test" env:"e2e_test"`
	Install     bool   `yaml:"install" env-default:"false" env:"e2e_install"`

	// Individual Test parameters
	SteadyState struct {
		Replicas     int    `yaml:"replicas" env-default:"2"`
		Duration     string `yaml:"duration" env-default:"60m"`
		VolumeSizeMb int    `yaml:"volumeSizeMb" env-default:"64"`
	} `yaml:"steadyState"`
	ReplicaPerturbation struct {
		Replicas     int    `yaml:"replicas" env-default:"3"`
		Duration     string `yaml:"duration" env-default:"60m"`
		VolumeSizeMb int    `yaml:"volumeSizeMb" env-default:"64"`
	} `yaml:"replicaPerturbation"`
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
