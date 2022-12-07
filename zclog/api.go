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
zclog/api.go 提供zcgolog日志输出接口函数
*/

import "fmt"

// Print 日志级别: DEBUG
func Print(v ...interface{}) {
	msgLogLevel := LOG_LEVEL_DEBUG
	outputLog(msgLogLevel, fmt.Sprint(v...))
}

// Printf 日志级别: DEBUG
func Printf(msg string, params ...interface{}) {
	msgLogLevel := LOG_LEVEL_DEBUG
	outputLog(msgLogLevel, msg, params...)
}

// Println 日志级别: DEBUG
func Println(v ...interface{}) {
	msgLogLevel := LOG_LEVEL_DEBUG
	outputLog(msgLogLevel, fmt.Sprint(v...))
}

// Debug 输出Debug日志
func Debug(v ...interface{}) {
	msgLogLevel := LOG_LEVEL_DEBUG
	outputLog(msgLogLevel, fmt.Sprint(v...))
}

// Debugf 输出Debug日志
func Debugf(msg string, params ...interface{}) {
	msgLogLevel := LOG_LEVEL_DEBUG
	outputLog(msgLogLevel, msg, params...)
}

// Debugln 输出Debug日志
func Debugln(v ...interface{}) {
	msgLogLevel := LOG_LEVEL_DEBUG
	outputLog(msgLogLevel, fmt.Sprint(v...))
}

// Info 输出Info日志
func Info(v ...interface{}) {
	msgLogLevel := LOG_LEVEL_INFO
	outputLog(msgLogLevel, fmt.Sprint(v...))
}

// Infof 输出Info日志
func Infof(msg string, params ...interface{}) {
	msgLogLevel := LOG_LEVEL_INFO
	outputLog(msgLogLevel, msg, params...)
}

// Infoln 输出Info日志
func Infoln(v ...interface{}) {
	msgLogLevel := LOG_LEVEL_INFO
	outputLog(msgLogLevel, fmt.Sprint(v...))
}

// Warn 输出Warn日志
func Warn(v ...interface{}) {
	msgLogLevel := LOG_LEVEL_WARNING
	outputLog(msgLogLevel, fmt.Sprint(v...))
}

// Warnf 输出Warn日志
func Warnf(msg string, params ...interface{}) {
	msgLogLevel := LOG_LEVEL_WARNING
	outputLog(msgLogLevel, msg, params...)
}

// Warnln 输出Warn日志
func Warnln(v ...interface{}) {
	msgLogLevel := LOG_LEVEL_WARNING
	outputLog(msgLogLevel, fmt.Sprint(v...))
}

// Error 输出Error日志
func Error(v ...interface{}) {
	msgLogLevel := LOG_LEVEL_ERROR
	outputLog(msgLogLevel, fmt.Sprint(v...))
}

// Errorf 输出Error日志
func Errorf(msg string, params ...interface{}) {
	msgLogLevel := LOG_LEVEL_ERROR
	outputLog(msgLogLevel, msg, params...)
}

// Errorln 输出Error日志
func Errorln(v ...interface{}) {
	msgLogLevel := LOG_LEVEL_ERROR
	outputLog(msgLogLevel, fmt.Sprint(v...))
}

// Panic 直接输出日志，终止当前goroutine
//goland:noinspection GoUnusedExportedFunction
func Panic(v ...interface{}) {
	msgLogLevel := LOG_LEVEL_PANIC
	outputLog(msgLogLevel, fmt.Sprint(v...))
}

// Panicf 直接输出日志，终止当前goroutine
//goland:noinspection GoUnusedExportedFunction
func Panicf(msg string, params ...interface{}) {
	msgLogLevel := LOG_LEVEL_PANIC
	outputLog(msgLogLevel, msg, params...)
}

// Panicln 直接输出日志，终止当前goroutine
//goland:noinspection GoUnusedExportedFunction
func Panicln(v ...interface{}) {
	msgLogLevel := LOG_LEVEL_PANIC
	outputLog(msgLogLevel, fmt.Sprint(v...))
}

// Fatal 直接输出日志，终止程序
//goland:noinspection GoUnusedExportedFunction
func Fatal(v ...interface{}) {
	msgLogLevel := LOG_LEVEL_FATAL
	outputLog(msgLogLevel, fmt.Sprint(v...))
}

// Fatalf 直接输出日志，终止程序
//goland:noinspection GoUnusedExportedFunction
func Fatalf(msg string, params ...interface{}) {
	msgLogLevel := LOG_LEVEL_FATAL
	outputLog(msgLogLevel, msg, params...)
}

// Fatalln 直接输出日志，终止程序
//goland:noinspection GoUnusedExportedFunction
func Fatalln(v ...interface{}) {
	msgLogLevel := LOG_LEVEL_FATAL
	outputLog(msgLogLevel, fmt.Sprint(v...))
}
