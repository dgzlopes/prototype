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
	Cluster string `json:"cluster"`
	Service string `json:"service"`
	Type    string `json:"type"`
	Config  string `json:"config"`
}

// applyCmd represents the apply command
var applyCmd = &cobra.Command{
	Use:   "apply -c CLUSTER -s SERVICE -t TYPE -f FILENAME",
	Short: "A brief description of your command",
	Long:  `A longer description that spans multiple lines`,
	Run: func(cmd *cobra.Command, args []string) {
		filePath := cmd.PersistentFlags().Lookup("filename").Value.String()
		if !fileExists(filePath) {
			fmt.Println("File doesn't exist: " + filePath)
			os.Exit(1)
		}
		file, _ := ioutil.ReadFile(filePath)
		json_data, err := json.Marshal(HTTPpayload{
			Cluster: cmd.PersistentFlags().Lookup("cluster").Value.String(),
			Service: cmd.PersistentFlags().Lookup("service").Value.String(),
			Type:    cmd.PersistentFlags().Lookup("type").Value.String(),
			Config:  string(file),
		})
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		_, err = http.Post(rootCmd.PersistentFlags().Lookup("endpoint").Value.String(), "application/json", bytes.NewBuffer(json_data))
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		fmt.Println("Profit!")
	},
}

func init() {
	rootCmd.AddCommand(applyCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	applyCmd.PersistentFlags().StringP("cluster", "c", "", "Cluster")
	applyCmd.PersistentFlags().StringP("service", "s", "", "Service")
	applyCmd.PersistentFlags().StringP("type", "t", "", "Type (lds,cds)")
	applyCmd.PersistentFlags().StringP("filename", "f", "", "Filename")
	applyCmd.MarkPersistentFlagRequired("cluster")
	applyCmd.MarkPersistentFlagRequired("service")
	applyCmd.MarkPersistentFlagRequired("type")
	applyCmd.MarkPersistentFlagRequired("filename")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	//applyCmd.Flags().String("prototype-endpoint", "http://localhost:10000/api/config", "Endpoint")
}

func fileExists(filename string) bool {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}
