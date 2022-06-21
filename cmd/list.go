/*
Copyright © 2022 NAME HERE codytan@qq.com

*/
package cmd

import (
	"encoding/csv"
	"fmt"
	"os"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/spf13/cobra"
)

var (
	cmdOpType        string
	cmdEndDate       string
	cmdPrefix        string
	cmdOutput        string
	cmdLimit         int
	cmdLargeFileSize int //超大文件的size，单位MB

	csvPath       string
	csvLargePath  string //超大文件单独csv
	csvWrite      *csv.Writer
	csvLargeWrite *csv.Writer

	endDate time.Time
)

// clearCmd represents the clear command
var listCmd = &cobra.Command{
	Use:   "list",
	Short: "列出符合条件object",
	Long:  `列出符合条件的object，暂支持prefix前缀过滤、end_date最后上传时间过滤、store_type存储类型过滤`,
	Run: func(cmd *cobra.Command, args []string) {
		//fmt.Println("clear called")
		initList()
		initQnConn()

		List()
	},
}

//遍历所有object，筛选符合条件存入csv
func List() {
	tmplog := logt.WithFields(logrus.Fields{"bucket": Bucket})

	delimiter := ""
	//初始列举marker为空
	marker := ""
	var total_count uint32
	var total_match int
	for {
		entries, _, nextMarker, hasNext, err := bucketManager.ListFiles(Bucket, Prefix, delimiter, marker, Limit)
		if err != nil {
			tmplog.Errorf("request qiniu list api error: %s, will try ", err)
			continue
		}

		for _, entry := range entries {
			total_count++
			put_time_unix := time.Unix(entry.PutTime*100/1e9, 0)

			if put_time_unix.Unix() < endDate.Unix() && entry.Type == cmdStoreType {
				total_match++

				mb_size := GetMbFileSize(entry.Fsize)
				line := []string{entry.Key, put_time_unix.Format("20060102"), fmt.Sprintf("%d", entry.Type), fmt.Sprintf("%d", mb_size)}

				if mb_size > cmdLargeFileSize {
					if err1 := csvLargeWrite.Write(line); err1 != nil {
						logt.Fatal("csv line write fail", err1)
					}
				} else {
					if err2 := csvWrite.Write(line); err2 != nil {
						logt.Fatal("csv-large line write fail", err2)
					}
				}

			}
		}

		csvWrite.Flush()

		tmplog.Info("process, already scan: ", total_count, ", match num: ", total_match)

		if total_match >= cmdLimit {
			tmplog.Info("already limit, limit: ", cmdLimit)
			break
		}

		if hasNext {
			marker = nextMarker
		} else {
			//list end
			break
		}
	}

	tmplog.Infof("finish, total_key:%d, total_match:%d, csv_file: %s", total_count, total_match, csvPath)
}

func init() {
	rootCmd.AddCommand(listCmd)

	listCmd.Flags().StringVar(&Bucket, "bucket", "", "bucket")
	listCmd.Flags().StringVar(&cmdPrefix, "prefix", "", "前缀过滤，默认无")
	listCmd.Flags().IntVar(&cmdStoreType, "store_type", 0, "存储类型过滤，0普通存储，1低频存储，2归档存储，3深度归档存储，默认0")
	listCmd.Flags().IntVar(&cmdLimit, "limit", 0, "匹配记录数，如输入大于0整数后，达到匹配数量后会停止，注由于是批次请求（每次1000），会以批次为单位使用limit，并非完全等于limit记录数")
	listCmd.Flags().IntVar(&cmdLargeFileSize, "large", 500, "超大文件的size设置，单位MB，默认500，用于区分超大文件单独放入large_xxx的csv中用于单独处理")
	listCmd.Flags().StringVar(&cmdEndDate, "end_date", "", "最后上传时间过滤，得到在此之前的object，格式如：20220101")
	listCmd.Flags().StringVarP(&cmdOutput, "out", "o", "", "输出csv文件名称")

	listCmd.MarkFlagRequired("bucket")
	listCmd.MarkFlagRequired("end_date")
}

func initList() {

	if cmdEndDate == "" {
		endDate = time.Now()
	} else {
		tmp_time, err := time.Parse("20060102", cmdEndDate)
		endDate = tmp_time
		if err != nil {
			logt.Fatal("time parse error,", err, tmp_time)
		}
	}

	if cmdOutput != "" {
		csvPath = cmdOutput
	} else {
		file_name := endDate.Format("20060102") + "_" + fmt.Sprintf("%d", cmdStoreType)
		if cmdPrefix != "" {
			file_name += "_" + cmdPrefix
		}
		csvPath = file_name + ".csv"
	}
	csvLargePath := "large_" + csvPath

	os.Remove(csvPath)
	os.Remove(csvLargePath)

	file, err := os.OpenFile(csvPath, os.O_RDWR|os.O_APPEND|os.O_CREATE, 0777)
	if err != nil {
		logt.Fatal("csv open fail")
	}
	csvWrite = csv.NewWriter(file)
	//defer file.Close()

	file_large, err1 := os.OpenFile(csvLargePath, os.O_RDWR|os.O_APPEND|os.O_CREATE, 0777)
	if err1 != nil {
		logt.Fatal("csv open fail")
	}
	csvLargeWrite = csv.NewWriter(file_large)
}
