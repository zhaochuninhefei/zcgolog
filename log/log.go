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
	"path"
	"runtime"
	"strconv"
	"sync"
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
	LOG_CHN_OVER_POLICY_DISCARD = iota + 1
	// 阻塞等待
	LOG_CHN_OVER_POLICY_BLOCK
	log_chn_over_policy_max
)

// 日志模式定义
const (
	// 本地模式: 日志同步输出且不支持在线修改指定logger的日志级别，日志文件不支持自动滚动，通常仅用于测试
	LOG_MODE_LOCAL = iota + 1
	// 服务器模式: 日志异步输出且支持在线修改指定logger的日志级别，日志文件支持自动滚动
	LOG_MODE_SERVER
	log_mode_max
)

// 日志配置
type Config struct {
	// 日志文件目录，默认: ~/zcgologs
	LogFileDir string
	// 日志文件名前缀，默认: zcgolog
	LogFileNamePrefix string
	// 日志文件大小上限，单位M，默认: 2
	LogFileMaxSizeM int
	// 全局日志级别，默认:DEBUG
	LogLevelGlobal int
	// 日志格式，默认: "%datetime %level %file %line %func %msg"，目前格式固定，该配置暂时没有使用
	LogLineFormat string
	// 日志模式，默认采用本地模式，以便于本地测试
	LogMod int
	// 日志缓冲通道容量，默认 4096
	LogChannelCap int
	// 日志缓冲通道填满后处理策略，默认:LOG_CHN_OVER_POLICY_DISCARD 丢弃该条日志
	LogChnOverPolicy int
	// 日志级别控制监听服务的Host，默认:localhost
	LogLevelCtlHost string
	// 日志级别控制监听服务的Port，默认:9300
	LogLevelCtlPort string
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
	if logConfig.LogLevelGlobal < LOG_LEVEL_DEBUG || logConfig.LogLevelGlobal >= log_level_max {
		return false, fmt.Errorf("全局日志级别不能超出有效范围")
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

var zcgologConfig *Config
var logMsgChn chan logMsg
var currentLogFile *os.File
var currentLogYMD string

var initDefaultLogConfigOnce sync.Once

// 初始化默认日志配置
func initDefaultLogConfig() {
	homeDir, err := Home()
	if err != nil {
		homeDir = "~"
	}
	defaultLogFileDir := path.Join(homeDir, "zcgologs")
	os.MkdirAll(defaultLogFileDir, os.ModePerm)
	zcgologConfig = &Config{
		LogFileDir:        defaultLogFileDir,
		LogFileNamePrefix: "zcgolog",
		LogFileMaxSizeM:   2,
		LogLevelGlobal:    LOG_LEVEL_DEBUG,
		LogLineFormat:     "%level %pushTime %file %line %callFunc %msg",
		LogChannelCap:     4096,
		LogChnOverPolicy:  LOG_CHN_OVER_POLICY_DISCARD,
		LogMod:            LOG_MODE_LOCAL,
		LogLevelCtlHost:   "localhost",
		LogLevelCtlPort:   "9300",
	}
}

var initLogOnce sync.Once

// 初始化log
//  获取日志文件，当天年月日，配置日志输出，配置日志前缀时间戳格式
func initLog() {
	// 获取日志文件
	logFilePath, todayYMD, err := GetLogFilePath(zcgologConfig)
	if err != nil {
		log.Fatal(err)
	}
	currentLogYMD = todayYMD
	currentLogFile, err = os.OpenFile(logFilePath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, os.ModePerm)
	if err != nil {
		log.Fatal(err)
	}
	// 日志同时输出到日志文件与控制台
	multiWriter := io.MultiWriter(os.Stdout, currentLogFile)
	log.SetOutput(multiWriter)
	// 日志前缀时间戳格式
	log.SetFlags(log.Ldate | log.Ltime)
}

// 初始化zcgolog
func InitLogger(initConfig *Config) {
	// 初始化默认配置(只会执行一次)
	initDefaultLogConfigOnce.Do(initDefaultLogConfig)
	// 从参数中获取有效配置覆盖logConfig
	if initConfig != nil {
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
	// 根据日志模式决定是否启用日志缓冲队列与在线修改日志级别功能
	switch zcgologConfig.LogMod {
	case LOG_MODE_SERVER:
		// 服务器模式下，每次调用InitLogger都会重新初始化log
		initLog()
		// 初始化日志缓冲队列
		logMsgChn = make(chan logMsg, zcgologConfig.LogChannelCap)
		// 启动日志拉取与输出
		go readAndWriteMsg()
		// 启动日志级别控制监听服务
		go runLogCtlServe()
	case LOG_MODE_LOCAL:
		// 本地日志模式下，初始化log只执行一次
		// 但本地模式一般不会调用InitLogger，因此需要在具体调用日志输出函数时再执行一次 initLogOnce.Do(initLog)
		initLogOnce.Do(initLog)
	}
}

// 从日志缓冲队列拉取并输出日志
func readAndWriteMsg() {
	for {
		msg := <-logMsgChn
		// 检查日志文件是否需要滚动
		curLogFileStat, _ := currentLogFile.Stat()
		todayYMD := getYMDToday()
		if todayYMD != currentLogYMD || curLogFileStat.Size() >= int64(zcgologConfig.LogFileMaxSizeM)*1024*1024 {
			logFilePath, ymd, err := GetLogFilePath(zcgologConfig)
			if err != nil {
				fmt.Printf("log/log.go readAndWriteMsg->GetLogFilePath 发生错误: %s\n", err)
				continue
			}
			currentLogYMD = ymd
			currentLogFile, err = os.OpenFile(logFilePath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, os.ModePerm)
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

func pushMsg(msgLogLevel int, msg string, params ...interface{}) {
	// 初始化默认配置(只会执行一次)
	initDefaultLogConfigOnce.Do(initDefaultLogConfig)
	// 获取调用堆栈信息
	pc, file, line, _ := runtime.Caller(2)
	// 调用处函数包路径
	myFunc := runtime.FuncForPC(pc).Name()
	// fmt.Printf("myFunc: %s\n", myFunc)
	// 获取函数对应的日志级别
	myLevel := logLevelCtl[myFunc]
	if myLevel == 0 {
		myLevel = zcgologConfig.LogLevelGlobal
	}
	// 判断该日志是否需要输出
	if myLevel > msgLogLevel {
		// fmt.Printf("myLevel > msgLogLevel %d > %d\n", myLevel, msgLogLevel)
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
		// 根据LogChnOverPolicy决定是否在缓冲通道已满时阻塞
		switch zcgologConfig.LogChnOverPolicy {
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
	case LOG_MODE_LOCAL:
		// 本地日志模式下，初始化log只会执行一次
		initLogOnce.Do(initLog)
		msgPrefix := fmt.Sprintf("%s 代码:%s %d 函数:%s ", LogLevels[msgLogLevel], file, line, myFunc)
		log.Printf(msgPrefix+msg, params...)
	}
}

// DEBUG日志输出控制
var logLevelCtl = map[string]int{}

// 处理日志级别调整请求
//  URL参数为logger和level;
//  logger是调整目标，对应具体函数的完整包名路径，如: gitee.com/zhaochuninhefei/zcgolog/log.writeLog
//  level是调整后的日志级别，支持从1到6，分别是 DEBUG,INFO,WARNNING,ERROR,CRITICAL,FATAL
//  一个完整的请求URL示例:http://localhost:9300/zcgolog/api/level/ctl?logger=gitee.com/zhaochuninhefei/zcgolog/log.writeLog&level=1
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

// 启动日志级别控制监听服务
//  host与端口取决于具体的日志配置;
//  URI固定为/zcgolog/api/level/ctl;
//  URL参数为logger和level;
//  logger是调整目标，对应具体函数的完整包名路径，如: gitee.com/zhaochuninhefei/zcgolog/log.writeLog ;
//  level是调整后的日志级别，支持从1到6，分别是 DEBUG,INFO,WARNNING,ERROR,CRITICAL,FATAL ;
//  一个完整的请求URL示例:http://localhost:9300/zcgolog/api/level/ctl?logger=gitee.com/zhaochuninhefei/zcgolog/log.writeLog&level=1
func runLogCtlServe() {
	listenAddress := zcgologConfig.LogLevelCtlHost + ":" + zcgologConfig.LogLevelCtlPort
	http.HandleFunc("/zcgolog/api/level/ctl", handleLogLevelCtl)
	Info("启动监听: http://%s/zcgolog/api/level/ctl", listenAddress)
	log.Fatal(http.ListenAndServe(listenAddress, nil))
}
