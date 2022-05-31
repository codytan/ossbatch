/*
Copyright © 2022 NAME HERE codytan@qq.com

*/
package cmd

import (
	"net/url"
	"os"
	"strings"
	"sync"
	"sync/atomic"

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
func DownFile(dst string, down_url string) (err error) {
	if down_url == "" {
		logt.Warn("url empty")
		atomic.AddUint32(&totalFileDownFailNum, 1)
		return
	}

	//有URL参数需去掉
	if index := strings.Index(dst, "?"); index != -1 {
		dst = string([]byte(dst)[:index])
	}

	dst, err = url.QueryUnescape(dst)
	if err != nil {
		logt.Error("escape fail, ", down_url, err)
		atomic.AddUint32(&totalFileDownFailNum, 1)
		return
	}

	//重复文件跳过
	if _, ok := FileIsExist(dst); ok {
		logt.Warn("file exist: ", dst)
		atomic.AddUint32(&totalFileDownRepeatNum, 1)
		return
	}

	try_num := 0
	for {
		resp, err := grab.Get(dst, down_url)
		if err != nil {
			logt.Errorf("get fail, try %d, url:%s, err:%s", try_num, down_url, err)
			if try_num > cmdTryNum {
				atomic.AddUint32(&totalFileDownFailNum, 1)
				break
			}
		} else {
			logt.Debug("saved to " + resp.Filename)
			atomic.AddUint32(&totalFileDownNum, 1)
			break
		}
		try_num++
	}
	//atomic.AddUint32(&totalFileDownFailNum, 1)
	return nil
}

//判断文件是否存在，存在返回true
func FileIsExist(path string) (os.FileInfo, bool) {
	f, err := os.Stat(path)
	return f, err == nil || os.IsExist(err)
}
