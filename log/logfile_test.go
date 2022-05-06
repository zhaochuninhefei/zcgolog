package log

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"testing"
)

func TestFirstLog(t *testing.T) {
	logConfig := DefaultLogConfig()
	logConfig.LogFileDir = "testdata/firstlog/"
	logFileName, _, err := GetLogFilePath(logConfig)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Printf("logFileName: %s\n", logFileName)
	os.Remove(logFileName)
}

func TestLastLog(t *testing.T) {
	ClearDir("testdata/lastlog")
	ymdToday := getYMDToday()
	log1 := "testdata/lastlog/zcgolog_" + ymdToday + "_00001.log"
	for i := 0; i < 100; i++ {
		writeTestLog(log1, "测试消息测试消息测试消息测试消息测试消息测试消息测试消息测试消息测试消息测试消息测试消息"+strconv.Itoa(i+1))
	}
	log2 := "testdata/lastlog/zcgolog_" + ymdToday + "_00002.log"
	for i := 0; i < 100; i++ {
		writeTestLog(log2, "测试消息测试消息测试消息测试消息测试消息测试消息测试消息测试消息测试消息测试消息测试消息"+strconv.Itoa(i+1))
	}
	logLast := "testdata/lastlog/zcgolog_" + ymdToday + "_00123.log"
	for i := 0; i < 10000; i++ {
		writeTestLog(logLast, "测试消息测试消息测试消息测试消息测试消息测试消息测试消息测试消息测试消息测试消息测试消息"+strconv.Itoa(i+1))
	}
	fileState, _ := os.Stat(logLast)
	fmt.Printf("logLast文件大小: %d\n", fileState.Size())

	logConfig := DefaultLogConfig()
	logConfig.LogFileDir = "testdata/lastlog/"
	logConfig.LogFileMaxSizeM = 1
	logFileName, _, err := GetLogFilePath(logConfig)
	if err != nil {
		t.Fatal(err)
	}
	writeTestLog(logFileName, "测试消息测试消息测试消息测试消息测试消息测试消息测试消息测试消息测试消息测试消息测试消息")
	fmt.Printf("logFileName: %s\n", logFileName)
	ClearDir("testdata/lastlog")
}

func TestGetYMDToday(t *testing.T) {
	ymd := getYMDToday()
	fmt.Println(ymd)
}

func writeTestLog(logFilePath string, msg string) {
	logFile, err := os.OpenFile(logFilePath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatal(err)
	}
	defer logFile.Close()
	log.SetOutput(logFile)
	log.Printf(msg + "\n")
}
