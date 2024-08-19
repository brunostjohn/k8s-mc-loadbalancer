package cmd

import (
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var rootCmd = &cobra.Command{
	Use:   "k8s-mc-loadbalancer",
	Short: "",
	Long:  "",
	Run:   runCommand,
}

func Execute() error {
	return rootCmd.Execute()
}

func init() {
	// cobra.OnInitialize()
	viper.AutomaticEnv()

	rootCmd.AddCommand(startCmd)
}

func runCommand(cmd *cobra.Command, args []string) {
	// Do stuff
}
