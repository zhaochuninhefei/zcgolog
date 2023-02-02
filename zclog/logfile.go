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
zclog/logfile.go 日志文件相关处理
*/

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"strconv"
	"strings"
	"time"
)

//goland:noinspection GoSnakeCaseUsage
const (
	OS_OUT_STDOUT = "OS.STDOUT"
)

// GetLogFilePathAndYMDToday 获取日志文件路径和当天年月日。
//  不存在当天对应日志文件时，创建新的日志文件；
//  存在当天对应日志文件时，获取最新的日志文件；
//  最新日志文件大小超过配置的日志文件大小上限时，创建新的日志文件。
//  每天的日志文件数量不能超过99999，否则会报错。
//  本地模式下，如果LogFileDir为空，则返回 ("OS.STDOUT", 当天日期, nil)表示只能输出到控制台。
func GetLogFilePathAndYMDToday(logConfig *Config) (string, string, error) {
	// 检查日志配置
	res, err := CheckConfig(logConfig)
	if err != nil {
		return "", "", err
	}
	// 检查通过但没有LogFileDir，此时只能输出到控制台
	if res == CONFIG_CHECK_RESULT_NOFILEDIR {
		return OS_OUT_STDOUT, getYMDToday(), nil
	}
	// 防止日志文件目录尚未创建
	_ = os.MkdirAll(logConfig.LogFileDir, os.ModePerm)
	// 读取日志目录下所有文件
	files, err := ioutil.ReadDir(logConfig.LogFileDir)
	if err != nil {
		return "", "", err
	}
	// 获取当天年月日，并生成日志文件前缀
	ymdToday := getYMDToday()
	prefix := logConfig.LogFileNamePrefix + "_" + ymdToday + "_"
	var fileNames []string
	// 过滤当天日志
	for _, f := range files {
		fileName := f.Name()
		if strings.HasPrefix(fileName, prefix) && strings.HasSuffix(fileName, ".log") {
			fileNames = append(fileNames, fileName)
		}
	}
	// 查找最新日志文件
	maxNumber := 0
	lastFileName := ""
	for _, filename := range fileNames {
		suffixIndex := strings.LastIndex(filename, ".log")
		tmp := filename[0:suffixIndex]
		arrs := strings.Split(tmp, "_")
		numberStr := arrs[len(arrs)-1]
		number, err := strconv.Atoi(numberStr)
		if err != nil {
			continue
		}
		if number > maxNumber {
			maxNumber = number
			lastFileName = filename
		}
	}
	if maxNumber > 0 {
		lastFilePath := path.Join(logConfig.LogFileDir, lastFileName)
		// 读取最新文件状态
		lastFileState, err := os.Stat(lastFilePath)
		if err != nil {
			return "", ymdToday, err
		}
		// 检查文件大小
		if lastFileState.Size() < int64(logConfig.LogFileMaxSizeM)*1024*1024 {
			return lastFilePath, ymdToday, nil
		}
	}
	// 当天没有日志文件，或最新日志文件大小已超出上限时，需要创建新的日志文件
	numberToday, err := nextNumberToday(maxNumber)
	if err != nil {
		return "", ymdToday, err
	}
	targetFileName := prefix + numberToday + ".log"
	targetFilePath := path.Join(logConfig.LogFileDir, targetFileName)
	err = ioutil.WriteFile(targetFilePath, []byte{}, 0644)
	if err != nil {
		return "", ymdToday, err
	}
	return targetFilePath, ymdToday, nil
}

// 获取当天年月日，格式:yyyyMMdd
func getYMDToday() string {
	now := time.Now()
	return fmt.Sprintf("%d%02d%02d", now.Year(), now.Month(), now.Day())
}

// 当天日志序列号+1
//  不能超过99999，否则报错
func nextNumberToday(maxNumber int) (string, error) {
	number := maxNumber + 1
	if number > 99999 {
		return "", fmt.Errorf("当天日志文件数量超过上限:99999")
	}
	return fmt.Sprintf("%05d", number), nil
}

func ClearDir(dirPath string) error {
	files, err := ioutil.ReadDir(dirPath)
	if err != nil {
		return err
	}
	for _, file := range files {
		_ = os.RemoveAll(path.Join(dirPath, file.Name()))
	}
	return nil
}
