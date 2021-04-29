package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/gorilla/mux"
)

type NodeList struct {
	Nodes []string `json:"nodes"`
}

func homePage(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "Welcome home!\n")
}

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
		fmt.Fprint(w, err.Error())
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
		fmt.Fprint(w, err.Error())
	}
	fmt.Fprint(w, "Successfully started network services\n")
}
