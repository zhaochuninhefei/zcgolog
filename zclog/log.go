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
zclog/log.go zcgolog基本处理，包括配置相关定义与处理、本地模式相关处理、日志输出处理等
*/

import (
	"fmt"
	"log"
	"os"
	"runtime"
	"time"
)

// 日志模式定义
//goland:noinspection GoSnakeCaseUsage
const (
	// LOG_MODE_LOCAL 本地模式: 日志同步输出且不支持在线修改指定logger的日志级别，日志文件不支持自动滚动，通常仅用于测试
	LOG_MODE_LOCAL = iota + 1
	// LOG_MODE_SERVER 服务器模式: 日志异步输出且支持在线修改指定logger的日志级别，日志文件支持自动滚动
	LOG_MODE_SERVER
	log_mode_max
)

// Config 日志配置
type Config struct {
	// 是否需要禁止输出到控制台，默认: false
	LogForbidStdout bool `json:"log_forbid_stdout" yaml:"log_forbid_stdout" mapstructure:"log_forbid_stdout"`
	// 日志文件目录，默认: 空，此时日志只输出到控制台
	LogFileDir string `json:"log_file_dir" yaml:"log_file_dir" mapstructure:"log_file_dir"`
	// 日志文件名前缀，默认: zcgolog
	LogFileNamePrefix string `json:"log_file_name_prefix" yaml:"log_file_name_prefix" mapstructure:"log_file_name_prefix"`
	// 日志文件大小上限，单位M，默认: 2
	LogFileMaxSizeM int `json:"log_file_max_size_m" yaml:"log_file_max_size_m" mapstructure:"log_file_max_size_m"`
	// 全局日志级别，默认:INFO
	LogLevelGlobal int `json:"log_level_global" yaml:"log_level_global" mapstructure:"log_level_global"`
	// 日志格式，默认: "%datetime %level %file %line %func %msg"，目前格式固定，该配置暂时没有使用
	LogLineFormat string `json:"log_line_format" yaml:"log_line_format" mapstructure:"log_line_format"`
	// 日志模式，默认采用本地模式，以便于本地测试
	LogMod int `json:"log_mod" yaml:"log_mod" mapstructure:"log_mod"`
	// 日志缓冲通道容量，默认 4096
	LogChannelCap int `json:"log_channel_cap" yaml:"log_channel_cap" mapstructure:"log_channel_cap"`
	// 日志缓冲通道填满后处理策略，默认:LOG_CHN_OVER_POLICY_DISCARD 丢弃该条日志
	LogChnOverPolicy int `json:"log_chn_over_policy" yaml:"log_chn_over_policy" mapstructure:"log_chn_over_policy"`
	// 日志级别控制监听服务的Host，默认:""
	LogLevelCtlHost string `json:"log_level_ctl_host" yaml:"log_level_ctl_host" mapstructure:"log_level_ctl_host"`
	// 日志级别控制监听服务的Port，默认:9300
	LogLevelCtlPort string `json:"log_level_ctl_port" yaml:"log_level_ctl_port" mapstructure:"log_level_ctl_port"`
}

// zcgoLogger
var zcgoLogger = log.New(os.Stdout, "", log.Ldate|log.Ltime)

// zcgolog配置
var zcgologConfig = &Config{
	LogForbidStdout:   false,
	LogFileDir:        "",
	LogFileNamePrefix: "zcgolog",
	LogFileMaxSizeM:   2,
	LogLevelGlobal:    LOG_LEVEL_INFO,
	LogLineFormat:     "%level %pushTime %file %line %callFunc %msg",
	LogChannelCap:     4096,
	LogChnOverPolicy:  LOG_CHN_OVER_POLICY_DISCARD,
	LogMod:            LOG_MODE_LOCAL,
	LogLevelCtlHost:   "",
	LogLevelCtlPort:   "9300",
}

// 当前日志文件
var currentLogFile *os.File

// 关闭当前日志文件
func closeCurrentLogFile() {
	if currentLogFile != nil {
		err := currentLogFile.Close()
		if err != nil {
			return
		}
	}
}

// 当天日期
var currentLogYMD string

