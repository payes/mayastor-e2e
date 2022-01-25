package tc

import (
	"fmt"
	"os"
	"sync"

	logf "sigs.k8s.io/controller-runtime/pkg/log"

	"gopkg.in/yaml.v2"

	"github.com/ilyakaznacheev/cleanenv"
)

const gConfigFile = "/config.yaml"

type ReplicaElimination struct {
	FsVolume        int    `yaml:"fsvolume" env-default:"0"`
	LocalVolume     int    `yaml:"localvolume" env-default:"0"`
	Replicas        int    `yaml:"replicas" env-default:"3"`
	ThinkTime       int    `yaml:"thinkTime" env-default:"500000"`
	ThinkTimeBlocks int    `yaml:"thinkTimeBlocks" env-default:"1000"`
	Timeout         string `yaml:"timeout" env-default:"5m"`
	VolumeSizeMb    int    `yaml:"volumeSizeMb" env-default:"512"`
}

// E2EConfig is a application configuration structure
type ExtendedTestConfig struct {
	RunName  string `yaml:"RunName" env-default:"unnamed" env:"RUNNAME"` // human-readable test instance name for logging
	TestName string `yaml:"testName" env-default:"default"`              // selected test to run

	// FIXME: handle empty poolDevice
	Duration     string `yaml:"duration" env-default:"60m" env:"DURATION"`
	E2eFioImage  string `yaml:"e2eFioImage" env-default:"mayadata/e2e-fio" env:"e2e_fio_image"`
	Msnodes      int    `yaml:"msnodes" env-default:"3" env:"e2e_msnodes"`
	PoolDevice   string `yaml:"poolDevice" env:"e2e_pool_device"`
	SendEvent    int    `yaml:"sendEvent" env-default:"1" env:"SENDEVENT"`
	SendXrayTest int    `yaml:"sendXrayTest" env-default:"1" env:"SENDXRAYTEST"`
	XrayTestID   string `yaml:"test" env:"e2e_test"`

	// Individual Test parameters
	SteadyState struct {
		Replicas        int `yaml:"replicas" env-default:"2"`
		ThinkTime       int `yaml:"thinkTime" env-default:"500000"`
		ThinkTimeBlocks int `yaml:"thinkTimeBlocks" env-default:"1000"`
		VolumeSizeMb    int `yaml:"volumeSizeMb" env-default:"64"`
	} `yaml:"steadyState"`

	NonSteadyState struct {
		ConcurrentVols  int    `yaml:"concurrentvols" env-default:"1"`
		Replicas        int    `yaml:"replicas" env-default:"2"`
		ThinkTime       int    `yaml:"thinkTime" env-default:"500000"`
		ThinkTimeBlocks int    `yaml:"thinkTimeBlocks" env-default:"1000"`
		Timeout         string `yaml:"timeout" env-default:"5m"`
		VolumeSizeMb    int    `yaml:"volumeSizeMb" env-default:"64"`
	} `yaml:"nonSteadyState"`

	ReplicaPerturbation struct {
		FsVolume                  int `yaml:"fsvolume" env-default:"0"`
		LocalVolume               int `yaml:"localvolume" env-default:"0"`
		OfflineDeviceTest         int `yaml:"offlineDeviceTest" env-default:"0"`
		OfflineDevAndReplicasTest int `yaml:"offlineDevAndReplicasTest" env-default:"0"`
		Replicas                  int `yaml:"replicas" env-default:"3"`
		ThinkTime                 int `yaml:"thinkTime" env-default:"500000"`
		ThinkTimeBlocks           int `yaml:"thinkTimeBlocks" env-default:"1000"`
		VolumeSizeMb              int `yaml:"volumeSizeMb" env-default:"512"`
	} `yaml:"replicaPerturbation"`

	ReplicaElimination ReplicaElimination `yaml:"replicaElimination"`
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
		logf.Log.Info("Config", "Using file", gConfigFile)
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
