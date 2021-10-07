package v1

type CPv1 struct {
	nodeIPAddresses *[]string
}

func (cp CPv1) Version() string {
	return "1.0.0"
}

func (cp CPv1) MajorVersion() int {
	return 1
}

func (cp CPv1) VolStateHealthy() string {
	return "Online"
}

func (cp CPv1) VolStateDegraded() string {
	return "Degraded"
}

func (cp CPv1) ChildStateOnline() string {
	return "Online"
}

func (cp CPv1) ChildStateDegraded() string {
	return "Degraded"
}

func (cp CPv1) ChildStateUnknown() string {
	return "Unknown"
}

func (cp CPv1) ChildStateFaulted() string {
	return "Faulted"
}

func (cp CPv1) NexusStateUnknown() string {
	return "Unknown"
}

func (cp CPv1) NexusStateOnline() string {
	return "Online"
}

func (cp CPv1) NexusStateDegraded() string {
	return "Degraded"
}

func (cp CPv1) NexusStateFaulted() string {
	return "Faulted"
}

func MakeCP(addr *[]string) CPv1 {
	return CPv1{
		nodeIPAddresses: addr,
	}
}
