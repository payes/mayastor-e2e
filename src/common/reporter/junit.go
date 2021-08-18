package reporter

import (
	"mayastor-e2e/common/e2e_config"

	. "github.com/onsi/ginkgo"
	"github.com/onsi/ginkgo/reporters"
)

func GetReporters(name string) []Reporter {
	cfg := e2e_config.GetConfig()

	if cfg.ReportsDir == "" {
		return []Reporter{}
	}
	testGroupPrefix := "e2e."
	xmlFileSpec := cfg.ReportsDir + "/" + testGroupPrefix + name + "-junit.xml"
	junitReporter := reporters.NewJUnitReporter(xmlFileSpec)
	return []Reporter{junitReporter}
}
