package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"

	"github.com/dgzlopes/prototype/pkg/util"
	"github.com/spf13/cobra"
)

// applyCmd represents the apply command
var applyCmd = &cobra.Command{
	Use:   "apply -c CLUSTER -s SERVICE -t TYPE -f FILENAME",
	Short: "Apply a set of configurations to Prototype.",
	Run: func(cmd *cobra.Command, args []string) {
		filePath := cmd.PersistentFlags().Lookup("filename").Value.String()
		if !fileExists(filePath) {
			fmt.Println("File doesn't exist: " + filePath)
			os.Exit(0)
		}
		file, _ := ioutil.ReadFile(filePath)
		json_data, err := json.Marshal(util.HTTPpayload{
			Cluster: cmd.PersistentFlags().Lookup("cluster").Value.String(),
			Service: cmd.PersistentFlags().Lookup("service").Value.String(),
			Type:    cmd.PersistentFlags().Lookup("type").Value.String(),
			Config:  string(file),
		})
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		_, err = http.Post(rootCmd.PersistentFlags().Lookup("endpoint").Value.String()+"/api/config", "application/json", bytes.NewBuffer(json_data))
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		fmt.Printf("Applied cluster:%s service:%s type:%s \n", cmd.PersistentFlags().Lookup("cluster").Value.String(), cmd.PersistentFlags().Lookup("service").Value.String(), cmd.PersistentFlags().Lookup("type").Value.String())
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
