package common

const NSE2EPrefix = "e2e-maya"
const NSDefault = "default"
const NSMayastor = "mayastor"
const CSIProvisioner = "io.openebs.csi-mayastor"
const DefaultVolumeSizeMb = 64
const DefaultFioSizeMb = 50

//  These variables match the settings used in createFioPodDef
const FioFsMountPoint = "/volume"
const FioBlockFilename = "/dev/sdm"
const FioFsFilename = FioFsMountPoint + "/fiotestfile"

// Relative path to the configuration directory WRT e2e root
const ConfigDir = "/configurations"
const DefaultConfigFileRelPath = ConfigDir + "/mayastor_ci_hcloud_e2e_config.yaml"
