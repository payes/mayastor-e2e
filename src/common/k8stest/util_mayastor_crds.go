package k8stest

// GetMSV Get pointer to a mayastor volume custom resource
// returns nil and no error if the msv is in pending state.
func GetMSV(uuid string) (*MayastorVolume, error) {
	return GetMsvIfc().getMSV(uuid)
}

// GetMsvNodes Retrieve the nexus node hosting the Mayastor Volume,
// and the names of the replica nodes
// function asserts if the volume CR is not found.
func GetMsvNodes(uuid string) (string, []string) {
	return GetMsvIfc().getMsvNodes(uuid)
}

func DeleteMsv(volName string) error {
	return GetMsvIfc().deleteMsv(volName)
}

func ListMsvs() ([]MayastorVolume, error) {
	return GetMsvIfc().listMsvs()
}

func SetMsvReplicaCount(uuid string, replicaCount int) error {
	return GetMsvIfc().setMsvReplicaCount(uuid, replicaCount)
}

func GetMsvState(uuid string) (string, error) {
	return GetMsvIfc().getMsvState(uuid)
}

func GetMsvReplicas(volName string) ([]Replica, error) {
	return GetMsvIfc().getMsvReplicas(volName)
}

func GetMsvNexusChildren(volName string) ([]NexusChild, error) {
	return GetMsvIfc().getMsvNexusChildren(volName)
}

func GetMsvNexusState(uuid string) (string, error) {
	return GetMsvIfc().getMsvNexusState(uuid)
}

func IsMsvPublished(uuid string) bool {
	return GetMsvIfc().isMsvPublished(uuid)
}

func IsMsvDeleted(uuid string) bool {
	return GetMsvIfc().isMsvDeleted(uuid)
}

func CheckForMsvs() (bool, error) {
	return GetMsvIfc().checkForMsvs()
}

func CheckAllMsvsAreHealthy() error {
	return GetMsvIfc().checkAllMsvsAreHealthy()
}
