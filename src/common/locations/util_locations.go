package locations

// For now the relative paths are hardcoded, there may be a case to make this
// more generic and data driven.

import (
	"os"
	"path"

	"mayastor-e2e/common/e2e_config"

	. "github.com/onsi/gomega"
)

func locationExists(path string) string {
	_, err := os.Stat(path)
	Expect(err).To(BeNil(), "%s", err)
	return path
}

// GetBuildInfoFile returns the path to build_info.json if one exists.
// build_info.json is typically part of the install-bundle
func GetBuildInfoFile() (string, error) {
	filePath := path.Clean(e2e_config.GetConfig().MayastorRootDir + "/scripts/../build_info.json")
	_, err := os.Stat(filePath)
	if err != nil {
		return "", err
	}
	return filePath, err
}

func GetMayastorScriptsDir() string {
	return locationExists(path.Clean(e2e_config.GetConfig().MayastorRootDir + "/scripts"))
}

func GetControlPlaneScriptsDir() string {
	return locationExists(path.Clean(e2e_config.GetConfig().MayastorRootDir + "/mcp/scripts"))
}

// GetGeneratedYamlsDir return the path to where Mayastor yaml files are generated this is a generated directory, so may not exist yet.
func GetGeneratedYamlsDir() string {
	return path.Clean(e2e_config.GetConfig().SessionDir + "/install-yamls")
}

// GetControlPlaneGeneratedYamlsDir return the path to where Mayastor yaml files are generated this is a generated directory, so may not exist yet.
func GetControlPlaneGeneratedYamlsDir() string {
	return path.Clean(e2e_config.GetConfig().SessionDir + "/install-yamls-control-plane")
}

// GetE2EAgentPath return the path e2e-agent yaml file
func GetE2EAgentPath() string {
	return path.Clean(e2e_config.GetConfig().E2eRootDir + "/tools/e2e-agent")
}

// GetE2EScriptsPath return the path e2e-agent yaml file
func GetE2EScriptsPath() string {
	return path.Clean(e2e_config.GetConfig().E2eRootDir + "/scripts")
}
