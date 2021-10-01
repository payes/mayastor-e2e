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

func GetMayastorScriptsDir() string {
	return locationExists(path.Clean(e2e_config.GetConfig().MayastorRootDir + "/scripts"))
}

func GetMCPScriptsDir() string {
	return locationExists(path.Clean(e2e_config.GetConfig().MayastorRootDir + "/mcp/scripts"))
}

// GetGeneratedYamlsDir return the path to where Mayastor yaml files are generated this is a generated directory, so may not exist yet.
func GetGeneratedYamlsDir() string {
	return path.Clean(e2e_config.GetConfig().SessionDir + "/install-yamls")
}

// GetGeneratedYamlsDir return the path to where Mayastor yaml files are generated this is a generated directory, so may not exist yet.
func GetControlPlaneGeneratedYamlsDir() string {
	return path.Clean(e2e_config.GetConfig().SessionDir + "/install-yamls-control-plane")
}

// GetE2EAgentPath return the path e2e-agent yaml file
func GetE2EAgentPath() string {
	return path.Clean(e2e_config.GetConfig().E2eRootDir + "/tools/e2e-agent")
}
