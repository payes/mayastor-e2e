package k8stest

import (
	"fmt"
	"strings"
	"time"

	"mayastor-e2e/common/custom_resources"
	agent "mayastor-e2e/common/e2e-agent"
	"mayastor-e2e/common/mayastorclient"

	logf "sigs.k8s.io/controller-runtime/pkg/log"
)

// ChecksumReplica checksums the partition p2 of the nvme target defined by the given uri
// It uses the e2e agent and the nvme client to connect to the target.
// the returned format is that generated by cksum: <checksum> <size> <device>
// e.g. "924018992 61849088 /dev/nvme0n1p2"
func ChecksumReplica(initiatorIP string, targetIP string, uri string) (string, error) {
	logf.Log.Info("checksumReplica", "nexusIP", initiatorIP, "nodeIP", targetIP, "uri", uri)

	nqnoffset := strings.Index(uri, "nqn.")
	nqnlong := uri[nqnoffset:]
	tailoffset := strings.Index(nqnlong, "?")
	nqn := nqnlong[:tailoffset]

	cmdArgs := []string{
		"nvme",
		"connect",
		"-a", targetIP,
		"-t", "tcp",
		"-s", "8420",
		"-n", nqn,
	}
	args := strings.Join(cmdArgs, " ")
	resp, err := agent.Exec(initiatorIP, args)
	resp = strings.TrimSpace(resp)
	if err != nil {
		logf.Log.Info("Running agent failed", "error", err)
		return "", err
	}
	if resp != "" { // connect should be silent
		return "", fmt.Errorf("nvme connect returned with %s", resp)
	}

	// nvme list returns all nvme devices in form n * [ ... <Node> ... <Model> ... \n]
	// we want the device (=Node) associated with Model = "Mayastor NVMe controller"
	// Node is typically "/dev/nvme0n1" and the partition "nvme0n1p2"
	devicePath, err := agent.Exec(initiatorIP, "nvme list")

	if err != nil {
		logf.Log.Info("Running agent failed", "error", err)
		return "", err
	}

	// from the string find the part containing "Mayastor NVMe controller"
	idx := strings.Index(devicePath, "Mayastor NVMe controller")
	if idx == -1 {
		return "", fmt.Errorf("Failed to find mayastor target")
	}
	devicePath = devicePath[:idx]

	// find the last device up to that point
	idx = strings.LastIndex(devicePath, "/dev/nvme")
	if idx == -1 {
		return "", fmt.Errorf("Failed to find mayastor device")
	}
	devicePath = devicePath[idx:]

	// extract up to the first space and add the partition suffix
	idx = strings.Index(devicePath, " ")
	if idx != -1 {
		devicePath = devicePath[:idx]
	}
	devicePath = devicePath + "p2"
	deviceOnly := devicePath[5:] // remove the /dev/ prefix

	// checksum the device
	// the returned format is <checksum> <size> <device>
	// e.g. "924018992 61849088 /dev/nvme0n1p2"
	args = "cksum " + devicePath
	cksumText, err := agent.Exec(initiatorIP, args)
	if err != nil {
		logf.Log.Info("Running agent failed", "error", err)
		return "", err
	}
	cksumText = strings.TrimSpace(cksumText)
	logf.Log.Info("Executed", "cmd", args, "got", cksumText)
	// double check the response contains the device name
	if !strings.Contains(cksumText, deviceOnly) {
		return "", fmt.Errorf("Unexpected result from cksum %v", cksumText)
	}

	args = "nvme disconnect -n" + nqn
	resp, err = agent.Exec(initiatorIP, args)
	if err != nil {
		logf.Log.Info("Running agent failed", "error", err)
		return "", err
	}
	logf.Log.Info("Executing", "cmd", args, "got", resp)

	// check that the device no longer exists
	resp, err = agent.Exec(initiatorIP, "ls /dev/")
	if err != nil {
		logf.Log.Info("Running agent failed", "error", err)
		return "", err
	}
	if strings.Contains(resp, deviceOnly) {
		return "", fmt.Errorf("Device %s still exists", deviceOnly)
	}
	return cksumText, nil
}

// ExcludeNexusReplica - ensure the volume has no nexus-local replica
// This depends on there being an unused mayastor instance available so
// e.g. a 2-replica volume needs at least a 3-node cluster
func ExcludeNexusReplica(nexusIP string, uuid string) (bool, error) {
	// get the nexus local device
	var nxlist []string
	nxlist = append(nxlist, nexusIP)
	nexusList, err := mayastorclient.ListNexuses(nxlist)
	if err != nil {
		return false, fmt.Errorf("Failed to list nexuses, err=%v", err)
	}
	if len(nexusList) == 0 {
		return false, fmt.Errorf("Expected to find at least 1 nexus")
	}

	nxChild := ""
	for _, nx := range nexusList {
		if nx.Uuid == uuid {
			for _, ch := range nx.Children {
				if strings.HasPrefix(ch.Uri, "bdev:///") {
					if nxChild != "" {
						return false, fmt.Errorf("More than 1 nexus local replica found")
					}
					nxChild = ch.Uri
				}
			}
			if nxChild == "" { // there is no local replica so we are done
				return false, nil
			}
			break
		}
	}
	if nxChild == "" {
		return false, fmt.Errorf("failed to find the nexus")
	}

	// fault the replica
	err = mayastorclient.FaultNexusChild(nexusIP, uuid, nxChild)
	if err != nil {
		return false, fmt.Errorf("Failed to fault child, err=%v", err)
	}

	// wait for the msv to become degraded
	state := ""
	const sleepTime = 10
	const timeOut = 120
	for ix := 0; ix < (timeOut-1)/sleepTime; ix++ {
		state, err = custom_resources.GetMsVolState(uuid)
		if err != nil {
			return false, fmt.Errorf("Failed to get state, err=%v", err)
		}
		if state == "degraded" {
			break
		}
		time.Sleep(sleepTime * time.Second)
	}
	if state != "degraded" {
		return true, fmt.Errorf("timeo out waiting for volume to become degraded")
	}

	// wait for the msv to become healthy - rebuilt with a non-nexus replica
	for ix := 0; ix < (timeOut-1)/sleepTime; ix++ {
		state, err = custom_resources.GetMsVolState(uuid)
		if err != nil {
			return false, fmt.Errorf("Failed to get state, err=%v", err)
		}
		if state == "healthy" {
			break
		}
		time.Sleep(sleepTime * time.Second)
	}
	if state != "healthy" {
		return true, fmt.Errorf("timeo out waiting for volume to become healthy")
	}
	return true, nil
}
