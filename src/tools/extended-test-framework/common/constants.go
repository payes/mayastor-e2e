package common

const NSE2EAgent = "e2e-agent"
const NSE2EPrefix = "e2e-maya"
const NSDefault = "default"
const CSIProvisioner = "io.openebs.csi-mayastor"
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

//Custom Resources
const CRDGroupName = "openebs.io"
const CRDPoolGroupVersion = "v1alpha1"
const CRDPoolsResourceName = "mayastorpools"
const CRDVolumeGroupVersion = "v1alpha1"
const CRDVolumesResourceName = "mayastorvolumes"
const CRDNodeGroupVersion = "v1alpha1"
const CRDNodesResourceName = "mayastornodes"

// Storageclass parameter keys
const ScProtocol = "protocol"
const ScFsType = "fsType"
const ScReplicas = "repl"
const ScLocal = "local"
const IOTimeout = "ioTimeout"

// Labels

const MayastorEngineLabel = "openebs.io/engine"
const MayastorEngineLabelValue = "mayastor"

//  These variables match the settings used in fsx pod definition

const FsxBlockFileName = "/dev/sdm"

// Mayastor kubectl plugin details
const KubectlMayastorPlugin = "kubectl-mayastor"
const PluginPort = "30011"
