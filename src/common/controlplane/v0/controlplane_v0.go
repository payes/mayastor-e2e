package v0

type CPv0p8 struct{}

func (cp CPv0p8) Version() string {
	return "0.8.2"
}

func (cp CPv0p8) MajorVersion() int {
	return 0
}

func (cp CPv0p8) VolStateHealthy() string {
	return "healthy"
}

func (cp CPv0p8) VolStateDegraded() string {
	return "degraded"
}

func (cp CPv0p8) ChildStateOnline() string {
	return "CHILD_ONLINE"
}

func (cp CPv0p8) ChildStateDegraded() string {
	return "CHILD_DEGRADED"
}

func (cp CPv0p8) ChildStateUnknown() string {
	return "CHILD_UNKNOWN"
}

func (cp CPv0p8) ChildStateFaulted() string {
	return "CHILD_FAULTED"
}

func (cp CPv0p8) NexusStateUnknown() string {
	return "NEXUS_UNKNOWN"
}

func (cp CPv0p8) NexusStateOnline() string {
	return "NEXUS_ONLINE"
}

func (cp CPv0p8) NexusStateDegraded() string {
	return "NEXUS_DEGRADED"
}

func (cp CPv0p8) NexusStateFaulted() string {
	return "NEXUS_FAULTED"
}

func MakeCP() CPv0p8 {
	return CPv0p8{}
}
