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
zclog/compatibility.go 为了兼容golang的log包，额外提供与log包相同的一些函数
*/

import (
	"io"
	"log"
)

func New(out io.Writer, prefix string, flag int) *log.Logger {
	return log.New(out, prefix, flag)
}

func Default() *log.Logger {
	return zcgoLogger
}

func SetOutput(w io.Writer) {
	zcgoLogger.SetOutput(w)
}

func Flags() int {
	return zcgoLogger.Flags()
}

func SetFlags(flag int) {
	zcgoLogger.SetFlags(flag)
}

func Prefix() string {
	return zcgoLogger.Prefix()
}

func SetPrefix(prefix string) {
	zcgoLogger.SetPrefix(prefix)
}

func Writer() io.Writer {
	return zcgoLogger.Writer()
}

func Output(calldepth int, s string) error {
	return zcgoLogger.Output(calldepth+1, s) // +1 for this frame.
}
