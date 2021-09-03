package cmd

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	"github.com/dgzlopes/prototype/pkg/util"
	"github.com/spf13/cobra"
)

// getCmd represents the apply command
var getCmd = &cobra.Command{
	Use:   "get config|protod",
	Short: "Get a list of configs or protods.",
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) != 1 || args[0] != "config" && args[0] != "protod" {
			fmt.Println("You have to specify config or protod.")
			fmt.Println("Example: protoctl get protod")
			os.Exit(0)
		}
		if args[0] == "config" {
			res, err := http.Get(rootCmd.PersistentFlags().Lookup("endpoint").Value.String() + "/api/config")
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
			var p []string
			err = json.NewDecoder(res.Body).Decode(&p)
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
			res.Body.Close()

			if len(p) == 0 {
				fmt.Println("No configs found.")
				os.Exit(0)
			}

			for _, v := range p {
				fmt.Println(v)
			}
		} else {
			res, err := http.Get(rootCmd.PersistentFlags().Lookup("endpoint").Value.String() + "/api/protod")
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
			var p []util.PrototypeRequest
			err = json.NewDecoder(res.Body).Decode(&p)
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
			res.Body.Close()

			if len(p) == 0 {
				fmt.Println("No protod found.")
				os.Exit(0)
			}

			for _, v := range p {
				fmt.Printf("Cluster: %s\n Service: %s\n ID: %s\n Tags: %s \n Version: %s \n State: %s \n Uptime: %s \n--\n", v.Cluster, v.Service, v.ID, v.Tags, v.EnvoyInfo.Version, v.EnvoyInfo.State, v.EnvoyInfo.Uptime)
			}
		}
	},
}

func init() {
	rootCmd.AddCommand(getCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	//applyCmd.Flags().String("prototype-endpoint", "http://localhost:10000/api/config", "Endpoint")
}
