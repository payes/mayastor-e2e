package client

import (
	"fmt"
	"mayastor-e2e/common/platform/types"
	"os/exec"
	"strings"
)

type hcloud struct {
}

func New() types.Platform {
	return &hcloud{}
}

func (h *hcloud) PowerOffNode(node string) error {
	cmd := exec.Command("bash", "-c", fmt.Sprintf("hcloud server poweroff %s", node))
	_, err := cmd.Output()
	return err
}

func (h *hcloud) PowerOnNode(node string) error {
	cmd := exec.Command("bash", "-c", fmt.Sprintf("hcloud server poweron %s", node))
	_, err := cmd.Output()
	return err
}

func (h *hcloud) GetNodeStatus(node string) (string, error) {
	cmd := exec.Command("bash", "-c", fmt.Sprintf("hcloud  server list | grep %s", node))
	stdout, err := cmd.Output()
	if err != nil {
		return "", err
	}
	if strings.Contains(string(stdout), "running") {
		return "running", nil
	}
	return "off", nil
}
