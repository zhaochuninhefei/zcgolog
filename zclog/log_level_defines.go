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
zclog/log_level_defines.go 日志级别定义
*/

import "strings"

// 日志级别定义
//goland:noinspection GoSnakeCaseUsage
const (
	// LOG_LEVEL_DEBUG debug 调试日志，生产环境通常关闭
	LOG_LEVEL_DEBUG = iota + 1
	// LOG_LEVEL_INFO info 重要信息日志，用于提示程序过程中的一些重要信息，慎用，避免过多的INFO日志
	LOG_LEVEL_INFO
	// LOG_LEVEL_WARNING warning 警告日志，用于警告用户可能会发生问题
	LOG_LEVEL_WARNING
	// LOG_LEVEL_ERROR error 一般错误日志，一般用于提示业务错误，程序通常不会因为这样的错误终止
	LOG_LEVEL_ERROR
	// LOG_LEVEL_PANIC panic 异常错误日志，一般用于预期外的错误，程序的当前Goroutine会终止并输出堆栈信息
	LOG_LEVEL_PANIC
	// LOG_LEVEL_FATAL fatal 致命错误日志，程序会马上终止
	LOG_LEVEL_FATAL
	// log_level_max 日志级别最大值，用于内部判断日志级别是否在合法范围内
	log_level_max
)

//goland:noinspection GoSnakeCaseUsage
const (
	LOG_LEVEL_DEBUG_STR    = "debug"
	LOG_LEVEL_INFO_STR     = "info"
	LOG_LEVEL_WARNING_STR  = "warning"
	LOG_LEVEL_WARNING_STR2 = "warn"
	LOG_LEVEL_ERROR_STR    = "error"
	LOG_LEVEL_PANIC_STR    = "panic"
	LOG_LEVEL_FATAL_STR    = "fatal"
	// LOG_LEVEL_CRITICAL_STR critical为兼容其他日志框架的日志级别，在zclog中等同于fatal
	LOG_LEVEL_CRITICAL_STR = "critical"
	LOG_LEVEL_CRITICAL     = LOG_LEVEL_FATAL
)

// LogLevels 日志级别格式化显示定义
var LogLevels = [...]string{
	LOG_LEVEL_DEBUG:   "[DEBUG]",
	LOG_LEVEL_INFO:    "[ INFO]",
	LOG_LEVEL_WARNING: "[ WARN]",
	LOG_LEVEL_ERROR:   "[ERROR]",
	LOG_LEVEL_PANIC:   "[PANIC]",
	LOG_LEVEL_FATAL:   "[FATAL]",
}

// GetLogLevelByStr 根据日志级别字符串获取对应日志级别
//  critical返回6,即与fatal相同;
//  warn与warning相同，都返回3。
func GetLogLevelByStr(levelStr string) int {
	levelStrLower := strings.ToLower(levelStr)
	switch levelStrLower {
	case LOG_LEVEL_DEBUG_STR:
		return LOG_LEVEL_DEBUG
	case LOG_LEVEL_INFO_STR:
		return LOG_LEVEL_INFO
	case LOG_LEVEL_WARNING_STR, LOG_LEVEL_WARNING_STR2:
		return LOG_LEVEL_WARNING
	case LOG_LEVEL_ERROR_STR:
		return LOG_LEVEL_ERROR
	case LOG_LEVEL_PANIC_STR:
		return LOG_LEVEL_PANIC
	case LOG_LEVEL_FATAL_STR:
		return LOG_LEVEL_FATAL
	case LOG_LEVEL_CRITICAL_STR:
		return LOG_LEVEL_CRITICAL
	default:
		return LOG_LEVEL_INFO
	}
}

// GetLogLevelStrByInt 根据日志级别int值获取对应的字符串
//  6不返回"critical"，而是返回"fatal"；
//  3不返回"warn"，而是返回"warning"。
func GetLogLevelStrByInt(levelInt int) string {
	switch levelInt {
	case LOG_LEVEL_DEBUG:
		return LOG_LEVEL_DEBUG_STR
	case LOG_LEVEL_INFO:
		return LOG_LEVEL_INFO_STR
	case LOG_LEVEL_WARNING:
		return LOG_LEVEL_WARNING_STR
	case LOG_LEVEL_ERROR:
		return LOG_LEVEL_ERROR_STR
	case LOG_LEVEL_PANIC:
		return LOG_LEVEL_PANIC_STR
	case LOG_LEVEL_FATAL:
		return LOG_LEVEL_FATAL_STR
	default:
		return LOG_LEVEL_INFO_STR
	}
}
