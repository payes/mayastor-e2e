package common

import "mayastor-e2e/common/e2e_config"

func IsControlPlaneMcp() bool {
	//FIXME should we assert if not moac or mcp2 ?
	return e2e_config.GetConfig().ControlPlane == CpMcp2
}

func IsControlPlaneMoac() bool {
	//FIXME should we assert if not moac or mcp2 ?
	return e2e_config.GetConfig().ControlPlane == CpMoac
}
