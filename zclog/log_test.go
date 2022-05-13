/*
   Copyright (c) 2022 zhaochun
   gitee.com/zhaochuninhefei/zcgolog is licensed under Mulan PSL v2.
   You can use this software according to the terms and conditions of the Mulan PSL v2.
   You may obtain a copy of Mulan PSL v2 at:
            http://license.coscl.org.cn/MulanPSL2
   THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND, EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT, MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
   See the Mulan PSL v2 for more details.
*/

package zclog

import (
	"fmt"
	"net/http"
	"testing"
	"time"
)

var end chan bool

func TestServerLog(t *testing.T) {
	fmt.Println("----- TestServerLog -----")
	ClearDir("testdata/serverlogs")
	end = make(chan bool, 64)
	logConfig := &Config{
		LogFileDir:     "testdata/serverlogs",
		LogMod:         LOG_MODE_SERVER,
		LogLevelGlobal: LOG_LEVEL_INFO,
	}
	InitLogger(logConfig)
	time.Sleep(time.Second)
	go writeLog()
	time.Sleep(time.Second)
	<-end
}

func writeLog() {
	// 1~15 输出INFO以上日志
	// 16~30 输出ERROR以上日志
	// 31~45 输出INFO以上日志
	// 46~60 输出DEBUG以上日志
	// 61~75 输出WARN以上日志
	// 76~100 输出INFO以上日志
	for i := 0; i < 100; i++ {
		if i == 15 {
			// 从16开始，控制全局日志级别为ERROR
			resp, err := http.Get("http://localhost:9300/zcgolog/api/level/global?level=4")
			if err != nil {
				fmt.Printf("请求zcgolog/level/ctl返回错误: %s\n", err)
			} else {
				fmt.Printf("请求zcgolog/level/ctl返回: %v\n", resp)
			}
		}
		if i == 30 {
			// 从31开始，控制全局日志级别为INFO
			resp, err := http.Get("http://localhost:9300/zcgolog/api/level/global?level=2")
			if err != nil {
				fmt.Printf("请求zcgolog/level/ctl返回错误: %s\n", err)
			} else {
				fmt.Printf("请求zcgolog/level/ctl返回: %v\n", resp)
			}
		}
		if i == 45 {
			// 从46开始，控制本函数的日志级别为DEBUG
			resp, err := http.Get("http://localhost:9300/zcgolog/api/level/ctl?logger=gitee.com/zhaochuninhefei/zcgolog/zclog.writeLog&level=1")
			if err != nil {
				fmt.Printf("请求zcgolog/level/ctl返回错误: %s\n", err)
			} else {
				fmt.Printf("请求zcgolog/level/ctl返回: %v\n", resp)
			}
		}
		if i == 60 {
			// 从61开始，控制本函数的日志级别为WARN
			resp, err := http.Get("http://localhost:9300/zcgolog/api/level/ctl?logger=gitee.com/zhaochuninhefei/zcgolog/zclog.writeLog&level=3")
			if err != nil {
				fmt.Printf("请求zcgolog/level/ctl返回错误: %s\n", err)
			} else {
				fmt.Printf("请求zcgolog/level/ctl返回: %v\n", resp)
			}
		}
		if i == 75 {
			// 从76开始，尝试控制本函数的日志级别为无效数值，此时目标函数将采用全局日志级别
			resp, err := http.Get("http://localhost:9300/zcgolog/api/level/ctl?logger=gitee.com/zhaochuninhefei/zcgolog/zclog.writeLog&level=7")
			if err != nil {
				fmt.Printf("请求zcgolog/level/ctl返回错误: %s\n", err)
			} else {
				fmt.Printf("请求zcgolog/level/ctl返回: %v\n", resp)
			}
		}
		switch (i + 1) % 15 {
		case 1:
			Print("测试写入日志", i+1)
		case 2:
			Printf("测试写入日志: %d", i+1)
		case 3:
			Println("测试写入日志", i+1)
		case 4:
			Debug("测试写入日志", i+1)
		case 5:
			Debugf("测试写入日志: %d", i+1)
		case 6:
			Debugln("测试写入日志", i+1)
		case 7:
			Info("测试写入日志", i+1)
		case 8:
			Infof("测试写入日志: %d", i+1)
		case 9:
			Infoln("测试写入日志", i+1)
		case 10:
			Warn("测试写入日志", i+1)
		case 11:
			Warnf("测试写入日志: %d", i+1)
		case 12:
			Warnln("测试写入日志", i+1)
		case 13:
			Error("测试写入日志", i+1)
		case 14:
			Errorf("测试写入日志: %d", i+1)
		case 0:
			Errorln("测试写入日志", i+1)
		}
	}
	end <- true
}

func TestServerLogScroll(t *testing.T) {
	fmt.Println("----- testServerLogScroll -----")
	end = make(chan bool, 64)
	logConfig := &Config{
		LogForbidStdout: true,
		LogFileDir:      "testdata/serverlogs",
		LogMod:          LOG_MODE_SERVER,
		LogLevelGlobal:  LOG_LEVEL_DEBUG,
		LogFileMaxSizeM: 1,
		LogChannelCap:   40960,
	}
	InitLogger(logConfig)
	time.Sleep(1 * time.Second)
	go writeLog10000()
	time.Sleep(1 * time.Second)
	<-end
	QuitMsgReader(1000)
}

func writeLog10000() {
	for i := 0; i < 10000; i++ {
		Debugf("测试写入日志writeLog10000writeLog10000writeLog10000writeLog10000writeLog10000: %d", i+1)
	}
	for {
		if len(logMsgChn) == 0 {
			break
		}
		time.Sleep(1 * time.Second)
	}
	end <- true
}

func TestLocalLogDefault(t *testing.T) {
	fmt.Println("----- TestLocalLogDefault -----")
	ClearDir("testdata/locallogs")
	for i := 0; i < 100; i++ {
		// 本地模式下，中途改变日志配置
		// 21开始日志级别调整为WARNING，info日志不输出
		if i == 20 {
			logConfig := &Config{
				LogLevelGlobal: LOG_LEVEL_WARNING,
			}
			InitLogger(logConfig)
		}
		// 41开始日志级别调整为DEBUG,info日志输出
		if i == 40 {
			logConfig := &Config{
				LogLevelGlobal: LOG_LEVEL_DEBUG,
			}
			InitLogger(logConfig)
		}
		// 51开始设置日志文件目录，日志开始同时输出在控制台和日志文件
		if i == 60 {
			logConfig := &Config{
				LogFileDir: "testdata/locallogs",
			}
			InitLogger(logConfig)
		}
		Infof("测试写入日志: %d", i+1)
	}
}

func TestLocalLog(t *testing.T) {
	fmt.Println("----- TestLocalLog -----")
	ClearDir("testdata/locallogs")
	// 在首次输出日志前设置日志目录:"testdata/locallogs"
	logConfig := &Config{
		LogFileDir:        "testdata/locallogs",
		LogFileNamePrefix: "TestLocalLog",
		LogLevelGlobal:    LOG_LEVEL_DEBUG,
	}
	InitLogger(logConfig)
	for i := 0; i < 100; i++ {
		Debugf("测试写入日志: %d", i+1)
	}
}

func TestClearLogs(t *testing.T) {
	ClearDir("testdata/locallogs")
	ClearDir("testdata/serverlogs")
}
