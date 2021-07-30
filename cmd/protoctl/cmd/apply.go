package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"

	"github.com/spf13/cobra"
)

type HTTPpayload struct {
	Cluster string   `json:"cluster"`
	Service string   `json:"service"`
	Type    string   `json:"type"`
	Tags    []string `json:"tags"`
	Config  string   `json:"config"`
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
		if !fileExists("./example/configs/cds.yaml") {
			fmt.Println("nein")
			return
		}
		file, _ := ioutil.ReadFile("./example/configs/lds.yaml")
		json_data, err := json.Marshal(HTTPpayload{
			Cluster: "default",
			Service: "quote",
			Type:    "lds",
			Tags:    []string{"env:production", "version:0.0.6-beta"},
			Config:  string(file),
		})
		if err != nil {
			fmt.Println(err)
		}

		_, err = http.Post("http://localhost:10000/api/config", "application/json", bytes.NewBuffer(json_data))
		if err != nil {
			fmt.Println(err)
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

func fileExists(filename string) bool {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}
