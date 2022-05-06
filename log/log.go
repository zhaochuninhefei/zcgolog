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
	"io"
	"log"
	"net/http"
	"os"
	"runtime"
	"strconv"
	"time"
)

// 日志级别定义
const (
	LOG_LEVEL_DEBUG = iota + 1
	LOG_LEVEL_INFO
	LOG_LEVEL_WARNING
	LOG_LEVEL_ERROR
	LOG_LEVEL_CRITICAL
	LOG_LEVEL_FATAL
	log_level_max
)

// 日志级别文字定义
var LogLevels = [...]string{
	LOG_LEVEL_DEBUG:    "[   DEBUG]",
	LOG_LEVEL_INFO:     "[    INFO]",
	LOG_LEVEL_WARNING:  "[ WARNING]",
	LOG_LEVEL_ERROR:    "[   ERROR]",
	LOG_LEVEL_CRITICAL: "[CRITICAL]",
	LOG_LEVEL_FATAL:    "[   FATAL]",
}

// 日志缓冲通道填满后处理策略定义
const (
	// 丢弃该条日志
	LOG_CHN_OVER_POLICY_DISCARD = iota
	// 阻塞等待
	LOG_CHN_OVER_POLICY_BLOCK
	log_chn_over_policy_max
)

// 日志配置
type Config struct {
	// 日志文件目录，默认: ~/zcgologs
	LogFileDir string
	// 日志文件名前缀，默认: zcgolog
	LogFileNamePrefix string
	// 日志文件大小上限，单位M，默认: 2
	LogFileMaxSizeM int
	// 全局日志级别
	LogLevelGlobal int
	// 日志格式，默认: "%datetime %level %file %line %func %msg"，目前格式固定，该配置暂时没有使用
	LogLineFormat string
	// 日志缓冲通道容量，默认 4096
	LogChannelCap int
	// 日志缓冲通道填满后处理策略
	LogChnOverPolicy int
}

// 默认日志配置
func DefaultLogConfig() *Config {
	return &Config{
		LogFileDir:        "~/zcgologs",
		LogFileNamePrefix: "zcgolog",
		LogFileMaxSizeM:   2,
		LogLevelGlobal:    LOG_LEVEL_INFO,
		LogLineFormat:     "%level %pushTime %file %line %callFunc %msg",
		LogChannelCap:     4096,
		LogChnOverPolicy:  LOG_CHN_OVER_POLICY_DISCARD,
	}
}

// 检查日志配置
func CheckConfig(logConfig *Config) (bool, error) {
	if logConfig == nil {
		return false, fmt.Errorf("日志配置不可为空")
	}
	if logConfig.LogFileDir == "" {
		return false, fmt.Errorf("日志目录不可为空")
	}
	if logConfig.LogFileNamePrefix == "" {
		return false, fmt.Errorf("日志文件名前缀不可为空")
	}
	if logConfig.LogFileMaxSizeM <= 0 {
		return false, fmt.Errorf("日志文件Size上限必须大于0")
	}
	return true, nil
}

type logMsg struct {
	// 发生时间
	pushTime time.Time
	// 日志级别
	logLevel int
	// 日志位置-代码文件
	callFile string
	// 日志位置-代码文件行数
	callLine int
	// 日志位置-调用函数
	callFunc string
	// 日志内容
	logMsg string
	// 日志内容参数
	logParams []interface{}
}

var logConfig *Config
var logMsgChn chan logMsg
var currentLogFile *os.File
var currentLogYMD string

func InitLogger(initConfig *Config) {
	logConfig = DefaultLogConfig()
	if initConfig != nil {
		if initConfig.LogChannelCap > 0 {
			logConfig.LogChannelCap = initConfig.LogChannelCap
		}
		if initConfig.LogChnOverPolicy >= 0 && initConfig.LogChnOverPolicy < log_chn_over_policy_max {
			logConfig.LogChnOverPolicy = initConfig.LogChnOverPolicy
		}
		if initConfig.LogFileDir != "" {
			logConfig.LogFileDir = initConfig.LogFileDir
		}
		if initConfig.LogFileMaxSizeM > 0 {
			logConfig.LogFileMaxSizeM = initConfig.LogFileMaxSizeM
		}
		if initConfig.LogFileNamePrefix != "" {
			logConfig.LogFileNamePrefix = initConfig.LogFileNamePrefix
		}
		if initConfig.LogLevelGlobal >= LOG_LEVEL_DEBUG || initConfig.LogLevelGlobal < log_level_max {
			logConfig.LogLevelGlobal = initConfig.LogLevelGlobal
		}
		if initConfig.LogLineFormat != "" {
			logConfig.LogLineFormat = initConfig.LogLineFormat
		}
	}
	logMsgChn = make(chan logMsg, initConfig.LogChannelCap)
	logFilePath, todayYMD, err := GetLogFilePath(logConfig)
	if err != nil {
		log.Fatal(err)
	}
	currentLogYMD = todayYMD
	currentLogFile, err = os.OpenFile(logFilePath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatal(err)
	}
	multiWriter := io.MultiWriter(os.Stdout, currentLogFile)
	log.SetOutput(multiWriter)
	log.SetFlags(log.Ldate | log.Ltime)
	go readAndWriteMsg()
	go runLogCtlServe()
}

