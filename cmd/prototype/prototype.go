package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"strings"

	badger "github.com/dgraph-io/badger/v3"
	"github.com/gorilla/mux"
)

type Client struct {
	db *badger.DB
}

// Set the given key with the given value.
func (c *Client) Set(key string, value []byte) error {
	err := c.db.Update(func(txn *badger.Txn) error {
		err := txn.Set([]byte(key), value)
		return err
	})
	return err
}

// Get returns the value for the given key.
func (c *Client) Get(key string) string {
	var valCopy []byte
	_ = c.db.View(func(txn *badger.Txn) error {
		item, _ := txn.Get([]byte(key))
		_ = item.Value(func(val []byte) error {
			valCopy = append([]byte{}, val...)
			return nil
		})
		return nil
	})
	return string(valCopy)
}

// GetByPrefix return all the key value pairs where the key starts with some prefix.
func (c *Client) GetByPrefix(prefix string, allVersions bool) map[string]string {
	m := make(map[string]string)
	c.db.View(func(txn *badger.Txn) error {
		it := txn.NewIterator(badger.IteratorOptions{
			PrefetchValues: true,
			PrefetchSize:   100,
			Reverse:        false,
			AllVersions:    allVersions,
		})
		defer it.Close()
		prefix := []byte(prefix)
		for it.Seek(prefix); it.ValidForPrefix(prefix); it.Next() {
			item := it.Item()
			k := item.Key()
			err := item.Value(func(v []byte) error {
				m[string(k)+"/"+strconv.FormatUint(item.Version(), 10)] = string(v)
				return nil
			})
			if err != nil {
				return err
			}
		}
		return nil
	})
	return m
}

// GetAllKeys returns all the keys.
func (c *Client) GetAllKeys() []string {
	s := []string{}
	_ = c.db.View(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		opts.PrefetchValues = false
		it := txn.NewIterator(opts)
		defer it.Close()
		for it.Rewind(); it.Valid(); it.Next() {
			item := it.Item()
			k := item.Key()
			s = append(s, string(k))
		}
		return nil
	})
	return s
}

// GetServiceConfigWithTags returns all config versions for some service, that completely match a set of tags
func (c *Client) GetServiceConfigWithTags(service string, tags []string, allVersions bool) map[string]string {
	configs := c.GetByPrefix("config/"+service, allVersions)
	m := make(map[string]string)
	for k, v := range configs {
		pass := true
		for _, tag := range tags {
			if !strings.Contains(k, "tag:"+tag) {
				pass = false
			}
		}
		if pass && len(tags) == strings.Count(k, "tag:") {
			m[k] = v
		}
	}
	return m
}

// SetServiceConfigWithTags stores a config for some service with tags
func (c *Client) SetServiceConfigWithTags(service string, cType string, tags []string, config []byte) {
	b := "config/" + service + "/" + cType
	for _, tag := range tags {
		b += "/tag:" + tag
	}
	c.Set(b, config)
}

type ProtoD struct {
	Name       string      `json:"name"`
	Tags       []string    `json:"tags"`
	ServerInfo *ServerInfo `json:"server_info"`
}

type ServerInfo struct {
	Version string `json:"version"`
	State   string `json:"state"`
}

func errorResponse(w http.ResponseWriter, message string, httpStatusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(httpStatusCode)
	resp := make(map[string]string)
	resp["message"] = message
	jsonResp, _ := json.Marshal(resp)
	w.Write(jsonResp)
}

func (c *Client) protodPath(w http.ResponseWriter, r *http.Request) {
	headerContentTtype := r.Header.Get("Content-Type")
	if headerContentTtype != "application/json" {
		errorResponse(w, "Content Type is not application/json", http.StatusUnsupportedMediaType)
		return
	}
	var protod ProtoD
	var unmarshalErr *json.UnmarshalTypeError

	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	err := decoder.Decode(&protod)
	if err != nil {
		if errors.As(err, &unmarshalErr) {
			errorResponse(w, "Bad Request. Wrong Type provided for field "+unmarshalErr.Field, http.StatusBadRequest)
		} else {
			errorResponse(w, "Bad Request "+err.Error(), http.StatusBadRequest)
		}
		return
	}
	configs := c.GetServiceConfigWithTags(protod.Name, protod.Tags, false)
	json.NewEncoder(w).Encode(configs)
	fmt.Println("Endpoint Hit: ProtoD")
}

type Send struct {
	Name   string   `json:"name"`
	Type   string   `json:"type"`
	Tags   []string `json:"tags"`
	Config string   `json:"config"`
}

func (c *Client) configPath(w http.ResponseWriter, r *http.Request) {
	headerContentTtype := r.Header.Get("Content-Type")
	if headerContentTtype != "application/json" {
		errorResponse(w, "Content Type is not application/json", http.StatusUnsupportedMediaType)
		return
	}
	var send Send
	var unmarshalErr *json.UnmarshalTypeError

	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	err := decoder.Decode(&send)
	if err != nil {
		if errors.As(err, &unmarshalErr) {
			errorResponse(w, "Bad Request. Wrong Type provided for field "+unmarshalErr.Field, http.StatusBadRequest)
		} else {
			errorResponse(w, "Bad Request "+err.Error(), http.StatusBadRequest)
		}
		return
	}
	c.SetServiceConfigWithTags(send.Name, send.Type, send.Tags, []byte(send.Config))
	fmt.Println("Endpoint Hit: Config")
}

func (c *Client) handleRequests() {
	myRouter := mux.NewRouter().StrictSlash(true)
	myRouter.HandleFunc("/api/protod", c.protodPath).Methods("POST")
	myRouter.HandleFunc("/api/config", c.configPath).Methods("POST")
	log.Fatal(http.ListenAndServe(":10000", myRouter))
}

func main() {
	db, _ := badger.Open(badger.DefaultOptions("").WithLoggingLevel(badger.ERROR).WithInMemory(true).WithNumVersionsToKeep(10))
	client := &Client{db: db}
	defer client.db.Close()

	cds, _ := ioutil.ReadFile("/home/dgzlopes/go/src/github.com/dgzlopes/prototype/example/configs/cds.yaml")
	lds, _ := ioutil.ReadFile("/home/dgzlopes/go/src/github.com/dgzlopes/prototype/example/configs/lds.yaml")

	client.SetServiceConfigWithTags("quote", "cds", []string{"env:production", "version:0.0.6-beta"}, cds)
	client.SetServiceConfigWithTags("quote", "lds", []string{"env:production", "version:0.0.6-beta"}, lds)
	fmt.Println(client.GetServiceConfigWithTags("quote", []string{"env:production", "version:0.0.6-beta"}, true))
	fmt.Println(client.GetAllKeys())
	client.handleRequests()
}
