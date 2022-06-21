/*
Copyright © 2022 NAME HERE codytan@qq.com

*/
package cmd

import (
	"errors"
	"fmt"
	"net/url"
	"os"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/cavaliergopher/grab/v3"
	"github.com/spf13/viper"

	"github.com/qiniu/go-sdk/v7/auth"
	"github.com/qiniu/go-sdk/v7/storage"
)

var (
	Bucket string
	Limit  int = 1000
	Prefix string

	cmdWorker    int
	cmdTryNum    int //下载重试次数
	cmdCsv       string
	cmdDomain    string
	cmdVerbose   int //是否打印详细信息
	cmdStoreType int

	wg            sync.WaitGroup
	mac           *auth.Credentials
	bucketManager *storage.BucketManager
	batchChan     chan []string
	batchPullChan chan string //下载的channel
)

//初始化七牛sdk
func initQnConn() {
	mac = auth.New(viper.GetString("ak"), viper.GetString("sk"))
	cfg := storage.Config{
		// 是否使用https域名进行资源管理
		UseHTTPS: false,
	}
	bucketManager = storage.NewBucketManager(mac, &cfg)
}

//下载一个文件
//dst 为目标文件，如不存在会自动创建，可以包含最终文件名
//url 下载地址
func DownFile(dst string, down_url string) error {
	if down_url == "" {
		logt.Warn("url empty")
		atomic.AddUint32(&totalFileDownFailNum, 1)
		return errors.New("url empty")
	}

	//有URL参数需去掉
	if index := strings.Index(dst, "?"); index != -1 {
		dst = string([]byte(dst)[:index])
	}

	dst, err := url.QueryUnescape(dst)
	if err != nil {
		logt.Errorf("escape fail %s, err: %s", down_url, err)
		atomic.AddUint32(&totalFileDownFailNum, 1)
		return errors.New("url escape fail")
	}

	//重复文件跳过
	if _, ok := FileIsExist(dst); ok {
		logt.Debug("file exist: ", dst)
		atomic.AddUint32(&totalFileDownRepeatNum, 1)
		return nil
	}

	try_count := 0
	for {
		if err := GrabDown(dst, down_url); err != nil {
			logt.Errorf("down fail, will try %d, url:%s, err:%s", try_count, down_url, err)
			//超过重试次数，参数错误，都停止
			if try_count >= cmdTryNum {
				atomic.AddUint32(&totalFileDownFailNum, 1)
				return err
			}

			try_count++
			time.Sleep(time.Duration(200) * time.Millisecond)
			continue

		} else {
			logt.Debugf("saved to %s", dst)
			atomic.AddUint32(&totalFileDownNum, 1)
			break
		}
	}

	//atomic.AddUint32(&totalFileDownFailNum, 1)
	return nil
}

//利用grab加入重试机制后的下载方法
func GrabDown(dst string, down_url string) error {
	try_count := 0
TRY_DOWN:
	try_count++
	client := grab.NewClient()
	req, _ := grab.NewRequest(dst, down_url)
	// req.SkipExisting = true
	resp := client.Do(req)

	//verbose下打印进度
	if cmdVerbose > 0 {
		// start UI loop
		t := time.NewTicker(1000 * time.Millisecond)
		defer t.Stop()

	Loop:
		for {
			select {
			case <-t.C:
				fmt.Printf("file:%s, size %d MB, (%.2f%%)\n",
					resp.Filename,
					GetMbFileSize(resp.Size()),
					100*resp.Progress())

			case <-resp.Done:
				// download is complete
				break Loop
			}
		}
	}

	if err := resp.Err(); err != nil {
		//由于网络中断不稳定导致的EOF重试
		if err.Error() == "unexpected EOF" {
			logt.Warnf("down warn, will try %d, dst: %s, err:%s", try_count, dst, err)
			time.Sleep(time.Duration(100) * time.Millisecond)
			goto TRY_DOWN
		}
		return err
	}

	return nil
}

//判断文件是否存在，存在返回true
func FileIsExist(path string) (os.FileInfo, bool) {
	f, err := os.Stat(path)
	return f, err == nil || os.IsExist(err)
}

func GetMbFileSize(b int64) int {
	return int(b / 1024 / 1024)
}

//debug file down func
// func FileDown(dst string, down_url string) error {

// 	client := http.Client{}
// 	//client.Timeout = time.Second * 60

// 	req, _ := http.NewRequest(http.MethodGet, down_url, nil)
// 	//req.Header.Add("Accept-Encoding", "gzip")

// 	//fmt.Printf("req: %+v \n\n", req)

// 	resp, err := client.Do(req)
// 	if err != nil {
// 		panic(err)
// 	}

// 	//fmt.Printf("resp: %+v \n", resp)

// 	raw := resp.Body
// 	defer raw.Close()

// 	if resp.StatusCode != 200 {
// 		return errors.New(fmt.Sprintf("http status code: %d", resp.StatusCode))
// 	}

// 	file, err := os.Create(dst)
// 	if err != nil {
// 		return err
// 	}
// 	defer file.Close()

// 	_, err = io.Copy(file, raw)
// 	if err != nil {
// 		return err
// 	}

// 	return nil
// }
