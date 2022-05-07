/*
   Copyright (c) 2022 zhaochun
   gitee.com/zhaochuninhefei/zcgolog is licensed under Mulan PSL v2.
   You can use this software according to the terms and conditions of the Mulan PSL v2.
   You may obtain a copy of Mulan PSL v2 at:
            http://license.coscl.org.cn/MulanPSL2
   THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND, EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT, MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
   See the Mulan PSL v2 for more details.
*/

package log

import (
	"fmt"
	"net/http"
	"testing"
	"time"
)

var end chan bool

func TestServerLog(t *testing.T) {
	end = make(chan bool, 64)
	logConfig := &Config{
		LogFileDir:     "testdata/serverlogs",
		LogMod:         LOG_MODE_SERVER,
		LogLevelGlobal: LOG_LEVEL_INFO,
	}
	InitLogger(logConfig)
	time.Sleep(3 * time.Second)
	go writeLog()
	time.Sleep(3 * time.Second)
	<-end
}

func writeLog() {
	for i := 0; i < 100; i++ {
		if i == 20 {
			// 从20开始，控制本函数的日志级别为DEBUG
			resp, err := http.Get("http://localhost:9300/zcgolog/api/level/ctl?logger=gitee.com/zhaochuninhefei/zcgolog/log.writeLog&level=1")
			if err != nil {
				fmt.Printf("请求zcgolog/level/ctl返回错误: %s\n", err)
			} else {
				fmt.Printf("请求zcgolog/level/ctl返回: %v\n", resp)
			}
		}
		if i == 70 {
			// 从70开始，控制本函数的日志级别为INFO
			resp, err := http.Get("http://localhost:9300/zcgolog/api/level/ctl?logger=gitee.com/zhaochuninhefei/zcgolog/log.writeLog&level=2")
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

func TestLocalLogDefault(t *testing.T) {
	for i := 0; i < 100; i++ {
		// 本地模式下，log的初始化只会执行一次，因此中途改变日志目录并不能生效，日志文件依然在默认的"~/zcgologs"下
		if i == 50 {
			logConfig := &Config{
				LogFileDir: "testdata/locallogs",
			}
			InitLogger(logConfig)
		}
		Debug("测试写入日志: %d", i+1)
	}
}

func TestLocalLog(t *testing.T) {
	// 在首次输出日志前设置日志目录，改为"testdata/locallogs"
	// 则后续所有日志都输出到"testdata/locallogs"目录下
	logConfig := &Config{
		LogFileDir: "testdata/locallogs",
	}
	InitLogger(logConfig)
	for i := 0; i < 100; i++ {
		Debug("测试写入日志: %d", i+1)
	}
}

func TestHome(t *testing.T) {
	fmt.Println(Home())
}
