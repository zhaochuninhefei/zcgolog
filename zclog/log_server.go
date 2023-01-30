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
zclog/log_server.go zcgolog服务器模式相关代码
*/

import (
	"fmt"
	"io"
	"os"
	"sync"
	"time"
)

// 日志缓冲通道填满后处理策略定义
//goland:noinspection GoSnakeCaseUsage
const (
	// LOG_CHN_OVER_POLICY_DISCARD 丢弃该条日志
	LOG_CHN_OVER_POLICY_DISCARD = iota + 1
	// LOG_CHN_OVER_POLICY_BLOCK 阻塞等待
	LOG_CHN_OVER_POLICY_BLOCK
	// log_chn_over_policy_max 日志缓冲通道填满后处理策略定义值上限
	log_chn_over_policy_max
)

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

// 日志缓冲通道
var logMsgChn chan logMsg

// 退出通道,用于监听是否需要退出对日志缓冲通道的监听。
// 服务器模式下刷新日志配置重启服务器模式前，需要先通过该通道，通知当前对日志缓冲通道的监听服务退出。
var quitChn = make(chan int)

// 启动zcgolog服务器模式
func startZcgologServer() {
	// 初始化zcgologger
	initZcgoLogger()
	// 启动日志缓冲通道监听
	go readAndWriteMsg()
	// 等待日志级别控制监听服务启动，
	// 防止runLogCtlServe执行时日志缓冲通道尚未初始化。
	_ = waitMsgReaderStart(3000)
	// 启动日志级别控制监听服务
	runLogCtlServeOnce.Do(startLogCtlServe)
}

// QuitMsgReader 停止对缓冲消息通道的监听
//  timeoutMilliSec 超时时间(毫秒),该值<=0时表示会一直等待直到监听停止。
func QuitMsgReader(timeoutMilliSec int) error {
	if !msgReaderRunning {
		return nil
	}
	// 请求停止对日志缓冲通道的监听
	quitChn <- 1
	startTime := time.Now()
	// 自旋等待日志缓冲通道监听停止
	for {
		time.Sleep(time.Millisecond * 500)
		if !msgReaderRunning {
			return nil
		}
		if timeoutMilliSec > 0 {
			timeNow := time.Now()
			if timeoutMilliSec < int(timeNow.Sub(startTime).Milliseconds()) {
				return fmt.Errorf("日志缓冲通道监听未能在超时时间内停止, 超时时间(毫秒): %d", timeoutMilliSec)
			}
		}
	}
}

// 等待日志缓冲通道监听启动
//  timeoutMilliSec 超时时间(毫秒),该值<=0时表示会一直等待直到监听启动。
func waitMsgReaderStart(timeoutMilliSec int) error {
	startTime := time.Now()
	for {
		if msgReaderRunning {
			return nil
		}
		if timeoutMilliSec > 0 {
			timeNow := time.Now()
			if timeoutMilliSec < int(timeNow.Sub(startTime).Milliseconds()) {
				return fmt.Errorf("日志缓冲通道监听未能在超时时间内启动, 超时时间(毫秒): %d", timeoutMilliSec)
			}
		}
		time.Sleep(time.Millisecond * 500)
	}
}

var msgReaderLock sync.Mutex
var msgReaderRunning = false

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
	// 上锁,确保logger操作的线程安全
	loggerLock.Lock()
	defer loggerLock.Unlock()
	// 临时切换zcgoLogger输出到控制台
	zcgoLogger.SetOutput(os.Stdout)
	closeCurrentLogFile()
	logFilePath, ymd, err := GetLogFilePathAndYMDToday(zcgologConfig)
	if err != nil {
		// 获取最新日志文件失败时，直接向控制台输出
		currentLogYMD = getYMDToday()
		currentLogFile = nil
		zcgoLogger.Printf("zclog/log.go readAndWriteMsg->GetLogFilePathAndYMDToday 发生错误: %s", err)
		return
	}
	currentLogYMD = ymd
	currentLogFile, err = os.OpenFile(logFilePath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, os.ModePerm)
	if err != nil {
		// 获取最新日志文件失败时，直接向控制台输出
		currentLogFile = nil
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

// 将日志消息推送到日志缓冲通道
func pushMsgToLogMsgChn(pushMsg logMsg) {
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
			fmt.Printf("日志缓冲通道已满，该条日志被丢弃:"+pushMsg.logMsg+"\n", pushMsg.logParams...)
			return
		}
		// TODO 考虑是否添加新的策略，比如将日志直接输出到fallback的输出流?
	}
}
