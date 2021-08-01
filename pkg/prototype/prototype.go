package prototype

import (
	"context"
	"encoding/json"
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
	db, err := badger.Open(badger.DefaultOptions("").WithLoggingLevel(badger.ERROR).WithInMemory(true).WithNumVersionsToKeep(10))
	if err != nil {
		return nil, err
	}

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
		// Add default configs
		cds, _ := ioutil.ReadFile("/home/dgzlopes/go/src/github.com/dgzlopes/prototype/example/configs/cds.yaml")
		lds, _ := ioutil.ReadFile("/home/dgzlopes/go/src/github.com/dgzlopes/prototype/example/configs/lds.yaml")

		p.Set("config/default/cds", cds)
		p.Set("config/default/lds", lds)

		// Set up API
		myRouter := mux.NewRouter().StrictSlash(true)
		myRouter.HandleFunc("/api/protod", p.protodPath).Methods("POST")
		myRouter.HandleFunc("/api/config", p.configPath).Methods("POST")
		log.Fatal(http.ListenAndServe(":10000", myRouter))
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

func (p *Prototype) protodPath(w http.ResponseWriter, r *http.Request) {
	headerContentTtype := r.Header.Get("Content-Type")
	if headerContentTtype != "application/json" {
		http.Error(w, "Content-Type must be application/json", http.StatusBadRequest)
		return
	}
	var protod util.PrototypeRequest

	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	err := decoder.Decode(&protod)
	if err != nil {
		http.Error(w, "Invalid JSON"+err.Error(), http.StatusBadRequest)
		return
	}
	configs := p.GetByPrefix("config/"+protod.Service, false)
	json.NewEncoder(w).Encode(configs)
	p.logger.WithFields(log.Fields{
		"cluster": protod.Cluster,
		"service": protod.Service,
		"id":      protod.ID,
		"tags":    protod.Tags,
	}).Info("Serving configs")
}

func (p *Prototype) configPath(w http.ResponseWriter, r *http.Request) {
	headerContentTtype := r.Header.Get("Content-Type")
	if headerContentTtype != "application/json" {
		http.Error(w, "Content-Type must be application/json", http.StatusBadRequest)
		return
	}
	var req util.HTTPpayload

	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	err := decoder.Decode(&req)
	if err != nil {
		http.Error(w, "Invalid JSON"+err.Error(), http.StatusBadRequest)
		return
	}
	p.Set("config/"+req.Service+"/"+req.Type, []byte(req.Config))
	p.logger.WithFields(log.Fields{
		"cluster": req.Cluster,
		"service": req.Service,
		"type":    req.Type,
	}).Info("Applied new config")
}
