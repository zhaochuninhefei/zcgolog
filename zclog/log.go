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
	"io"
	"log"
	"net/http"
	"os"
	"runtime"
	"strconv"
	"sync"
	"time"
)

// 日志级别定义
const (
	// debug 调试日志，生产环境通常关闭
	LOG_LEVEL_DEBUG = iota + 1
	// info 重要信息日志，用于提示程序过程中的一些重要信息，慎用，避免过多的INFO日志
	LOG_LEVEL_INFO
	// warning 警告日志，用于警告用户可能会发生问题
	LOG_LEVEL_WARNING
	// error 一般错误日志，一般用于提示业务错误，程序通常不会因为这样的错误终止
	LOG_LEVEL_ERROR
	// panic 异常错误日志，一般用于预期外的错误，程序的当前Goroutine会终止并输出堆栈信息
	LOG_LEVEL_PANIC
	// fatal 致命错误日志，程序会马上终止
	LOG_LEVEL_FATAL
	// 日志级别最大值，用于内部判断日志级别是否在合法范围内
	log_level_max
)

// 日志级别文字定义
var LogLevels = [...]string{
	LOG_LEVEL_DEBUG:   "[DEBUG]",
	LOG_LEVEL_INFO:    "[ INFO]",
	LOG_LEVEL_WARNING: "[ WARN]",
	LOG_LEVEL_ERROR:   "[ERROR]",
	LOG_LEVEL_PANIC:   "[PANIC]",
	LOG_LEVEL_FATAL:   "[FATAL]",
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
	// 是否需要同时输出到控制台，默认: false
	LogForbidStdout bool
	// 日志文件目录，默认: 空，此时日志只输出到控制台
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

// 日志消息，日志缓冲通道用
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

// zcgoLogger
var zcgoLogger *log.Logger

// zcgolog配置
var zcgologConfig *Config

// 日志缓冲通道
var logMsgChn chan logMsg

// 退出通道,用于监听是否需要退出对日志缓冲通道的监听。
// 服务器模式下刷新日志配置重启服务器模式前，需要先通过该通道，通知当前对日志缓冲通道的监听服务退出。
var quitChn = make(chan int)

// 当前日志文件
var currentLogFile *os.File

func closeCurrentLogFile() {
	if currentLogFile != nil {
		currentLogFile.Close()
	}
}

// 当天日期
var currentLogYMD string

var initDefaultLogConfigOnce sync.Once

// 初始化默认日志配置
func initDefaultLogConfig() {
	// homeDir, err := Home()
	// if err != nil {
	// 	homeDir = "~"
	// }
	// defaultLogFileDir := path.Join(homeDir, "zcgologs")
	// os.MkdirAll(defaultLogFileDir, os.ModePerm)
	zcgologConfig = &Config{
		LogForbidStdout:   false,
		LogFileDir:        "",
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

// 服务器模式下配置golang的log
func configGolangLogForServer() {
	// // 日志前缀时间戳格式
	// log.SetFlags(log.Ldate | log.Ltime)
	// 关闭当前日志文件
	closeCurrentLogFile()
	// 获取最新日志文件
	logFilePath, todayYMD, err := GetLogFilePathAndYMDToday(zcgologConfig)
	if err != nil {
		// 服务器模式下，日志文件必须存在
		log.Panic(err)
	}
	currentLogYMD = todayYMD
	currentLogFile, err = os.OpenFile(logFilePath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, os.ModePerm)
	if err != nil {
		// 服务器模式下，日志文件必须成功打开
		log.Panic(err)
	}
	if !zcgologConfig.LogForbidStdout {
		// 日志同时输出到日志文件与控制台
		multiWriter := io.MultiWriter(os.Stdout, currentLogFile)
		zcgoLogger = log.New(multiWriter, "", log.Ldate|log.Ltime)
		// log.SetOutput(multiWriter)
	} else {
		// 日志只输出到日志文件
		// log.SetOutput(currentLogFile)
		zcgoLogger = log.New(currentLogFile, "", log.Ldate|log.Ltime)
	}
}

var configGolangLogForLocalOnce sync.Once

// 本地模式下配置golang的log
//  获取日志文件，当天年月日，配置日志输出，配置日志前缀时间戳格式
func configGolangLogForLocal() {
	// // 日志前缀时间戳格式
	// log.SetFlags(log.Ldate | log.Ltime)
	// 关闭当前日志文件
	closeCurrentLogFile()
	// 获取最新日志文件
	logFilePath, todayYMD, err := GetLogFilePathAndYMDToday(zcgologConfig)
	if err != nil {
		// 未能成功获取日志文件时，直接输出到控制台
		currentLogYMD = getYMDToday()
		currentLogFile = nil
		zcgoLogger = log.New(os.Stdout, "", log.Ldate|log.Ltime)
		// log.SetOutput(os.Stdout)
		zcgoLogger.Println(err.Error())
		return
	}
	if logFilePath == OS_OUT_STDOUT {
		// LogFileDir为空时，直接输出到控制台
		currentLogYMD = todayYMD
		currentLogFile = nil
		zcgoLogger = log.New(os.Stdout, "", log.Ldate|log.Ltime)
		// log.SetOutput(os.Stdout)
		return
	}
	currentLogYMD = todayYMD
	currentLogFile, err = os.OpenFile(logFilePath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, os.ModePerm)
	if err != nil {
		// 未能成功获取日志文件时，直接输出到控制台
		currentLogFile = nil
		zcgoLogger = log.New(os.Stdout, "", log.Ldate|log.Ltime)
		// log.SetOutput(os.Stdout)
		zcgoLogger.Println(err.Error())
		return
	}
	if !zcgologConfig.LogForbidStdout {
		// 日志同时输出到日志文件与控制台
		multiWriter := io.MultiWriter(os.Stdout, currentLogFile)
		// log.SetOutput(multiWriter)
		zcgoLogger = log.New(multiWriter, "", log.Ldate|log.Ltime)
	} else {
		// 日志只输出到日志文件
		// log.SetOutput(currentLogFile)
		zcgoLogger = log.New(currentLogFile, "", log.Ldate|log.Ltime)
	}
}

// 初始化zcgolog
func InitLogger(initConfig *Config) {
	// 初始化默认配置(只会执行一次)
	initDefaultLogConfigOnce.Do(initDefaultLogConfig)
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
	// 根据日志模式决定是否启用日志缓冲队列与在线修改日志级别功能
	switch zcgologConfig.LogMod {
	case LOG_MODE_SERVER:
		// fmt.Println("LOG_MODE_SERVER zcgologConfig.LogMod:", zcgologConfig.LogMod)
		// 重启zcgolog服务器模式
		restartZcgologServer()
	case LOG_MODE_LOCAL:
		// fmt.Println("LOG_MODE_LOCAL zcgologConfig.LogMod:", zcgologConfig.LogMod)
		// 本地日志模式下，初始化log只执行一次
		// 但本地模式可能不会调用InitLogger，因此需要在首次调用日志输出函数时执行一次
		configGolangLogForLocalOnce.Do(configGolangLogForLocal)
	}
}

// 重启zcgolog服务器模式
func restartZcgologServer() {
	// 请求停止日志缓冲通道监听
	err := QuitMsgReader(3000)
	if err != nil {
		log.Panic(err)
	}
	// 根据最新的zcgologConfig配置golang的log
	configGolangLogForServer()
	// 启动日志缓冲通道监听
	go readAndWriteMsg()
	// 等待3秒再启动日志级别控制监听服务，
	// 防止runLogCtlServe执行时日志缓冲通道尚未初始化。
	time.Sleep(3 * time.Second)
	// 启动日志级别控制监听服务
	runLogCtlServeOnce.Do(startLogCtlServe)
}

// 停止对缓冲消息通道的监听
func QuitMsgReader(timeoutMicroSec int) error {
	if msgReaderRunning {
		quitChn <- 1
	}
	startTime := time.Now()
	for {
		time.Sleep(time.Microsecond * 500)
		if !msgReaderRunning {
			return nil
		}
		if timeoutMicroSec > 0 {
			timeNow := time.Now()
			if timeoutMicroSec < int(timeNow.Sub(startTime).Microseconds()) {
				return fmt.Errorf("日志缓冲通道监听未能在超时时间内停止, 超时时间(毫秒): %d", timeoutMicroSec)
			}
		}
	}
}

var msgReaderLock sync.Mutex
var msgReaderRunning bool = false

// 从日志缓冲通道拉取并输出日志
func readAndWriteMsg() {
	// 通过排他锁控制同时只能有一个Goroutine执行该函数
	msgReaderLock.Lock()
	defer msgReaderLock.Unlock()
	// 初始化日志缓冲通道
	logMsgChn = make(chan logMsg, zcgologConfig.LogChannelCap)
	Info("readAndWriteMsg开始")
	msgReaderRunning = true
	defer closeCurrentLogFile()
	for {
		// select IO多路复用 监听日志缓冲通道和退出通道
		select {
		case <-quitChn:
			// 接收到退出指令
			msgReaderRunning = false
			zcgoLogger.Println("readAndWriteMsg结束")
			return
		case msg := <-logMsgChn:
			// 接收到日志消息
			// 检查日志文件是否需要滚动
			if currentLogFile != nil {
				curLogFileStat, _ := currentLogFile.Stat()
				todayYMD := getYMDToday()
				// 当天日期发生变化或当前日志文件大小超过上限时，做日志文件滚动处理
				if todayYMD != currentLogYMD || curLogFileStat.Size() >= int64(zcgologConfig.LogFileMaxSizeM)*1024*1024 {
					scrollLogFile()
				}
			}
			msgPrefix := fmt.Sprintf("%s 时间:%s 代码:%s %d 函数:%s ", LogLevels[msg.logLevel], msg.pushTime.Format(LOG_TIME_FORMAT_YMDHMS), msg.callFile, msg.callLine, msg.callFunc)
			zcgoLogger.Printf(msgPrefix+msg.logMsg, msg.logParams...)
		}
	}
}

// 日志文件滚动处理
func scrollLogFile() {
	closeCurrentLogFile()
	logFilePath, ymd, err := GetLogFilePathAndYMDToday(zcgologConfig)
	if err != nil {
		// 获取最新日志文件失败时，直接向控制台输出
		currentLogYMD = getYMDToday()
		currentLogFile = nil
		zcgoLogger.SetOutput(os.Stdout)
		zcgoLogger.Printf("zclog/log.go readAndWriteMsg->GetLogFilePathAndYMDToday 发生错误: %s", err)
		return
	}
	currentLogYMD = ymd
	currentLogFile, err = os.OpenFile(logFilePath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, os.ModePerm)
	if err != nil {
		// 获取最新日志文件失败时，直接向控制台输出
		currentLogFile = nil
		zcgoLogger.SetOutput(os.Stdout)
		zcgoLogger.Printf("zclog/log.go readAndWriteMsg->os.OpenFile 发生错误: %s", err)
		return
	}
	// 重新设置log输出目标
	if !zcgologConfig.LogForbidStdout {
		// 日志同时输出到日志文件与控制台
		multiWriter := io.MultiWriter(os.Stdout, currentLogFile)
		zcgoLogger.SetOutput(multiWriter)
	} else {
		// 日志只输出到日志文件
		zcgoLogger.SetOutput(currentLogFile)
	}
}

// 输出日志
func outputLog(msgLogLevel int, msg string, params ...interface{}) {
	// 初始化默认配置(只会执行一次)
	initDefaultLogConfigOnce.Do(initDefaultLogConfig)
	// 获取调用堆栈信息
	pc, file, line, _ := runtime.Caller(2)
	// 调用处函数包路径
	myFunc := runtime.FuncForPC(pc).Name()
	// Panic与Fatal直接调用log包处理
	if msgLogLevel == LOG_LEVEL_PANIC {
		configGolangLogForLocalOnce.Do(configGolangLogForLocal)
		msgPrefix := fmt.Sprintf("%s 代码:%s %d 函数:%s ", LogLevels[msgLogLevel], file, line, myFunc)
		zcgoLogger.Panicf(msgPrefix+msg, params...)
	}
	if msgLogLevel == LOG_LEVEL_FATAL {
		configGolangLogForLocalOnce.Do(configGolangLogForLocal)
		msgPrefix := fmt.Sprintf("%s 代码:%s %d 函数:%s ", LogLevels[msgLogLevel], file, line, myFunc)
		zcgoLogger.Fatalf(msgPrefix+msg, params...)
	}
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
		if msgReaderRunning {
			// 根据LogChnOverPolicy决定是否在缓冲通道已满时阻塞
			switch zcgologConfig.LogChnOverPolicy {
			case LOG_CHN_OVER_POLICY_BLOCK:
				// 阻塞模式下，如果缓冲通道已满，则当前goroutine将在此阻塞等待，
				// 直到下游readAndWriteMsg的goroutine将消息拉走，缓冲通道有空间空出来。
				logMsgChn <- pushMsg
			case LOG_CHN_OVER_POLICY_DISCARD:
				// 丢弃模式下，如果缓冲通道已满，则进入select的default分支，丢弃该条日志，但会直接在控制台输出。
				select {
				case logMsgChn <- pushMsg:
					return
				default:
					fmt.Printf("日志缓冲通道已满，该条日志被丢弃:"+msg+"\n", params...)
					return
				}
				// TODO 考虑是否添加新的策略，比如将日志直接输出到fallback的输出流?
			}
		} else {
			// 服务器模式下日志缓冲通道监听服务已停止时，改为本地模式输出日志
			configGolangLogForLocalOnce.Do(configGolangLogForLocal)
			msgPrefix := fmt.Sprintf("%s 代码:%s %d 函数:%s ", LogLevels[msgLogLevel], file, line, myFunc)
			zcgoLogger.Printf(msgPrefix+msg, params...)
		}
	case LOG_MODE_LOCAL:
		// 本地日志模式下，configGolangLog只会执行一次
		configGolangLogForLocalOnce.Do(configGolangLogForLocal)
		msgPrefix := fmt.Sprintf("%s 代码:%s %d 函数:%s ", LogLevels[msgLogLevel], file, line, myFunc)
		zcgoLogger.Printf(msgPrefix+msg, params...)
	}
}

// DEBUG日志输出控制
var logLevelCtl = map[string]int{}
var runLogCtlServeOnce sync.Once

// 处理日志级别调整请求
//  URL参数为logger和level;
//  logger是调整目标，对应具体函数的完整包名路径，如: gitee.com/zhaochuninhefei/zcgolog/log.writeLog
//  level是调整后的日志级别，支持从1到6，分别是 DEBUG,INFO,WARNNING,ERROR,CRITICAL,FATAL
//  一个完整的请求URL示例:http://localhost:9300/zcgolog/api/level/ctl?logger=gitee.com/zhaochuninhefei/zcgolog/zclog.writeLog&level=1
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
//  一个完整的请求URL示例:http://localhost:9300/zcgolog/api/level/ctl?logger=gitee.com/zhaochuninhefei/zcgolog/zclog.writeLog&level=1
func runLogCtlServe() {
	listenAddress := zcgologConfig.LogLevelCtlHost + ":" + zcgologConfig.LogLevelCtlPort
	http.HandleFunc("/zcgolog/api/level/ctl", handleLogLevelCtl)
	Infof("启动监听: http://%s/zcgolog/api/level/ctl", listenAddress)
	zcgoLogger.Fatal(http.ListenAndServe(listenAddress, nil))
}

// 异步启动日志级别控制监听服务
func startLogCtlServe() {
	go runLogCtlServe()
}
