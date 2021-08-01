package prototype

import (
	"context"
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"strconv"

	log "github.com/sirupsen/logrus"

	"github.com/cortexproject/cortex/pkg/util/services"
	badger "github.com/dgraph-io/badger/v3"
	"github.com/dgzlopes/prototype/pkg/util"
	"github.com/gorilla/mux"
)

type Prototype struct {
	services []services.Service
	logger   *log.Logger

	db *badger.DB

	cfg *Config
}

func New(cfg *Config, logger *log.Logger) ([]services.Service, error) {
	db, _ := badger.Open(badger.DefaultOptions("").WithLoggingLevel(badger.ERROR).WithInMemory(true).WithNumVersionsToKeep(10))
	p := &Prototype{
		cfg:    cfg,
		db:     db,
		logger: logger,
	}

	// Add main worker
	p.services = append(p.services, services.NewBasicService(nil, p.run, p.shutDown))

	return p.services, nil
}

func (p *Prototype) run(ctx context.Context) error {
	go func() {
		cds, _ := ioutil.ReadFile("/home/dgzlopes/go/src/github.com/dgzlopes/prototype/example/configs/cds.yaml")
		lds, _ := ioutil.ReadFile("/home/dgzlopes/go/src/github.com/dgzlopes/prototype/example/configs/lds.yaml")

		p.SetServiceConfigWithTags("default", "quote", "cds", cds)
		p.SetServiceConfigWithTags("default", "quote", "lds", lds)
		//fmt.Println(p.GetServiceConfigWithTags("default", "quote", true))
		//fmt.Println(p.GetAllKeys())
		p.handleRequests()
	}()
	<-ctx.Done()
	return nil
}

func (p *Prototype) shutDown(_ error) error {
	p.logger.Info("Shutting down prototype")
	p.db.Close()
	return nil
}

// Set the given key with the given value.
func (p *Prototype) Set(key string, value []byte) error {
	err := p.db.Update(func(txn *badger.Txn) error {
		err := txn.Set([]byte(key), value)
		return err
	})
	return err
}

// Get returns the value for the given key.
func (p *Prototype) Get(key string) string {
	var valCopy []byte
	_ = p.db.View(func(txn *badger.Txn) error {
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
func (p *Prototype) GetByPrefix(prefix string, allVersions bool) map[string]string {
	m := make(map[string]string)
	p.db.View(func(txn *badger.Txn) error {
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
func (p *Prototype) GetAllKeys() []string {
	s := []string{}
	_ = p.db.View(func(txn *badger.Txn) error {
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
func (p *Prototype) GetServiceConfigWithTags(cluster string, service string, allVersions bool) map[string]string {
	return p.GetByPrefix("config/"+service, allVersions)
}

// SetServiceConfigWithTags stores a config for some service with tags
func (p *Prototype) SetServiceConfigWithTags(cluster string, service string, cType string, config []byte) {
	b := "config/" + service + "/" + cType
	p.Set(b, config)
}

func errorResponse(w http.ResponseWriter, message string, httpStatusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(httpStatusCode)
	resp := make(map[string]string)
	resp["message"] = message
	jsonResp, _ := json.Marshal(resp)
	w.Write(jsonResp)
}

func (p *Prototype) protodPath(w http.ResponseWriter, r *http.Request) {
	headerContentTtype := r.Header.Get("Content-Type")
	if headerContentTtype != "application/json" {
		errorResponse(w, "Content Type is not application/json", http.StatusUnsupportedMediaType)
		return
	}
	var protod util.PrototypeRequest
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
	configs := p.GetServiceConfigWithTags(protod.Cluster, protod.Service, false)
	json.NewEncoder(w).Encode(configs)
	p.logger.Info("Endpoint Hit: ProtoD:  ", protod.Service, protod.Tags, protod.ID, protod.EnvoyInfo)
}

func (p *Prototype) configPath(w http.ResponseWriter, r *http.Request) {
	headerContentTtype := r.Header.Get("Content-Type")
	if headerContentTtype != "application/json" {
		errorResponse(w, "Content Type is not application/json", http.StatusUnsupportedMediaType)
		return
	}
	var send util.HTTPpayload
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
	p.SetServiceConfigWithTags(send.Cluster, send.Service, send.Type, []byte(send.Config))
	p.logger.Info("Endpoint Hit: Config")
	p.logger.Info(send)
}

func (c *Prototype) handleRequests() {
	myRouter := mux.NewRouter().StrictSlash(true)
	myRouter.HandleFunc("/api/protod", c.protodPath).Methods("POST")
	myRouter.HandleFunc("/api/config", c.configPath).Methods("POST")
	log.Fatal(http.ListenAndServe(":10000", myRouter))
}
