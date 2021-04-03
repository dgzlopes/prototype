package main

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
)

type Send struct {
	Name   string   `json:"name"`
	Type   string   `json:"type"`
	Tags   []string `json:"tags"`
	Config string   `json:"config"`
}

type createConfigCmd struct {
	Endpoint string   `arg:"" help:"Prototype endpoint with format (scheme://hostname)"`
	Name     string   `arg:"" help:"Service name"`
	Type     string   `arg:"" help:"Config type (cds,lds)"`
	Path     string   `arg:"" help:"Path to the config file"`
	Tags     []string `arg:"" help:"Tags to add with format: name:val name:val" type:"tag" name:"tag"`
}

func (cmd *createConfigCmd) Run() error {
	file, _ := ioutil.ReadFile(cmd.Path)
	json_data, err := json.Marshal(Send{
		Name:   cmd.Name,
		Type:   cmd.Type,
		Tags:   cmd.Tags,
		Config: string(file),
	})
	if err != nil {
		return err
	}

	_, err = http.Post(cmd.Endpoint+"/api/config", "application/json", bytes.NewBuffer(json_data))
	if err != nil {
		return err
	}
	return nil
}
