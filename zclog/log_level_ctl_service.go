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
zclog/log_level_ctl_service.go 日志级别控制服务，提供httpAPI来在线调整日志级别
*/

import (
	"fmt"
	"log"
	"net/http"
	"strconv"
	"sync"
)

// 日志级别控制
var logLevelCtl = map[string]int{}
var Level = LOG_LEVEL_INFO
var runLogCtlServeOnce sync.Once

// 处理指定函数的日志级别调整请求
//  URL参数为logger和level;
//  logger是调整目标，对应具体函数的完整包名路径，如: gitee.com/zhaochuninhefei/zcgolog/log.writeLog
//  level是调整后的日志级别，支持从1到6，分别是 DEBUG,INFO,WARNNING,ERROR,PANIC,FATAL
//  一个完整的请求URL示例:http://localhost:9300/zcgolog/api/level/ctl?logger=gitee.com/zhaochuninhefei/zcgolog/zclog.writeLog&level=1
func handleLogLevelCtl(w http.ResponseWriter, req *http.Request) {
	query := req.URL.Query()
	logger := query.Get("logger")
	level := query.Get("level")
	targetLevel, err := strconv.Atoi(level)
	if err != nil {
		_, err := fmt.Fprintf(w, "发生错误: %s\n", err)
		if err != nil {
			log.Printf("发生预期外错误: %s", err)
			return
		}
	} else {
		if targetLevel >= LOG_LEVEL_DEBUG && targetLevel < log_level_max {
			logLevelCtl[logger] = targetLevel
			_, err := fmt.Fprintf(w, "操作成功\n")
			if err != nil {
				log.Printf("发生预期外错误: %s", err)
				return
			}
		} else {
			logLevelCtl[logger] = 0
			_, err := fmt.Fprintf(w, "传入的level不在有效范围,目标函数仍采取全局日志级别\n")
			if err != nil {
				log.Printf("发生预期外错误: %s", err)
				return
			}
		}
	}
}

// 处理全局日志级别调整请求
func handleLogLevelCtlGlobal(w http.ResponseWriter, req *http.Request) {
	query := req.URL.Query()
	level := query.Get("level")
	targetLevel, err := strconv.Atoi(level)
	if err != nil {
		_, err := fmt.Fprintf(w, "发生错误: %s\n", err)
		if err != nil {
			log.Printf("发生预期外错误: %s", err)
			return
		}
	} else {
		if targetLevel >= LOG_LEVEL_DEBUG && targetLevel < log_level_max {
			Level = targetLevel
			_, err := fmt.Fprintf(w, "操作成功\n")
			if err != nil {
				log.Printf("发生预期外错误: %s", err)
				return
			}
		} else {
			Level = zcgologConfig.LogLevelGlobal
			_, err := fmt.Fprintf(w, "传入的level不在有效范围,全局日志级别恢复为启动配置\n")
			if err != nil {
				log.Printf("发生预期外错误: %s", err)
				return
			}
		}
	}
}

// 处理日志级别查询请求
func handleLogLevelQuery(w http.ResponseWriter, req *http.Request) {
	query := req.URL.Query()
	logger := query.Get("logger")
	var resultLevel int
	if logger != "" {
		resultLevel = logLevelCtl[logger]
	}
	if resultLevel == 0 {
		resultLevel = Level
	}
	result := GetLogLevelStrByInt(resultLevel)
	_, err := fmt.Fprint(w, result)
	if err != nil {
		log.Printf("发生预期外错误: %s", err)
		return
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
	http.HandleFunc("/zcgolog/api/level/global", handleLogLevelCtlGlobal)
	http.HandleFunc("/zcgolog/api/level/query", handleLogLevelQuery)
	//goland:noinspection HttpUrlsUsage
	Infof("启动日志级别控制监听服务: [http://%s/zcgolog/api/level/**]", listenAddress)
	zcgoLogger.Fatal(http.ListenAndServe(listenAddress, nil))
}

// 异步启动日志级别控制监听服务
func startLogCtlServe() {
	go runLogCtlServe()
}
