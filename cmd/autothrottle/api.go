package main

import (
	"fmt"
	"io"
	"log"
	"net/http"

	"github.com/DataDog/topicmappr/kafkazk"
)

type APIConfig struct {
	Listen   string
	ZKPrefix string
}

var (
	// Misc. things.
	rateSettingsZNode        = "override_rate"
	incorrectMethod   string = "disallowed method\n"
)

func initAPI(c *APIConfig, zk *kafkazk.ZK) {
	p := fmt.Sprintf("/%s/%s", c.ZKPrefix, rateSettingsZNode)
	m := http.NewServeMux()

	err := zk.InitRawClient()
	if err != nil {
		log.Fatal(err)
	}

	// Check ZK for znode.
	exists, err := zk.Exists(p)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(exists)

	if !exists {
		err = zk.Create(p, "")
	}

	m.HandleFunc("/get_throttle", func(w http.ResponseWriter, req *http.Request) { getThrottle(w, req, zk, p) })
	m.HandleFunc("/set_throttle", func(w http.ResponseWriter, req *http.Request) { setThrottle(w, req, zk, p) })

	go func() {
		err := http.ListenAndServe(c.Listen, m)
		if err != nil {
			log.Fatal(err)
		}
	}()
}

func getThrottle(w http.ResponseWriter, req *http.Request, zk *kafkazk.ZK, p string) {
	logReq(req)
	if req.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		io.WriteString(w, incorrectMethod)
		return
	}
}

func setThrottle(w http.ResponseWriter, req *http.Request, zk *kafkazk.ZK, p string) {
	logReq(req)
	if req.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		io.WriteString(w, incorrectMethod)
		return
	}

	rate := req.URL.Query().Get("rate")
	if rate == "" {
		io.WriteString(w, "Rate param must be supplied\n")
		return
	}

	err := zk.Set(p, rate)
	if err != nil {
		errS := fmt.Sprintf("Error setting throttle: %s\n", err)
		io.WriteString(w, errS)
	} else {
		resp := fmt.Sprintf("Throttle set to %s\n", rate)
		io.WriteString(w, resp)
	}
}

func logReq(req *http.Request) {
	log.Printf("[API] %s %s %s\n", req.Method, req.RequestURI, req.RemoteAddr)
}
