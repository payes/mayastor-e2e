package common

const NSE2EAgent = "e2e-agent"
const NSE2EPrefix = "e2e-maya"
const NSDefault = "default"
const DefaultIOTimeout = 60
const DefaultVolumeSizeMb = 64
const DefaultFioSizeMb = 50

const SmallClaimSizeMb = 64
const LargeClaimSizeMb = 500

//  These variables match the settings used in createFioPodDef

const FioFsMountPoint = "/volume"
const FioBlockFilename = "/dev/sdm"
const FioFsFilename = FioFsMountPoint + "/fiotestfile"

// ConfigDir  Relative path to the configuration directory WRT e2e root.
// See common/e2e_config/e2e_config.go

// DefaultConfigFileRelPath  Relative path to default configuration file.
// See common/e2e_config/e2e_config.go

// Storageclass parameter keys
const ScProtocol = "protocol"
const ScFsType = "fsType"
const ScReplicas = "repl"
const ScLocal = "local"
const IOTimeout = "ioTimeout"

//  These variables match the settings used in fsx pod definition

const FsxBlockFileName = "/dev/sdm"
