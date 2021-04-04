package cmd

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"
)

// Manifest
type Manifest struct {
	Name      string            `yaml:"name"`
	Tags      map[string]string `yaml:"tags"`
	Resources []Resources       `yaml:"resources"`
}

// Resources
type Resources struct {
	Type string `yaml:"type"`
	Path string `yaml:"path"`
}

type HTTPpayload struct {
	Name   string   `json:"name"`
	Type   string   `json:"type"`
	Tags   []string `json:"tags"`
	Config string   `json:"config"`
}

// applyCmd represents the apply command
var applyCmd = &cobra.Command{
	Use:   "apply",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Args: cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		var manifest Manifest
		yamlFile, err := ioutil.ReadFile(args[0])
		if err != nil {
			log.Fatal(err)
		}
		err = yaml.Unmarshal(yamlFile, &manifest)
		if err != nil {
			log.Fatal(err)
		}
		var tags []string
		for tagName, tagVal := range manifest.Tags {
			tags = append(tags, tagName+":"+tagVal)
		}
		for _, resource := range manifest.Resources {
			err = createResource(manifest.Name, tags, resource)
			if err != nil {
				log.Fatal(err)
			}
		}
		fmt.Print("Resources applied")
	},
}

func init() {
	rootCmd.AddCommand(applyCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// applyCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// applyCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

func createResource(name string, tags []string, resource Resources) error {
	if !fileExists(resource.Path) {
		return errors.New("File doesn't exist: " + resource.Path)
	}
	file, _ := ioutil.ReadFile(resource.Path)
	json_data, err := json.Marshal(HTTPpayload{
		Name:   name,
		Type:   resource.Type,
		Tags:   tags,
		Config: string(file),
	})
	if err != nil {
		return err
	}

	_, err = http.Post("http://localhost:10000/api/config", "application/json", bytes.NewBuffer(json_data))
	if err != nil {
		return err
	}
	return nil
}

func fileExists(filename string) bool {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}
