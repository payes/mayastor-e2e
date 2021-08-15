package client

import (
	"fmt"
	"mayastor-e2e/common/platform/types"
	"os/exec"
	"strings"

	logf "sigs.k8s.io/controller-runtime/pkg/log"
)

type hcloud struct {
}

func New() types.Platform {
	return &hcloud{}
}

func (h *hcloud) PowerOffNode(node string) error {
	logf.Log.Info("Power off", "node", node)
	cmd := exec.Command("sh", "-c", fmt.Sprintf("hcloud server poweroff %s", node))
	_, err := cmd.Output()
	return err
}

func (h *hcloud) PowerOnNode(node string) error {
	logf.Log.Info("Power on", "node", node)
	cmd := exec.Command("sh", "-c", fmt.Sprintf("hcloud server poweron %s", node))
	_, err := cmd.Output()
	return err
}

func (h *hcloud) RebootNode(node string) error {
	logf.Log.Info("Reboot", "node", node)
	cmd := exec.Command("sh", "-c", fmt.Sprintf("hcloud server reboot %s", node))
	_, err := cmd.Output()
	return err
}

func (h *hcloud) DetachVolume(volName string) error {
	logf.Log.Info("Detach Volume ", "volName", volName)
	cmd := exec.Command("sh", "-c", fmt.Sprintf("hcloud volume detach %s", volName))
	_, err := cmd.Output()
	return err
}

func (h *hcloud) AttachVolume(volName, node string) error {
	logf.Log.Info("Attach Volume to node", "volName", volName, "node", node)
	cmd := exec.Command("sh", "-c", fmt.Sprintf("hcloud volume attach %s --server %s", volName, node))
	_, err := cmd.Output()
	return err
}

func (h *hcloud) GetNodeStatus(node string) (string, error) {
	logf.Log.Info("Get status", "node", node)
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
