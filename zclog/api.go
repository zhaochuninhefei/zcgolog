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

// 日志级别: DEBUG
func Print(v ...interface{}) {
	msgLogLevel := LOG_LEVEL_DEBUG
	outputLog(msgLogLevel, fmt.Sprint(v...))
}

// 日志级别: DEBUG
func Printf(msg string, params ...interface{}) {
	msgLogLevel := LOG_LEVEL_DEBUG
	outputLog(msgLogLevel, msg, params...)
}

// 日志级别: DEBUG
func Println(v ...interface{}) {
	msgLogLevel := LOG_LEVEL_DEBUG
	outputLog(msgLogLevel, fmt.Sprint(v...))
}

func Debug(v ...interface{}) {
	msgLogLevel := LOG_LEVEL_DEBUG
	outputLog(msgLogLevel, fmt.Sprint(v...))
}

func Debugf(msg string, params ...interface{}) {
	msgLogLevel := LOG_LEVEL_DEBUG
	outputLog(msgLogLevel, msg, params...)
}

func Debugln(v ...interface{}) {
	msgLogLevel := LOG_LEVEL_DEBUG
	outputLog(msgLogLevel, fmt.Sprint(v...))
}

func Info(v ...interface{}) {
	msgLogLevel := LOG_LEVEL_INFO
	outputLog(msgLogLevel, fmt.Sprint(v...))
}

func Infof(msg string, params ...interface{}) {
	msgLogLevel := LOG_LEVEL_INFO
	outputLog(msgLogLevel, msg, params...)
}

func Infoln(v ...interface{}) {
	msgLogLevel := LOG_LEVEL_INFO
	outputLog(msgLogLevel, fmt.Sprint(v...))
}

func Warn(v ...interface{}) {
	msgLogLevel := LOG_LEVEL_WARNING
	outputLog(msgLogLevel, fmt.Sprint(v...))
}

func Warnf(msg string, params ...interface{}) {
	msgLogLevel := LOG_LEVEL_WARNING
	outputLog(msgLogLevel, msg, params...)
}

func Warnln(v ...interface{}) {
	msgLogLevel := LOG_LEVEL_WARNING
	outputLog(msgLogLevel, fmt.Sprint(v...))
}

func Error(v ...interface{}) {
	msgLogLevel := LOG_LEVEL_ERROR
	outputLog(msgLogLevel, fmt.Sprint(v...))
}

func Errorf(msg string, params ...interface{}) {
	msgLogLevel := LOG_LEVEL_ERROR
	outputLog(msgLogLevel, msg, params...)
}

func Errorln(v ...interface{}) {
	msgLogLevel := LOG_LEVEL_ERROR
	outputLog(msgLogLevel, fmt.Sprint(v...))
}

// 直接输出日志，终止当前goroutine
func Panic(v ...interface{}) {
	msgLogLevel := LOG_LEVEL_PANIC
	outputLog(msgLogLevel, fmt.Sprint(v...))
}

// 直接输出日志，终止当前goroutine
func Panicf(msg string, params ...interface{}) {
	msgLogLevel := LOG_LEVEL_PANIC
	outputLog(msgLogLevel, msg, params...)
}

// 直接输出日志，终止当前goroutine
func Panicln(v ...interface{}) {
	msgLogLevel := LOG_LEVEL_PANIC
	outputLog(msgLogLevel, fmt.Sprint(v...))
}

// 直接输出日志，终止程序
func Fatal(v ...interface{}) {
	msgLogLevel := LOG_LEVEL_FATAL
	outputLog(msgLogLevel, fmt.Sprint(v...))
}

// 直接输出日志，终止程序
func Fatalf(msg string, params ...interface{}) {
	msgLogLevel := LOG_LEVEL_FATAL
	outputLog(msgLogLevel, msg, params...)
}

// 直接输出日志，终止程序
func Fatalln(v ...interface{}) {
	msgLogLevel := LOG_LEVEL_FATAL
	outputLog(msgLogLevel, fmt.Sprint(v...))
}
