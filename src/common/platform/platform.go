package platform

import (
	"mayastor-e2e/common/e2e_config"
	hcloudClient "mayastor-e2e/common/platform/hcloud"
	types "mayastor-e2e/common/platform/types"
)

func Create() types.Platform {
	cfg := e2e_config.GetConfig()
	switch cfg.E2ePlatform {
	case "Hetzner":
		return hcloudClient.New()
	}
	return nil
}
