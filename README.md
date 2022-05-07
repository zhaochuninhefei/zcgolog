zcgolog
==========

# 介绍
一个简单的golang日志框架，支持服务器模式和本地模式。其中服务器日志提供了以下功能:
- 在线修改具体函数的日志级别(仅在服务器模式下支持)
- 日志异步输出(仅在服务器模式下支持)
- 日志滚动，按天滚动，当天文件按配置的size滚动(仅在服务器模式下支持)
- 日志同时输出到文件与控制台
- 日志输出代码文件路径，代码行数，以及调用方函数包路径信息

本地模式与golang自己的`log`包基本相同，不具备日志级别在线修改、异步输出、文件滚动功能。

# 使用
zcgolog的使用很简单，直接依赖即可使用，默认使用本地模式，如果要使用服务器模式，只需要在代码中添加zcgolog的配置与初始化即可。

依赖zcgolog，在相关工程的`go.mod`文件中添加:
```
gitee.com/zhaochuninhefei/zcgolog v0.0.3
```

并在对应的代码中使用:
```
import "gitee.com/zhaochuninhefei/zcgolog/log"
...

func test() {
    ...
    log.Debug("这是一条测试消息")
}
```

然后执行`go mod tidy`即可。
> 如果无法下载`gitee.com/zhaochuninhefei/zcgolog`，请将`gitee.com/zhaochuninhefei/zcgolog`设置为go的私有仓库，允许直接下载即可:
```sh
go env -w GOPRIVATE=gitee.com/zhaochuninhefei
```

## 服务器模式
如果需要在一个服务中使用zcgolog，并使用其日志级别修改等功能，那么就需要在服务启动时配置并初始化zcgolog，代码示例如下:

```go
// 该函数在服务启动主函数如main函数中调用即可
func initZcgolog() {
	zcgologConf := &log.Config{
        // 指定日志目录
		LogFileDir:        "/logs",
        // 指定日志文件名前缀
		LogFileNamePrefix: "fabric-ca-server-zcgolog",
        // 指定全局日志级别
		LogLevelGlobal:    log.LOG_LEVEL_INFO,
        // 指定日志模式为服务器模式
		LogMod:            log.LOG_MODE_SERVER,
	}
    // 初始化log
	log.InitLogger(zcgologConf)
}
```

其他相关配置的默认值参考`配置默认值`一节。

服务器模式下日志输出格式:
```
写入文件时间 日志级别 日志输出请求时间 代码位置(输出该条日志的代码文件及行数) 函数包路径 日志内容  
```
> 因为是异步输出，所以有两个时间戳。前者是实际写入日志文件的时间，后者是调用方请求写日志的时间。

示例如下:
```
2022/05/07 16:56:39 [   DEBUG] 时间:2022-05-07 16:56:39 代码:/home/zhaochun/work/sources/gitee.com/zhaochuninhefei/zcgolog/log/log_test.go 56 函数:gitee.com/zhaochuninhefei/zcgolog/log.writeLog 测试日志
```


### 服务器模式下在线修改指定函数的日志级别
zcgolog在服务器模式下提供了在线修改日志级别的httpAPI，无需重启服务。以`curl`为例，使用方法如下:

```sh
curl "http://localhost:9300/zcgolog/api/level/ctl?logger=gitee.com/zhaochuninhefei/zcgolog/log.writeLog&level=1"
```

说明:
- host与port : 根据配置确定，默认是`localhost:9300`
- uri : `/zcgolog/api/level/ctl`
- URL参数 : logger和level。logger是调整目标，对应具体函数的完整包名路径，如: `gitee.com/zhaochuninhefei/zcgolog/log.writeLog`；level是调整后的日志级别，支持从1到6，分别是DEBUG,INFO,WARNNING,ERROR,CRITICAL,FATAL。

修改成功后会返回消息:`操作成功`。


## 本地模式
本地模式无需额外配置，当然也支持自定义配置，方法与服务器模式一样，注意`LogMod`采用默认值，或配置为`log.LOG_MODE_LOCAL`。

本地模式下日志输出格式:
```
写入文件时间 日志级别 代码位置(输出该条日志的代码文件及行数) 函数包路径 日志内容  
```

示例如下:
```
2022/05/07 16:57:31 [   DEBUG] 代码:/home/zhaochun/work/sources/gitee.com/zhaochuninhefei/zcgolog/log/log_test.go 82 函数:gitee.com/zhaochuninhefei/zcgolog/log.TestLocalLog 测试日志
```

## 配置默认值
各个配置的默认值如下:
- LogFileDir : 当前用户Home目录/zcgologs，即 `~/zcgologs`
- LogFileNamePrefix : `zcgolog`，日志文件命名约定: `[LogFileNamePrefix]_[年月日]_[%05d].log`，例如: `zcgolog_20220507_00001.log`
- LogFileMaxSizeM : `2`，单个日志文件最大Size，在服务器模式下，日志文件以天为单位滚动，当天日志文件到达上限时再次滚动，文件名最后的序号+1。每天最多允许滚动99999个日志文件。本地模式不支持日志滚动。
- LogLevelGlobal : `LOG_LEVEL_DEBUG`,int类型，值为1。目前支持的日志级别:LOG_LEVEL_DEBUG,LOG_LEVEL_INFO,LOG_LEVEL_WARNING,LOG_LEVEL_ERROR,LOG_LEVEL_CRITICAL,LOG_LEVEL_FATAL,对应的数值从1到6。
- LogLineFormat : "%level %pushTime %file %line %callFunc %msg"，目前日志格式固定，该配置暂时没有使用。
- LogMod : `LOG_MODE_LOCAL`,int类型，值为1,目前支持 LOG_MODE_LOCAL:1 与 LOG_MODE_SERVER:2。
- LogChannelCap : `4096`,int类型，日志缓冲通道的容量，可以根据实际情况调整。仅在服务器模式下支持。
- LogChnOverPolicy : `LOG_CHN_OVER_POLICY_DISCARD`,int类型，值为1。日志缓冲通道已满时的日志处理策略，默认策略是丢弃该条日志，另一个策略是`LOG_CHN_OVER_POLICY_BLOCK`，阻塞等待。两种策略都不是很理想，一般还是调大LogChannelCap确保通道不会被打满。仅在服务器模式下支持。
- LogLevelCtlHost : `localhost`，日志级别调整监听服务的Host，一般不用调整。仅在服务器模式下支持。
- LogLevelCtlPort ： `9300`，日志级别调整监听服务的端口，可根据实际情况调整。仅在服务器模式下支持。

# 其他说明
底层写日志时，直接使用的golang自己的`log`包，因此zcgolog的配置会影响程序中其他使用golang的log包的日志输出，包括其输出目标会被改为同时输出到控制台和zcgolog配置的日志文件，以及其前缀时间戳格式。

