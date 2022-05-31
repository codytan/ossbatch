/*
Copyright © 2022 NAME HERE codytan@qq.com

*/
package cmd

import (
	"encoding/csv"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/qiniu/go-sdk/v7/storage"
	"github.com/spf13/cobra"
)

var (
	totalFileDownNum       uint32 //总成功下载数量
	totalFileDownFailNum   uint32 //总下载失败数量
	totalFileDownRepeatNum uint32 //总重复数量
)

// clearCmd represents the clear command
var PullCmd = &cobra.Command{
	Use:   "pull",
	Short: "批量拉取object到本地",
	Long:  `批量拉取object到本地，支持多协程`,
	Run: func(cmd *cobra.Command, args []string) {
		batchPullChan = make(chan string)
		initQnConn()

		wg.Add(cmdWorker)
		for i := 0; i < cmdWorker; i++ {
			go DownWorker(fmt.Sprintf("worker-%d", i))
		}

		ImportDataForPull(cmdCsv)
		close(batchPullChan)
		wg.Wait()
	},
}

//下载的协程
func DownWorker(workerName string) error {
	logt.Info("start worker " + workerName)

	defer wg.Done()

	for {
		key, ok := <-batchPullChan
		if !ok {
			logt.Infof("worker %s stop", workerName)
			return nil
		}

		deadline := time.Now().Add(time.Second * 7200).Unix() //2小时有效期
		privateAccessURL := storage.MakePrivateURL(mac, cmdDomain, key, deadline)

		pathSlice := strings.Split(privateAccessURL, "/")
		//加上当前目录做保护
		dst := "./" + strings.Join(pathSlice[3:], string(os.PathSeparator))
		//fmt.Println(privateAccessURL, dst)
		DownFile(dst, privateAccessURL)
		logt.Infof("process SuccDown: %d, FailDown: %d, Repeat: %d", totalFileDownNum, totalFileDownFailNum, totalFileDownRepeatNum)
	}
}

//导入csv中文件进入channel
func ImportDataForPull(csvPath string) error {
	file, err := os.Open(csvPath)
	if err != nil {
		logt.Fatal(err)
	}
	defer file.Close()

	read_csv := csv.NewReader(file)

	i := 0
	for {
		i++
		line, err := read_csv.Read()

		if err != nil {
			if err.Error() == "EOF" {
				logt.Infof("csv read all, num: %v", i)
				break
			} else {
				logt.Error(err)
			}
		}

		batchPullChan <- line[0]

	}
	return nil
}

func init() {
	rootCmd.AddCommand(PullCmd)

	PullCmd.Flags().StringVar(&Bucket, "bucket", "", "bucket name")
	PullCmd.Flags().IntVar(&cmdWorker, "worker", 10, "并行处理的协程数量，根据机器和网络决定，默认10")
	PullCmd.Flags().IntVar(&cmdTryNum, "try", 10, "文件下载重试次数，默认10次，在网络不好特别大文件时有效")
	PullCmd.Flags().StringVar(&cmdDomain, "domain", "", "下载域名")
	PullCmd.Flags().StringVar(&cmdCsv, "csv", "", "需要处理的csv文件路径")

	PullCmd.MarkFlagRequired("bucket")
	PullCmd.MarkFlagRequired("domain")
	PullCmd.MarkFlagRequired("csv")
}
