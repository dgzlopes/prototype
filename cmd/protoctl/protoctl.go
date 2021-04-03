package main

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
)

type Send struct {
	Name   string   `json:"name"`
	Type   string   `json:"type"`
	Tags   []string `json:"tags"`
	Config string   `json:"config"`
}

func main() {
	cds, _ := ioutil.ReadFile("/home/dgzlopes/go/src/github.com/dgzlopes/prototype/example/configs/cds.yaml")
	json_data, err := json.Marshal(Send{
		Name:   "quote",
		Type:   "cds",
		Tags:   []string{"env:production", "version:0.0.6-beta"},
		Config: string(cds),
	})

	if err != nil {
		log.Fatal(err)
	}

	_, err = http.Post("http://localhost:10000/api/config", "application/json",
		bytes.NewBuffer(json_data))

	if err != nil {
		log.Fatal(err)
	}
}