func readAndWriteMsg() {
	for {
		msg := <-logMsgChn
		// 检查日志文件是否需要滚动
		curLogFileStat, _ := currentLogFile.Stat()
		todayYMD := getYMDToday()
		if todayYMD != currentLogYMD || curLogFileStat.Size() >= int64(logConfig.LogFileMaxSizeM)*1024*1024 {
			logFilePath, ymd, err := GetLogFilePath(logConfig)
			if err != nil {
				fmt.Printf("log/log.go readAndWriteMsg->GetLogFilePath 发生错误: %s\n", err)
				continue
			}
			currentLogYMD = ymd
			currentLogFile, err = os.OpenFile(logFilePath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
			if err != nil {
				fmt.Printf("log/log.go readAndWriteMsg->os.OpenFile 发生错误: %s\n", err)
				continue
			}
			// 重新设置log输出目标
			multiWriter := io.MultiWriter(os.Stdout, currentLogFile)
			log.SetOutput(multiWriter)
		}
		msgPrefix := fmt.Sprintf("%s 时间:%s 代码:%s %d 函数:%s ", LogLevels[msg.logLevel], msg.pushTime.Format("2006-01-02 15:04:05"), msg.callFile, msg.callLine, msg.callFunc)
		log.Printf(msgPrefix+msg.logMsg, msg.logParams...)
	}
}

func Debug(msg string, params ...interface{}) {
	msgLogLevel := LOG_LEVEL_DEBUG
	pushMsg(msgLogLevel, msg, params...)
}

func Info(msg string, params ...interface{}) {
	msgLogLevel := LOG_LEVEL_INFO
	pushMsg(msgLogLevel, msg, params...)
}

func pushMsg(msgLogLevel int, msg string, params ...interface{}) {
	pc, file, line, _ := runtime.Caller(2)
	myFunc := runtime.FuncForPC(pc).Name()
	// fmt.Printf("myFunc: %s\n", myFunc)
	myLevel := logLevelCtl[myFunc]
	if myLevel == 0 {
		myLevel = logLevelCtl["default"]
	}
	if myLevel > msgLogLevel {
		// fmt.Printf("myLevel > msgLogLevel %d > %d\n", myLevel, msgLogLevel)
		return
	}
	pushMsg := logMsg{
		pushTime:  time.Now(),
		logLevel:  msgLogLevel,
		callFile:  file,
		callLine:  line,
		callFunc:  myFunc,
		logMsg:    msg,
		logParams: params,
	}
	switch logConfig.LogChnOverPolicy {
	case LOG_CHN_OVER_POLICY_BLOCK:
		logMsgChn <- pushMsg
	case LOG_CHN_OVER_POLICY_DISCARD:
		select {
		case logMsgChn <- pushMsg:
			return
		default:
			fmt.Printf("日志缓冲通道已满，该条日志被丢弃:"+msg+"\n", params...)
			return
		}
	}
}

// DEBUG日志输出控制
var logLevelCtl = map[string]int{
	"default": LOG_LEVEL_INFO,
}

func handleLogLevelCtl(w http.ResponseWriter, req *http.Request) {
	query := req.URL.Query()
	logger := query.Get("logger")
	level := query.Get("level")
	var err error
	// fmt.Printf("logger: %s\n", logger)
	// fmt.Printf("logLevelCtl[logger] before: %d\n", logLevelCtl[logger])
	logLevelCtl[logger], err = strconv.Atoi(level)
	// fmt.Printf("logLevelCtl[logger] after: %d\n", logLevelCtl[logger])
	if err != nil {
		fmt.Fprintf(w, "发生错误: %s\n", err)
	} else {
		fmt.Fprintf(w, "操作成功\n")
	}
}

func runLogCtlServe() {
	http.HandleFunc("/zcgolog/level/ctl", handleLogLevelCtl)
	log.Fatal(http.ListenAndServe("localhost:9300", nil))
}
