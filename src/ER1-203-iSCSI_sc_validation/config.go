package iscsi_sc_validation

import (
	"mayastor-e2e/common"
	"mayastor-e2e/common/e2e_config"
)

var defTimeoutSecs = "360s" // timeout in seconds

type ScIscsiValidationConfig struct {
	Protocol   common.ShareProto
	FsType     common.FileSystemType
	Replicas   int
	ScName     string
	PvcName    string
	PvName     string
	FioPodName string
	PvcSize    int
	Uuid       string
}

func GenerateScIscsiValidationConfig(testName string) *ScIscsiValidationConfig {
	params := e2e_config.GetConfig().ScIscsiValidation
	c := &ScIscsiValidationConfig{
		Protocol:   common.ShareProto(params.UnsupportedProtocol),
		FsType:     common.Ext4FsType,
		PvcSize:    params.VolMb,
		Replicas:   params.Replicas,
		ScName:     testName + "-sc",
		PvcName:    testName + "-pvc",
		FioPodName: testName + "-fio",
	}
	return c
}
