package main

import (
	"io/ioutil"
	"log"
	"os"
	"time"

	"os/exec"
)

const (
	SYSRQ_TRIGGER_FILE = "/host/proc/sysrq-trigger"
)

func Setup() error {
	var (
		err error
		cmd *exec.Cmd
	)

	e2eHostAddr := os.Getenv("E2E_HOST_ADDR")
	ports := []string{os.Getenv("REST_PORT"), os.Getenv("MAYASTOR_PORT")}
	for _, port := range ports {
		if e2eHostAddr != "" {
			cmd = exec.Command(
				"iptables", "-t", "mangle", "-i", "eth0", "-s", e2eHostAddr,
				"-I", "PREROUTING", "-p", "tcp", "--dport", port, "-j", "ACCEPT",
				"-m", "comment", "--comment", "mayastor-e2e-test")
		} else {
			cmd = exec.Command("iptables", "-t", "mangle", "-i", "eth0",
				"-I", "PREROUTING", "-p", "tcp", "--dport", port, "-j", "ACCEPT",
				"-m", "comment", "--comment", "mayastor-e2e-test")

		}
		_, err = cmd.Output()
		if err != nil {
			return err
		}
	}
	return nil
}

// UngracefulReboot crashes and reboots the host machine
func UngracefulReboot() error {
	log.Printf("Rebooting node ungracefully")
	time.Sleep(2 * time.Second)
	data := []byte("c")
	err := ioutil.WriteFile(SYSRQ_TRIGGER_FILE, data, 0644)
	return err
}

// GracefulReboot reboots the host gracefully
// It is not yet supported
func GracefulReboot() error {
	/*
		// Send SIGTERM to all processes
		data := []byte("e")
		if err := ioutil.WriteFile(SYSRQ_TRIGGER_FILE, data, 0644); err != nil {
			return err
		}

		time.Sleep(5 * time.Second)

		// Send SIGKILL to remaining processes
		data = []byte("i")
		if err := ioutil.WriteFile(SYSRQ_TRIGGER_FILE, data, 0644); err != nil {
			return err
		}

		time.Sleep(5 * time.Second)

		// Reboot the node
		data = []byte("b")
		if err := ioutil.WriteFile(SYSRQ_TRIGGER_FILE, data, 0644); err != nil {
			return err
		}
	*/
	return nil
}

// DropConnectionsFromNodes creates rules to drop connections from other k8s nodes
func DropConnectionsFromNodes(nodes []string) error {
	log.Printf("Drop Connections from %v", nodes)
	for _, node := range nodes {
		// because we specify eth0 interface we can DROP even our IP w/o problems, because self communication uses lo
		cmd := exec.Command("iptables", "-t", "mangle", "-I", "PREROUTING", "-i", "eth0", "-s", node, "-j", "DROP", "-m", "comment", "--comment", "mayastor-e2e-test")
		_, err := cmd.Output()
		if err != nil {
			return err
		}
	}
	return nil
}

// AcceptConnectionsFromNodes removes the rules set by
// DropConnectionsFromNodes so that other k8s nodes can reach this node again
func AcceptConnectionsFromNodes(nodes []string) error {
	log.Printf("Accept Connections from %v", nodes)
	for _, node := range nodes {
		cmd := exec.Command("iptables", "-t", "mangle", "-D", "PREROUTING", "-i", "eth0", "-s", node, "-j", "DROP", "-m", "comment", "--comment", "mayastor-e2e-test")
		_, err := cmd.Output()
		if err != nil {
			return err
		}
	}
	return nil
}
