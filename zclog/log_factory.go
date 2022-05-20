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

/*
zclog/log_factory.go zcgolog工厂，用于初始化zcgoLogger
*/

import (
	"io"
	"log"
	"os"
	"sync"
)

var loggerLock sync.Mutex

// 初始化zcgoLogger
//  设置zcgoLogger的输出目标与日志前缀格式
func initZcgoLogger() {
	// 停止日志缓冲通道监听
	// 防止应用程序在已经开启服务器模式后，刷新logger配置时与日志缓冲通道监听处理(readAndWriteMsg)中对日志文件的处理发生冲突。
	err := QuitMsgReader(30000)
	if err != nil {
		log.Panic(err)
	}
	// 上锁,确保logger操作的线程安全
	loggerLock.Lock()
	defer loggerLock.Unlock()
	// 临时切换zcgoLogger输出到控制台
	zcgoLogger.SetOutput(os.Stdout)
	// 设置日志前缀格式
	zcgoLogger.SetFlags(log.Ldate | log.Ltime)
	// 关闭当前日志文件
	closeCurrentLogFile()
	// 获取最新日志文件
	logFilePath, todayYMD, err := GetLogFilePathAndYMDToday(zcgologConfig)
	if err != nil {
		// 未能成功获取日志文件时，直接输出到控制台
		currentLogYMD = getYMDToday()
		currentLogFile = nil
		zcgoLogger.Println(err.Error())
		return
	}
	if logFilePath == OS_OUT_STDOUT {
		// LogFileDir为空时，直接输出到控制台
		currentLogYMD = todayYMD
		currentLogFile = nil
		return
	}
	currentLogYMD = todayYMD
	currentLogFile, err = os.OpenFile(logFilePath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, os.ModePerm)
	if err != nil {
		// 未能成功打开日志文件时，直接输出到控制台
		currentLogFile = nil
		zcgoLogger.Println(err.Error())
		return
	}
	if !zcgologConfig.LogForbidStdout {
		// 日志同时输出到日志文件与控制台
		multiWriter := io.MultiWriter(os.Stdout, currentLogFile)
		zcgoLogger.SetOutput(multiWriter)
	} else {
		// 日志只输出到日志文件
		zcgoLogger.SetOutput(currentLogFile)
	}
}
