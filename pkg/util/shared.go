package util

import "io/ioutil"

type PrototypeRequest struct {
	Name string   `json:"name"`
	ID   string   `json:"id"`
	Tags []string `json:"tags"`
}

// WriteToDisk writes some data to disk
func WriteToDisk(fileName string, data []byte) error {
	err := ioutil.WriteFile(fileName, data, 0644)
	if err != nil {
		return err
	}
	return nil
}
