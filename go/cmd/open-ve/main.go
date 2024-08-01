package main

import (
	"github.com/shibukazu/open-ve/go/cmd/run"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func main() {
	cobra.OnInitialize(Init)
	rootCmd := &cobra.Command{
		Use:   "open-ve",
		Short: "Open-VE: A powerful solution that simplifies the management of validation rules, ensuring consistent validation across all layers, including frontend, BFF, and microservices, through a single, simple API.",
		Long:  "Open-VE: A powerful solution that simplifies the management of validation rules, ensuring consistent validation across all layers, including frontend, BFF, and microservices, through a single, simple API.",
	}

	runCmd := run.NewRunCommand()
	rootCmd.AddCommand(runCmd)

	if err := rootCmd.Execute(); err != nil {
		panic(err)
	}

}

func Init() {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")

	viper.SetEnvPrefix("OPEN-VE")
	viper.AutomaticEnv()

	configPaths := []string{"$HOME/.open-ve", "."}
	for _, path := range configPaths {
		viper.AddConfigPath(path)
	}
}
