package cmd

import (
	"fmt"
	"testing"
)

func TestDownLagerFile(t *testing.T) {

	initConfig()
	initQnConn()

	down_url := "https://mirror.bjtu.edu.cn/ubuntu-releases/22.04/ubuntu-22.04-desktop-amd64.iso"
	fmt.Println(down_url)

	err := DownFile("./aaa", down_url)
	if err != nil {
		t.Fatal(err)
	}
}
