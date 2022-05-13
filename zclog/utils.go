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
zclog/utils.go zcgolog相关通用处理
*/

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"os/user"
	"runtime"
	"strings"
)

const (
	LOG_TIME_FORMAT_YMDHMS        = "2006-01-02 15:04:05"
	CONFIG_CHECK_RESULT_NG        = -1
	CONFIG_CHECK_RESULT_OK        = 1
	CONFIG_CHECK_RESULT_NOFILEDIR = 2
)

// 获取当前用户Home目录
func Home() (string, error) {
	// 优先使用当前系统用户的Home目录
	user, err := user.Current()
	if nil == err {
		return user.HomeDir, nil
	}
	// 判断操作系统是否windows
	if runtime.GOOS == "windows" {
		return homeWindows()
	}
	return homeUnix()
}

func homeUnix() (string, error) {
	if home := os.Getenv("HOME"); home != "" {
		return home, nil
	}
	var stdout bytes.Buffer
	cmd := exec.Command("sh", "-c", "eval echo ~$USER")
	cmd.Stdout = &stdout
	if err := cmd.Run(); err != nil {
		return "", err
	}
	result := strings.TrimSpace(stdout.String())
	if result == "" {
		return "", errors.New("无法获取当前用户Home目录(unix)")
	}
	return result, nil
}

func homeWindows() (string, error) {
	drive := os.Getenv("HOMEDRIVE")
	path := os.Getenv("HOMEPATH")
	home := drive + path
	if drive == "" || path == "" {
		home = os.Getenv("USERPROFILE")
	}
	if home == "" {
		return "", errors.New("无法获取当前用户Home目录(windows)")
	}
	return home, nil
}

// 检查日志配置
//  返回 1 代表检查OK, error为nil;
//  返回 2 代表检查OK但LogFileDir为空，error为nil;
//  返回 -9 代表检查失败，error非nil;
func CheckConfig(logConfig *Config) (int, error) {
	if logConfig == nil {
		return CONFIG_CHECK_RESULT_NG, fmt.Errorf("日志配置不可为空")
	}
	if logConfig.LogFileNamePrefix == "" {
		return CONFIG_CHECK_RESULT_NG, fmt.Errorf("日志文件名前缀不可为空")
	}
	if logConfig.LogFileMaxSizeM <= 0 {
		return CONFIG_CHECK_RESULT_NG, fmt.Errorf("日志文件Size上限必须大于0")
	}
	if logConfig.LogLevelGlobal < LOG_LEVEL_DEBUG || logConfig.LogLevelGlobal >= log_level_max {
		return CONFIG_CHECK_RESULT_NG, fmt.Errorf("全局日志级别不能超出有效范围")
	}
	// LogFileDir在本地模式下可以为空，服务器模式下不可为空
	if logConfig.LogMod == LOG_MODE_SERVER && logConfig.LogFileDir == "" {
		return CONFIG_CHECK_RESULT_NG, fmt.Errorf("服务器模式下日志目录不可为空")
	}
	if logConfig.LogMod == LOG_MODE_LOCAL && logConfig.LogFileDir == "" {
		return CONFIG_CHECK_RESULT_NOFILEDIR, nil
	}
	return CONFIG_CHECK_RESULT_OK, nil
}
