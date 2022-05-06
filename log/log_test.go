package log

import (
	"fmt"
	"net/http"
	"testing"
	"time"
)

var end chan bool

func TestLog(t *testing.T) {
	end = make(chan bool, 64)
	logConfig := DefaultLogConfig()
	logConfig.LogFileDir = "testdata/logs"
	InitLogger(logConfig)
	time.Sleep(3 * time.Second)
	go writeLog()
	time.Sleep(3 * time.Second)
	<-end
}

func writeLog() {
	for i := 0; i < 100; i++ {
		if i == 50 {
			resp, err := http.Get("http://localhost:9300/zcgolog/level/ctl?logger=gitee.com/zhaochuninhefei/zcgolog/log.writeLog&level=1")
			if err != nil {
				fmt.Printf("请求zcgolog/level/ctl返回错误: %s\n", err)
			} else {
				fmt.Printf("请求zcgolog/level/ctl返回: %v\n", resp)
			}
		}
		Debug("测试写入日志: %d", i+1)
	}
	end <- true
}

func Test01(t *testing.T) {
	test := map[string]int{
		"a": 100,
	}
	fmt.Println(test["test"])
	fmt.Println(test["a"])
}
