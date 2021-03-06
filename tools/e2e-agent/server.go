package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"strings"

	"github.com/gorilla/mux"
)

type NodeList struct {
	Nodes []string `json:"nodes"`
}

type Device struct {
	Device string `json:"device"`
	Table  string `json:"table"`
}

type ControlledDevice struct {
	Device string `json:"device"`
	State  string `json:"state"`
}

func homePage(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "Welcome home!\n")
}

type CmdList struct {
	Cmd string `json:"cmd"`
}

const (
	InternalServerErrorCode      = 500
	UnprocessableEntityErrorCode = 422
)

func main() {
	if err := Setup(); err != nil {
		log.Fatal(err)
	}
	handleRequests()
}

func handleRequests() {
	podIP := os.Getenv("MY_POD_IP")
	restPort := os.Getenv("REST_PORT")
	router := mux.NewRouter().StrictSlash(true)
	router.HandleFunc("/", homePage)
	router.HandleFunc("/ungracefulReboot", ungracefulReboot).Methods("POST")
	router.HandleFunc("/gracefulReboot", gracefulReboot).Methods("POST")
	router.HandleFunc("/dropConnectionsFromNodes", dropConnectionsFromNodes).Methods("POST")
	router.HandleFunc("/acceptConnectionsFromNodes", acceptConnectionsFromNodes).Methods("POST")
	router.HandleFunc("/createFaultyDevice", createFaultyDevice).Methods("POST")
	router.HandleFunc("/exec", execCmd).Methods("POST")
	router.HandleFunc("/devicecontrol", controlDevice).Methods("POST")
	router.HandleFunc("/killmayastor", killMayastor).Methods("POST")
	log.Fatal(http.ListenAndServe(podIP+":"+restPort, router))
}

func ungracefulReboot(w http.ResponseWriter, r *http.Request) {
	go func() {
		if err := UngracefulReboot(); err != nil {
			log.Print(err)
		}
	}()
}

func gracefulReboot(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "Graceful reboots are not yet supported")
}

func dropConnectionsFromNodes(w http.ResponseWriter, r *http.Request) {
	var list NodeList
	d := json.NewDecoder(r.Body)
	if err := d.Decode(&list); err != nil {
		fmt.Fprint(w, err.Error())
	}
	if err := DropConnectionsFromNodes(list.Nodes); err != nil {
		w.WriteHeader(InternalServerErrorCode)
		fmt.Fprint(w, err.Error())
		return
	}
	fmt.Fprint(w, "Successfully stopped network services\n")
}

func acceptConnectionsFromNodes(w http.ResponseWriter, r *http.Request) {
	var list NodeList
	d := json.NewDecoder(r.Body)
	if err := d.Decode(&list); err != nil {
		fmt.Fprint(w, err.Error())
	}
	err := AcceptConnectionsFromNodes(list.Nodes)
	if err != nil {
		w.WriteHeader(InternalServerErrorCode)
		fmt.Fprint(w, err.Error())
		return
	}
	fmt.Fprint(w, "Successfully started network services\n")
}

func execCmd(w http.ResponseWriter, r *http.Request) {
	var cmdline CmdList
	var cmd *exec.Cmd
	d := json.NewDecoder(r.Body)
	if err := d.Decode(&cmdline); err != nil {
		fmt.Fprint(w, err.Error())
	}
	if len(cmdline.Cmd) == 0 {
		w.WriteHeader(UnprocessableEntityErrorCode)
		fmt.Fprint(w, "no command passed")
		return
	}
	cmdArgs := strings.Split(cmdline.Cmd, " ")
	cmdName := cmdArgs[0]
	if len(cmdArgs) > 1 {
		cmd = exec.Command(cmdName, cmdArgs[1:]...)
	} else {
		cmd = exec.Command(cmdName)
		log.Printf("%s\n", cmdName)
	}
	output, err := cmd.CombinedOutput()
	if err != nil {
		w.WriteHeader(InternalServerErrorCode)
		fmt.Fprint(w, err.Error())
	} else {
		fmt.Fprint(w, string(output))
	}
}

func createFaultyDevice(w http.ResponseWriter, r *http.Request) {
	var (
		device Device
		cmd    *exec.Cmd
	)
	d := json.NewDecoder(r.Body)
	if err := d.Decode(&device); err != nil {
		fmt.Fprint(w, err.Error())
	}
	if len(device.Device) == 0 {
		w.WriteHeader(UnprocessableEntityErrorCode)
		fmt.Fprint(w, "no device passed")
		return
	}
	if len(device.Table) == 0 {
		w.WriteHeader(UnprocessableEntityErrorCode)
		fmt.Fprint(w, "no table passed")
		return
	}
	f, err := os.Create("table")
	if err != nil {
		fmt.Println(err)
		return
	}
	_, err = f.WriteString(device.Table)
	if err != nil {
		fmt.Println(err)
		f.Close()
		return
	}
	err = f.Close()
	if err != nil {
		fmt.Println(err)
		return
	}
	devName := strings.Split(device.Device, "/")

	cmdStr := "dmsetup create" + " " + devName[2] + " " + "table"
	cmdArgs := strings.Split(cmdStr, " ")
	cmdName := cmdArgs[0]
	if len(cmdArgs) > 1 {
		cmd = exec.Command(cmdName, cmdArgs[1:]...)
	} else {
		cmd = exec.Command(cmdName)
	}
	output, err := cmd.CombinedOutput()
	if err != nil {
		w.WriteHeader(InternalServerErrorCode)
		fmt.Fprint(w, err.Error())
	} else {
		fmt.Fprint(w, string(output))
	}
}

func controlDevice(w http.ResponseWriter, r *http.Request) {
	var device ControlledDevice
	params := make([]string, 2)

	d := json.NewDecoder(r.Body)
	if err := d.Decode(&device); err != nil {
		fmt.Fprint(w, err.Error())
	}
	if len(device.Device) == 0 {
		w.WriteHeader(UnprocessableEntityErrorCode)
		fmt.Fprint(w, "no device passed")
		return
	}
	if device.State != "offline" && device.State != "running" {
		w.WriteHeader(UnprocessableEntityErrorCode)
		fmt.Fprint(w, "invalid state")
		return
	}

	cmdStr := "bash"
	params[0] = "-c"
	params[1] = "echo " + device.State + " > /host/sys/block/" + device.Device + "/device/state"

	cmd := exec.Command(cmdStr, params...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		w.WriteHeader(InternalServerErrorCode)
		fmt.Fprint(w, err.Error())
		return
	}
	fmt.Fprint(w, string(output))
}

func killMayastor(w http.ResponseWriter, r *http.Request) {
	params := make([]string, 2)

	cmdStr := "bash"
	params[0] = "-c"
	params[1] = "MS=$(pidof mayastor) && kill -9 $MS"

	cmd := exec.Command(cmdStr, params...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		w.WriteHeader(InternalServerErrorCode)
		fmt.Fprint(w, err.Error())
		return
	}
	fmt.Fprint(w, string(output))
}