// InitLogger 初始化zcgolog
func InitLogger(initConfig *Config) {
	// 从参数中获取有效配置覆盖logConfig
	if initConfig != nil {
		if initConfig.LogForbidStdout {
			zcgologConfig.LogForbidStdout = initConfig.LogForbidStdout
		}
		if initConfig.LogMod > 0 && initConfig.LogMod < log_mode_max {
			zcgologConfig.LogMod = initConfig.LogMod
		}
		if initConfig.LogLevelCtlHost != "" {
			zcgologConfig.LogLevelCtlHost = initConfig.LogLevelCtlHost
		}
		if initConfig.LogLevelCtlPort != "" {
			zcgologConfig.LogLevelCtlPort = initConfig.LogLevelCtlPort
		}
		if initConfig.LogChannelCap > 0 {
			zcgologConfig.LogChannelCap = initConfig.LogChannelCap
		}
		if initConfig.LogChnOverPolicy > 0 && initConfig.LogChnOverPolicy < log_chn_over_policy_max {
			zcgologConfig.LogChnOverPolicy = initConfig.LogChnOverPolicy
		}
		if initConfig.LogFileDir != "" {
			zcgologConfig.LogFileDir = initConfig.LogFileDir
		}
		if initConfig.LogFileMaxSizeM > 0 {
			zcgologConfig.LogFileMaxSizeM = initConfig.LogFileMaxSizeM
		}
		if initConfig.LogFileNamePrefix != "" {
			zcgologConfig.LogFileNamePrefix = initConfig.LogFileNamePrefix
		}
		if initConfig.LogLevelGlobal > 0 && initConfig.LogLevelGlobal < log_level_max {
			zcgologConfig.LogLevelGlobal = initConfig.LogLevelGlobal
		}
		if initConfig.LogLineFormat != "" {
			zcgologConfig.LogLineFormat = initConfig.LogLineFormat
		}
	}
	// 设置全局日志级别
	Level = zcgologConfig.LogLevelGlobal
	// 根据日志模式决定是否启用日志缓冲队列与在线修改日志级别功能
	switch zcgologConfig.LogMod {
	case LOG_MODE_SERVER:
		// 启动zcgolog服务器模式
		startZcgologServer()
	case LOG_MODE_LOCAL:
		// 初始化zcgoLogger
		initZcgoLogger()
	}
}

// 输出日志
func outputLog(msgLogLevel int, msg string, params ...interface{}) {
	// 获取日志接口调用方的程序计数器，文件名以及行号
	pc, file, line, _ := runtime.Caller(2)
	// 调用处函数包路径
	myFunc := runtime.FuncForPC(pc).Name()
	// Panic与Fatal直接调用log包处理
	if msgLogLevel == LOG_LEVEL_PANIC {
		msgPrefix := fmt.Sprintf("%s 代码:%s %d 函数:%s ", LogLevels[msgLogLevel], file, line, myFunc)
		// 输出panic日志并抛出panic，当前goroutine终止
		zcgoLogger.Panicf(msgPrefix+msg, params...)
	}
	if msgLogLevel == LOG_LEVEL_FATAL {
		msgPrefix := fmt.Sprintf("%s 代码:%s %d 函数:%s ", LogLevels[msgLogLevel], file, line, myFunc)
		// 输出fatal日志并终止程序
		zcgoLogger.Fatalf(msgPrefix+msg, params...)
	}
	// 获取函数对应的日志级别
	myLevel := logLevelCtl[myFunc]
	if myLevel == 0 {
		// 没有特别指定调用方函数的日志级别时，使用全局日志级别
		myLevel = Level
	}
	// 判断该日志是否需要输出
	if myLevel > msgLogLevel {
		return
	}
	// 根据日志模式判断同步还是异步输出
	switch zcgologConfig.LogMod {
	case LOG_MODE_SERVER:
		pushMsg := logMsg{
			pushTime:  time.Now(),
			logLevel:  msgLogLevel,
			callFile:  file,
			callLine:  line,
			callFunc:  myFunc,
			logMsg:    msg,
			logParams: params,
		}
		if msgReaderRunning {
			// 将日志消息推送到日志缓冲通道
			pushMsgToLogMsgChn(pushMsg)
		} else {
			// 服务器模式下日志缓冲通道监听服务已停止时，直接输出日志
			msgPrefix := fmt.Sprintf("%s 代码:%s %d 函数:%s ", LogLevels[msgLogLevel], file, line, myFunc)
			zcgoLogger.Printf(msgPrefix+msg, params...)
		}
	case LOG_MODE_LOCAL:
		// 本地日志模式下，直接输出日志
		msgPrefix := fmt.Sprintf("%s 代码:%s %d 函数:%s ", LogLevels[msgLogLevel], file, line, myFunc)
		zcgoLogger.Printf(msgPrefix+msg, params...)
	}
}
