package cp_openapi

type CPv1oa struct {
	oa oaClient
}

func (cp CPv1oa) Version() string {
	return "1.0.0"
}

func (cp CPv1oa) MajorVersion() int {
	return 1
}

func (cp CPv1oa) VolStateHealthy() string {
	return "Online"
}

func (cp CPv1oa) VolStateDegraded() string {
	return "Degraded"
}

func (cp CPv1oa) ChildStateOnline() string {
	return "Online"
}

func (cp CPv1oa) ChildStateDegraded() string {
	return "Degraded"
}

func (cp CPv1oa) ChildStateUnknown() string {
	return "Unknown"
}

func (cp CPv1oa) ChildStateFaulted() string {
	return "Faulted"
}

func (cp CPv1oa) NexusStateUnknown() string {
	return "Unknown"
}

func (cp CPv1oa) NexusStateOnline() string {
	return "Online"
}

func (cp CPv1oa) NexusStateDegraded() string {
	return "Degraded"
}

func (cp CPv1oa) NexusStateFaulted() string {
	return "Faulted"
}

func (cp CPv1oa) MspGrpcStateToCrdState(mspState int) string {
	switch mspState {
	case 0:
		return "Pending"
	case 1:
		return "Online"
	case 2:
		return "Degraded"
	case 3:
		return "Faulted"
	default:
		return "Offline"
	}
}

func MakeCP(addr *[]string) CPv1oa {
	return CPv1oa{
		oa: makeClient(addr),
	}
}
