/*
Copyright © 2022 NAME HERE codytan@qq.com

*/
package cmd

import (
	"encoding/csv"
	"fmt"
	"os"
	"sync/atomic"

	"github.com/sirupsen/logrus"

	"github.com/qiniu/go-sdk/v7/sms/rpc"
	"github.com/qiniu/go-sdk/v7/storage"
	"github.com/spf13/cobra"
)

var (
	totalNum, totalSuccNum uint32
)

const (
	CLEAR_LIST          string = "list"
	CLEAR_DELETE        string = "delete"
	CLEAR_CHANGE_TYPE_0 string = "change0" //变为标准存储
	CLEAR_CHANGE_TYPE_1 string = "change1" //变为低频存储
	CLEAR_CHANGE_TYPE_2 string = "change2" //归档存储
	CLEAR_CHANGE_TYPE_3 string = "change3" //深度归档存储
)

type ClearFile struct {
	Key     string
	PutTime int64
}

// clearCmd represents the clear command
var BatchCmd = &cobra.Command{
	Use:   "batch",
	Short: "批量操作object",
	Long:  `批量操作object，可转深度归档存储，可删除，节省费用`,
	Run: func(cmd *cobra.Command, args []string) {
		batchChan = make(chan []string)
		initQnConn()

		wg.Add(cmdWorker)
		for i := 0; i < cmdWorker; i++ {
			go batchOpWorker(fmt.Sprintf("worker-%d", i))
		}

		ImportData(cmdCsv)
		close(batchChan)
		wg.Wait()
		logt.Infof("total_num: %d, succ_num: %d", totalNum, totalSuccNum)
	},
}

func ImportData(csvPath string) error {
	file, err := os.Open(csvPath)
	if err != nil {
		logt.Fatal(err)
	}
	defer file.Close()

	read_csv := csv.NewReader(file)

	i := 0
	var batch_slice []string
	for {
		i++
		line, err := read_csv.Read()

		if err != nil {
			if err.Error() == "EOF" {
				logt.Infof("csv read all, num: %v", i)
				pushChan(batch_slice)
				break
			} else {
				logt.Error(err)
			}
		}

		batch_slice = append(batch_slice, line[0])

		if i%1000 == 0 {
			pushChan(batch_slice)
			//fmt.Println(batch_slice)
			batch_slice = nil
		}
	}
	return nil
}

func pushChan(list []string) {

	keyOps := make([]string, 0, len(list))

	for _, item := range list {

		switch cmdOpType {
		case "change0":
			keyOps = append(keyOps, storage.URIChangeType(Bucket, item, 0))
		case "change1":
			keyOps = append(keyOps, storage.URIChangeType(Bucket, item, 1))
		case "change2":
			keyOps = append(keyOps, storage.URIChangeType(Bucket, item, 2))
		case "change3":
			keyOps = append(keyOps, storage.URIChangeType(Bucket, item, 3))
		case "delete":
			keyOps = append(keyOps, storage.URIDelete(Bucket, item))
		}
	}

	batchChan <- keyOps
}

//批量操作worker
func batchOpWorker(workerName string) (err error) {
	tmplog := logt.WithFields(logrus.Fields{"worker": workerName, "bucket": Bucket})
	tmplog.Info("start worker")

	defer wg.Done()

	for {
		batchOps, ok := <-batchChan
		if !ok {
			tmplog.Info("worker stop")
			return
		}

		var succ_num uint32
		var try_num int
		rets, err := bucketManager.Batch(batchOps)
		// fmt.Printf("%+v", rets)
		if err != nil {
			// 遇到错误
			if _, ok := err.(*rpc.ErrorInfo); ok {
				for _, ret := range rets {
					// 200 为成功
					fmt.Printf("%d\n", ret.Code)
					if ret.Code != 200 {
						tmplog.Errorf("batch error1, try %d, err: %s,", try_num, ret.Data.Error)
					}
				}
			} else {
				tmplog.Errorf("batch error1, try %d, err: %s,", try_num, err)
			}
			try_num++
			if try_num <= cmdTryNum {
				//继续重试
				continue
			}

		} else {
			//重置重试次数
			try_num = 0
			// 完全成功
			for _, ret := range rets {
				// 200 为成功
				//fmt.Printf("%+v", ret)
				if ret.Code != 200 {
					fmt.Println("ret_code:", ret.Code, "error:", ret.Data.Error, "hash:", ret.Data.Hash, "error:", ret.Data.Error)
				} else {
					succ_num++
				}
			}
		}

		total_num := uint32(len(batchOps))
		atomic.AddUint32(&totalNum, total_num)
		atomic.AddUint32(&totalSuccNum, succ_num)
		tmplog.Infof("batch op finish, total_num:%d, succ num:%d", total_num, succ_num)
		tmplog.Infof("total_process, succ: %d, total: %d \n", totalSuccNum, totalNum)
	}
	return
}

func init() {
	rootCmd.AddCommand(BatchCmd)

	BatchCmd.Flags().StringVar(&Bucket, "bucket", "", "bucket name")
	BatchCmd.Flags().IntVar(&cmdWorker, "worker", 20, "并行处理的协程数量，根据机器和网络及七牛服务处理能力决定，默认20，实测50性能较佳")
	BatchCmd.Flags().IntVar(&cmdTryNum, "try", 10, "请求重试次数，默认10，应对网络或服务端接口不稳定情况")
	BatchCmd.Flags().StringVar(&cmdOpType, "type", "", "批量操作类型，change0为转普通存储，change1为转低频存储，change2为转归档存储，change3为转深度归档存储，delete为删除")
	BatchCmd.Flags().StringVar(&cmdCsv, "csv", "", "需要处理的csv文件路径")

	BatchCmd.MarkFlagRequired("bucket")
	BatchCmd.MarkFlagRequired("type")
	BatchCmd.MarkFlagRequired("csv")
}
