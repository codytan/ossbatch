/*
Copyright © 2022 NAME HERE codytan@qq.com

*/
package cmd

import (
	"fmt"
	"io"
	"os"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var logt = logrus.New()

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "ob",
	Short: "oss batch operate tool",
	Long:  `对象存储批量处理工具，暂支持七牛云`,
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	// Log as JSON instead of the default ASCII formatter.
	//logt.SetFormatter(&logrus.JSONFormatter{})

	// now_time := time.Now().Format("20060102")
	// log_file := now_time + ".log"
	// file, err := os.OpenFile(log_file, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	// if err != nil {
	// 	logt.Fatal("open log file fatal")
	// }
	//defer file.Close()
	mw := io.MultiWriter(os.Stdout)
	logt.SetOutput(mw)
	logt.SetLevel(logrus.DebugLevel)

	// Output to stdout instead of the default stderr
	// Can be any io.Writer, see below for File example
	//logt.SetOutput(os.Stdout)

	cobra.OnInitialize(initConfig)
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	// Search config in home directory with name ".td" (without extension).
	homeDir, _ := os.UserHomeDir()
	//fmt.Println(homeDir)
	viper.AddConfigPath(homeDir + "/")
	viper.SetConfigType("yaml")
	viper.SetConfigName(".ossbatch")

	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			viper.WriteConfigAs(homeDir + "/.ossbatch")
		} else {
			fmt.Println(err)
		}
	}
}
