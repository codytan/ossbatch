/*
Copyright © 2022 NAME HERE codytan@qq.com

*/
package cmd

import (
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	ConfAk string
	ConfSk string
)

// configCmd represents the config command
var configCmd = &cobra.Command{
	Use:   "config",
	Short: "config ak and sk",
	Long:  `配置OSS访问AK/SK`,
	Run: func(cmd *cobra.Command, args []string) {
		viper.WriteConfig()
		logt.Info("config save succ")
	},
}

func init() {
	rootCmd.AddCommand(configCmd)

	configCmd.Flags().StringVar(&ConfAk, "ak", "", "输入AK")
	configCmd.Flags().StringVar(&ConfSk, "sk", "", "输入SK")

	configCmd.MarkFlagRequired("ak")
	configCmd.MarkFlagRequired("sk")

	viper.BindPFlag("ak", configCmd.Flags().Lookup("ak"))
	viper.BindPFlag("sk", configCmd.Flags().Lookup("sk"))
}
