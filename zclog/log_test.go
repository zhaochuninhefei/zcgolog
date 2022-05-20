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
	"io/ioutil"
	"net/http"
	"testing"
	"time"
)

const (
	log_ctl_server_address         = "http://localhost:19300"
	log_ctl_uri_level_query        = "/zcgolog/api/level/query"
	log_ctl_uri_level_query_logger = "/zcgolog/api/level/query?logger=gitee.com/zhaochuninhefei/zcgolog/zclog.writeLog"
)

var end chan bool

func TestServerLog(t *testing.T) {
	fmt.Println("----- TestServerLog -----")
	ClearDir("testdata/serverlogs")
	end = make(chan bool, 64)
	logConfig := &Config{
		LogFileDir:      "testdata/serverlogs",
		LogMod:          LOG_MODE_SERVER,
		LogLevelGlobal:  LOG_LEVEL_INFO,
		LogLevelCtlPort: "19300",
	}
	InitLogger(logConfig)
	time.Sleep(time.Second)
	go writeLog()
	time.Sleep(time.Second)
	<-end
}

func writeLog() {
	// 查看当前日志级别
	showLogLevelNow()
	// 1~15 输出INFO以上日志
	// 16~30 输出ERROR以上日志
	// 31~45 输出INFO以上日志
	// 46~60 输出DEBUG以上日志
	// 61~75 输出WARN以上日志
	// 76~100 输出INFO以上日志
	for i := 0; i < 100; i++ {
		if i == 15 {
			// 从16开始，控制全局日志级别为ERROR
			changeLogLevel("/zcgolog/api/level/global?level=4")
		}
		if i == 30 {
			// 从31开始，控制全局日志级别为INFO
			changeLogLevel("/zcgolog/api/level/global?level=2")
		}
		if i == 45 {
			// 从46开始，控制本函数的日志级别为DEBUG
			changeLogLevel("/zcgolog/api/level/ctl?logger=gitee.com/zhaochuninhefei/zcgolog/zclog.writeLog&level=1")
		}
		if i == 60 {
			// 从61开始，控制本函数的日志级别为WARN
			changeLogLevel("/zcgolog/api/level/ctl?logger=gitee.com/zhaochuninhefei/zcgolog/zclog.writeLog&level=3")
		}
		if i == 75 {
			// 从76开始，尝试控制本函数的日志级别为无效数值，此时目标函数将采用全局日志级别
			changeLogLevel("/zcgolog/api/level/ctl?logger=gitee.com/zhaochuninhefei/zcgolog/zclog.writeLog&level=7")
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

func changeLogLevel(uri string) {
	// uri := "/zcgolog/api/level/global?level=4"
	resp, err := http.Get(log_ctl_server_address + uri)
	if err != nil {
		fmt.Printf("请求 %s 返回错误: %s\n", uri, err)
	} else {
		response, _ := ioutil.ReadAll(resp.Body)
		fmt.Printf("请求 %s 返回: %s\n", uri, response)
		// 查看当前日志级别
		showLogLevelNow()
	}
}

func showLogLevelNow() {
	// 查看当前全局日志级别
	resp, err := http.Get(log_ctl_server_address + log_ctl_uri_level_query)
	if err != nil {
		fmt.Printf("请求 %s 返回错误: %s\n", log_ctl_uri_level_query, err)
	} else {
		response, _ := ioutil.ReadAll(resp.Body)
		fmt.Printf("===== 当前全局日志级别: %s\n", response)
	}
	// 查看指定logger的日志级别
	resp, err = http.Get(log_ctl_server_address + log_ctl_uri_level_query_logger)
	if err != nil {
		fmt.Printf("请求 %s 返回错误: %s\n", log_ctl_uri_level_query_logger, err)
	} else {
		response, _ := ioutil.ReadAll(resp.Body)
		fmt.Printf("===== 当前指定logger日志级别: %s\n", response)
	}
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
		LogLevelCtlPort: "19300",
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

func TestLogLevel(t *testing.T) {
	fmt.Println("debug :", LogLevels[LOG_LEVEL_DEBUG])
	fmt.Println("fatal :", LogLevels[LOG_LEVEL_FATAL])
	fmt.Println("critical :", LogLevels[LOG_LEVEL_CRITICAL])

	logLevel := GetLogLevelByStr("debug")
	if logLevel != LOG_LEVEL_DEBUG {
		t.Fatal("debug未能获取对应日志级别")
	}
	fmt.Println("debug获取到日志级别:", logLevel)

	logLevel = GetLogLevelByStr("warn")
	if logLevel != LOG_LEVEL_WARNING {
		t.Fatal("warn未能获取对应日志级别")
	}
	fmt.Println("warn获取到日志级别:", logLevel)

	logLevel = GetLogLevelByStr("warning")
	if logLevel != LOG_LEVEL_WARNING {
		t.Fatal("warning未能获取对应日志级别")
	}
	fmt.Println("warning获取到日志级别:", logLevel)

	logLevel = GetLogLevelByStr("fatal")
	if logLevel != LOG_LEVEL_FATAL {
		t.Fatal("fatal未能获取对应日志级别")
	}
	fmt.Println("fatal获取到日志级别:", logLevel)

	logLevel = GetLogLevelByStr("critical")
	if logLevel != LOG_LEVEL_CRITICAL {
		t.Fatal("critical未能获取对应日志级别")
	}
	fmt.Println("critical获取到日志级别:", logLevel)
}
