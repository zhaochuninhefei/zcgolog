package zclog

import (
	"io"
	"log"
	"os"
	"sync"
)

// const (
// 	logger_init_type_default = iota + 1
// 	logger_init_type_local
// 	logger_init_type_server
// )

// var loggerHasInitialized bool = false
var loggerLock sync.Mutex

// 初始化zcgoLogger
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
	// // 根据initType做不同场景的logger初始化
	// switch initType {
	// case logger_init_type_local:
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
	// case logger_init_type_server:
	// 	if err != nil {
	// 		// 服务器模式下，日志文件必须存在
	// 		log.Panic(err)
	// 	}
	// 	currentLogYMD = todayYMD
	// 	currentLogFile, err = os.OpenFile(logFilePath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, os.ModePerm)
	// 	if err != nil {
	// 		// 服务器模式下，日志文件必须成功打开
	// 		log.Panic(err)
	// 	}
	// 	if !zcgologConfig.LogForbidStdout {
	// 		// 日志同时输出到日志文件与控制台
	// 		multiWriter := io.MultiWriter(os.Stdout, currentLogFile)
	// 		zcgoLogger.SetOutput(multiWriter)
	// 	} else {
	// 		// 日志只输出到日志文件
	// 		zcgoLogger.SetOutput(currentLogFile)
	// 	}
	// }
}
