package common

const NSE2EAgent = "e2e-agent"
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

// ConfigDir  Relative path to the configuration directory WRT e2e root.
const ConfigDir = "/configurations"

// DefaultConfigFileRelPath  Relative path to default configuration file.
const DefaultConfigFileRelPath = ConfigDir + "/mayastor_ci_hcloud_e2e_config.yaml"

//Custom Resources
const CRDGroupName = "openebs.io"
const CRDPoolGroupVersion = "v1alpha1"
const CRDPoolsResourceName = "mayastorpools"
const CRDVolumeGroupVersion = "v1alpha1"
const CRDVolumesResourceName = "mayastorvolumes"
const CRDNodeGroupVersion = "v1alpha1"
const CRDNodesResourceName = "mayastornodes"
