/*
Copyright © 2022 NAME HERE codytan@qq.com

*/
package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var version string = "0.2.0"

var VersionCmd = &cobra.Command{
	Use:   "version",
	Short: "version",
	Long:  `version`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("version:", version)
	},
}

func init() {
	rootCmd.AddCommand(VersionCmd)
}
