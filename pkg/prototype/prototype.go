package prototype

import (
	"context"
	_ "embed"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	log "github.com/sirupsen/logrus"

	"github.com/cortexproject/cortex/pkg/util/services"
	badger "github.com/dgraph-io/badger/v3"
	"github.com/dgzlopes/prototype/pkg/util"
	"github.com/ghodss/yaml"
	"github.com/gorilla/mux"
)

//go:embed templates/index.html
var indexPage []byte

type Prototype struct {
	services []services.Service
	logger   *log.Logger

	db *Badger

	cfg *util.Config
}

type Badger struct {
	i *badger.DB
}

func New(cfg *util.Config, logger *log.Logger) ([]services.Service, error) {
	db, err := badger.Open(badger.DefaultOptions("").WithLoggingLevel(badger.ERROR).WithInMemory(true).WithNumVersionsToKeep(10))
	if err != nil {
		return nil, err
	}

	p := &Prototype{
		cfg:    cfg,
		db:     &Badger{i: db},
		logger: logger,
	}

	// Add main worker
	p.services = append(p.services, services.NewBasicService(nil, p.run, p.shutDown))

	return p.services, nil
}

func (p *Prototype) run(ctx context.Context) error {
	go func() {
		myRouter := mux.NewRouter().StrictSlash(true)
		myRouter.HandleFunc("/api/protod", p.protodPath).Methods("POST")
		myRouter.HandleFunc("/api/protod", p.protodGetAll).Methods("GET")
		myRouter.HandleFunc("/api/config", p.configPath).Methods("POST")
		myRouter.HandleFunc("/api/config", p.configGetAll).Methods("GET")
		myRouter.HandleFunc("/api/config/{cluster}/{service}/{type}", p.configGetPath).Methods("GET")
		myRouter.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
			// Use templating to ship this info on the webpage
			for _, key := range p.db.GetAllKeys() {
				if strings.Contains(key, "/config/") {
					p.logger.Info("config")
				} else {
					p.logger.Info("instance")
				}
			}
			w.WriteHeader(200)
			w.Write(indexPage)
		})
		log.Fatal(http.ListenAndServe(":10000", myRouter))
	}()
	<-ctx.Done()
	return nil
}

func (p *Prototype) shutDown(_ error) error {
	p.logger.Info("Shutting down prototype")
	p.db.i.Close()
	return nil
}

// Set the given key with the given value.
func (b *Badger) Set(key string, value []byte) error {
	err := b.i.Update(func(txn *badger.Txn) error {
		err := txn.Set([]byte(key), value)
		return err
	})
	return err
}

// Get returns the value for the given key.
func (b *Badger) Get(key string) (string, error) {
	var valCopy []byte
	err := b.i.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte(key))
		if err != nil {
			return err
		}
		err = item.Value(func(val []byte) error {
			valCopy = append([]byte{}, val...)
			return nil
		})
		return err
	})
	return string(valCopy), err
}

// GetAllKeys returns all the keys.
func (b *Badger) GetAllKeys() []string {
	s := []string{}
	_ = b.i.View(func(txn *badger.Txn) error {
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

// GetByPrefix return all the key value pairs where the key starts with some prefix.
func (b *Badger) GetByPrefix(prefix string, allVersions bool) map[string]string {
	m := make(map[string]string)
	b.i.View(func(txn *badger.Txn) error {
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
	configs := p.db.GetByPrefix(keyGen(protod.Cluster, protod.Service, "", ""), false)
	json.NewEncoder(w).Encode(configs)
	p.logger.WithFields(log.Fields{
		"cluster": protod.Cluster,
		"service": protod.Service,
		"id":      protod.ID,
		"tags":    protod.Tags,
	}).Info("Serving configs")

	// Store Protod metadata in DB
	enc, err := protod.EncodeToBytes()
	if err != nil {
		http.Error(w, "Could not encode ProtoD", http.StatusInternalServerError)
		return
	}
	p.db.Set(keyGen(protod.Cluster, protod.Service, "", protod.ID), enc)
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
	p.db.Set(keyGen(req.Cluster, req.Service, req.Type, ""), []byte(req.Config))
	p.logger.WithFields(log.Fields{
		"cluster": req.Cluster,
		"service": req.Service,
		"type":    req.Type,
	}).Info("Applied new config")
}

func (p *Prototype) protodGetAll(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var toret []util.PrototypeRequest
	v := p.db.GetAllKeys()
	for _, a := range v {
		if strings.Contains(a, "/id/") {
			v, err := p.db.Get(a)
			if err != nil {
				http.Error(w, "Could not find Protod", http.StatusNotFound)
				return
			}
			proto, err := util.DecodeFromBytes([]byte(v))
			if err != nil {
				http.Error(w, "Could not decode Protod", http.StatusInternalServerError)
				return
			}
			toret = append(toret, proto)
		}
	}
	json.NewEncoder(w).Encode(toret)
}

func (p *Prototype) configGetAll(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var toret []string
	v := p.db.GetAllKeys()
	for _, a := range v {
		if strings.Contains(a, "/config/") {
			toret = append(toret, a)
		}
	}
	json.NewEncoder(w).Encode(toret)
}

func (p *Prototype) configGetPath(w http.ResponseWriter, r *http.Request) {
	// TODO: Let the user pick a specific version? Or return all the versions?
	w.Header().Set("Content-Type", "application/json")
	vars := mux.Vars(r)
	v, err := p.db.Get(keyGen(vars["cluster"], vars["service"], vars["type"], ""))
	if err != nil {
		http.Error(w, "Could not find config", http.StatusNotFound)
		return
	}
	// From YAML string to JSON response
	j, _ := yaml.YAMLToJSON([]byte(v))
	var u map[string]interface{}
	_ = json.Unmarshal(j, &u)
	json.NewEncoder(w).Encode(u)
}

func keyGen(cluster string, service string, rType string, id string) string {
	key := cluster + "/" + service
	if id != "" {
		key += "/id/" + id
	}
	if rType != "" {
		key += "/config/" + rType
	}
	return key
}
