package types

type Platform interface {
	PowerOnNode(node string) error
	PowerOffNode(node string) error
	GetNodeStatus(node string) (string, error)
}
