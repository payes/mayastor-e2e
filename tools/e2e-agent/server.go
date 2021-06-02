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
	router.HandleFunc("/exec", execCmd).Methods("POST")
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
		args := strings.Join(cmdArgs[1:], " ")
		cmd = exec.Command(cmdName, args)
		log.Printf("%s, %s\n", cmdName, args)
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
