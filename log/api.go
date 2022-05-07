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

func Debug(msg string, params ...interface{}) {
	msgLogLevel := LOG_LEVEL_DEBUG
	pushMsg(msgLogLevel, msg, params...)
}

func Info(msg string, params ...interface{}) {
	msgLogLevel := LOG_LEVEL_INFO
	pushMsg(msgLogLevel, msg, params...)
}

func Warn(msg string, params ...interface{}) {
	msgLogLevel := LOG_LEVEL_WARNING
	pushMsg(msgLogLevel, msg, params...)
}

func Error(msg string, params ...interface{}) {
	msgLogLevel := LOG_LEVEL_ERROR
	pushMsg(msgLogLevel, msg, params...)
}

func Critical(msg string, params ...interface{}) {
	msgLogLevel := LOG_LEVEL_CRITICAL
	pushMsg(msgLogLevel, msg, params...)
}

func Fatal(msg string, params ...interface{}) {
	msgLogLevel := LOG_LEVEL_FATAL
	pushMsg(msgLogLevel, msg, params...)
}

func Debugf(msg string, params ...interface{}) {
	msgLogLevel := LOG_LEVEL_DEBUG
	pushMsg(msgLogLevel, msg, params...)
}

func Infof(msg string, params ...interface{}) {
	msgLogLevel := LOG_LEVEL_INFO
	pushMsg(msgLogLevel, msg, params...)
}

func Warnf(msg string, params ...interface{}) {
	msgLogLevel := LOG_LEVEL_WARNING
	pushMsg(msgLogLevel, msg, params...)
}

func Errorf(msg string, params ...interface{}) {
	msgLogLevel := LOG_LEVEL_ERROR
	pushMsg(msgLogLevel, msg, params...)
}

func Criticalf(msg string, params ...interface{}) {
	msgLogLevel := LOG_LEVEL_CRITICAL
	pushMsg(msgLogLevel, msg, params...)
}

func Fatalf(msg string, params ...interface{}) {
	msgLogLevel := LOG_LEVEL_FATAL
	pushMsg(msgLogLevel, msg, params...)
}

func Debugln(msg string, params ...interface{}) {
	msgLogLevel := LOG_LEVEL_DEBUG
	pushMsg(msgLogLevel, msg, params...)
}

func Infoln(msg string, params ...interface{}) {
	msgLogLevel := LOG_LEVEL_INFO
	pushMsg(msgLogLevel, msg, params...)
}

func Warnln(msg string, params ...interface{}) {
	msgLogLevel := LOG_LEVEL_WARNING
	pushMsg(msgLogLevel, msg, params...)
}

func Errorln(msg string, params ...interface{}) {
	msgLogLevel := LOG_LEVEL_ERROR
	pushMsg(msgLogLevel, msg, params...)
}

func Criticalln(msg string, params ...interface{}) {
	msgLogLevel := LOG_LEVEL_CRITICAL
	pushMsg(msgLogLevel, msg, params...)
}

func Fatalln(msg string, params ...interface{}) {
	msgLogLevel := LOG_LEVEL_FATAL
	pushMsg(msgLogLevel, msg, params...)
}
