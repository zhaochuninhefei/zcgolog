zcgolog
==========

# 介绍
一个简单的golang日志框架，支持服务器模式和本地模式，底层仍然使用golang的`log`包，在其基础上提供了以下功能:
- 在线修改具体函数的日志级别(仅在服务器模式下支持)
- 日志异步输出(仅在服务器模式下支持)
- 日志滚动，按天滚动，当天文件按配置的size滚动(仅在服务器模式下支持)
- 支持日志同时输出到文件与控制台
- 日志格式输出代码文件路径，代码行数，以及调用方函数包路径信息

本地模式与golang的`log`包基本相同，不具备日志级别在线修改、异步输出、文件滚动功能。

# 使用
zcgolog的使用很简单，直接依赖即可使用，默认使用本地模式，如果要使用服务器模式，只需要在代码中添加zcgolog的配置与初始化即可。

在对应的代码中使用:
```
import "gitee.com/zhaochuninhefei/zcgolog/zclog"
...

func test() {
    ...
    zclog.Debug("这是一条测试消息")
}
```

然后在相关工程的`go.mod`文件中添加:
```
gitee.com/zhaochuninhefei/zcgolog latest
```

然后在`go.mod`文件目录下执行`go mod tidy`即可。
> `go mod tidy`命令遇到版本号`latest`时会自动下载最新版本。
> 
> 如果无法下载`gitee.com/zhaochuninhefei/zcgolog`，请将`gitee.com/zhaochuninhefei/zcgolog`设置为go的私有仓库，允许直接下载即可:
```sh
go env -w GOPRIVATE=gitee.com/zhaochuninhefei/zcgolog
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
		LogFileNamePrefix: "xxxxxxx",
        // 指定全局日志级别
		LogLevelGlobal:    log.LOG_LEVEL_INFO,
        // 指定日志模式为服务器模式
		LogMod:            log.LOG_MODE_SERVER,
	}
    // 初始化log
	log.InitLogger(zcgologConf)
}
```

其他相关配置的默认值参考`配置及其默认值`一节。

服务器模式下日志输出格式:
```
[写入文件时间] [日志级别] [日志输出请求时间] [代码位置(输出该条日志的代码文件及行数)] [函数包路径] [日志内容]  
```
> 因为是异步输出，所以有两个时间戳。前者是实际写入日志文件的时间，后者是调用方请求写日志的时间。

示例如下:
```
2022/05/07 16:56:39 [DEBUG] 时间:2022-05-07 16:56:39 代码:/home/zhaochun/work/sources/gitee.com/zhaochuninhefei/zcgolog/log/log_test.go 56 函数:gitee.com/zhaochuninhefei/zcgolog/log.writeLog 测试日志
```


### 服务器模式下在线修改指定函数的日志级别
zcgolog在服务器模式下提供了在线修改日志级别的httpAPI，无需重启服务。以`curl`为例，使用方法如下:

```sh
# 修改全局日志级别
# 如果level传入[1~6]以外的值，则全局日志级别恢复为启动时的配置
# 修改成功返回 "操作成功"
curl "http://localhost:9300/zcgolog/api/level/global?level=1"

# 修改指定函数的日志级别
# 如果level传入[1~6]以外的值，则作为0处理，该函数的日志级别将采用全局日志级别
# 修改成功返回 "操作成功"
curl "http://localhost:9300/zcgolog/api/level/ctl?logger=gitee.com/zhaochuninhefei/zcgolog/log.writeLog&level=1"

# 查看全局日志级别
# 返回值是日志级别对应的字符串，如 "debug","info","warning","error","panic","fatal"
curl "http://localhost:9300/zcgolog/api/level/query"

# 查看指定函数的日志级别
# 返回值是日志级别对应的字符串，如 "debug","info","warning","error","panic","fatal"
curl "http://localhost:9300/zcgolog/api/level/query?logger=gitee.com/zhaochuninhefei/zcgolog/log.writeLog"
```

zcgolog的在线日志级别调整与查看的HttpAPI列表:

| uri | URL参数 | 用途 |
| --- | --- | --- |
| /zcgolog/api/level/ctl | logger和level。logger是调整目标，对应具体函数的完整包名路径，如: `gitee.com/zhaochuninhefei/zcgolog/log.writeLog`；level是调整后的日志级别，支持从1到6，分别是DEBUG,INFO,WARNNING,ERROR,PANIC,FATAL。 | 用于在线修改目标函数的日志级别。 |
| /zcgolog/api/level/global | level,指定全局日志级别 | 用于在线修改全局日志级别。 |
| /zcgolog/api/level/query | logger,指定需要查看日志级别的目标函数,不传参数代表查看全局日志级别。 | 用于查看全局或指定函数的日志级别。 |

> HttpAPI的host与port根据配置确定，默认是`:9300`，具体配置参见后续的`配置及其默认值`一节。

## 本地模式
本地模式无需额外配置，当然也支持自定义配置，方法与服务器模式一样，注意`LogMod`采用默认值，或配置为`log.LOG_MODE_LOCAL`。
> 本地模式默认只输出到控制台。输出日志文件需要显式配置，参考后续的`配置及其默认值`中的相关说明。

本地模式下日志输出格式:
```
[写入文件时间] [日志级别] [代码位置(输出该条日志的代码文件及行数)] [函数包路径] [日志内容]  
```

示例如下:
```
2022/05/07 16:57:31 [DEBUG] 代码:/home/zhaochun/work/sources/gitee.com/zhaochuninhefei/zcgolog/log/log_test.go 82 函数:gitee.com/zhaochuninhefei/zcgolog/log.TestLocalLog 测试日志
```

## 配置及其默认值
各个配置的说明以及默认值如下:
- LogMod : 日志模式，默认值`LOG_MODE_LOCAL`,int类型，值为1,目前支持 本地模式(LOG_MODE_LOCAL:1) 与 服务器模式(LOG_MODE_SERVER:2)。
- LogFileDir : 日志文件目录。服务器模式下必须显式配置一个非空目录，没有默认值。本地模式下默认为空，此时日志只输出到控制台，显式配置则同时输出到日志文件与控制台。
- LogFileNamePrefix : 日志文件名前缀，默认值`zcgolog`。完整的日志文件命名约定: `[LogFileNamePrefix]_[年月日]_[%05d].log`，例如: `zcgolog_20220507_00001.log`，只在LogFileDir非空时有效。
- LogForbidStdout :  是否禁止输出到控制台，默认值`false`。
- LogLevelGlobal : 全局日志级别，默认值`LOG_LEVEL_INFO`,int类型，值为2。目前支持的日志级别:LOG_LEVEL_DEBUG,LOG_LEVEL_INFO,LOG_LEVEL_WARNING,LOG_LEVEL_ERROR,LOG_LEVEL_PANIC,LOG_LEVEL_FATAL,对应的数值从1到6。具体每个日志级别的说明，参考后续的`支持的日志级别`。
- LogLineFormat : 日志格式，目前日志格式固定，该配置暂时没有使用。
- LogFileMaxSizeM : 单个日志文件Size上限(单位:M)，默认值`2`。在服务器模式下，日志文件以天为单位滚动，当天日志文件到达上限时再次滚动，文件名最后的序号+1。每天最多允许滚动99999个日志文件。仅在服务器模式下支持。
- LogChannelCap : 日志缓冲通道的容量，默认值`4096`,int类型，可以根据实际情况调整，尤其日志输出并发较高时请将该值调大。仅在服务器模式下支持。
- LogChnOverPolicy : 日志缓冲通道已满时的日志处理策略，默认值`LOG_CHN_OVER_POLICY_DISCARD`,int类型，值为1。默认策略是丢弃该条日志(但会输出到控制台)，另一个策略是`LOG_CHN_OVER_POLICY_BLOCK`，阻塞等待。两种策略都不是很理想，一般还是调大LogChannelCap确保通道不会被打满。仅在服务器模式下支持。
- LogLevelCtlHost : 日志级别调整监听服务的Host，默认为空，即监听程序主机的各个IP。可根据实际需要调整，比如配置为`localhost`时将只能在程序主机本地访问，其他网络地址无法访问到该服务。仅在服务器模式下支持。
- LogLevelCtlPort ： 日志级别调整监听服务的端口，默认值`9300`。可根据实际情况调整。仅在服务器模式下支持。

## 支持的日志级别
```go
	// debug 调试日志，生产环境通常关闭
	LOG_LEVEL_DEBUG = 1
	// info 重要信息日志，用于提示程序过程中的一些重要信息，慎用，避免过多的INFO日志
	LOG_LEVEL_INFO = 2
	// warning 警告日志，用于警告用户可能会发生问题
	LOG_LEVEL_WARNING = 3
	// error 一般错误日志，一般用于提示业务错误，程序通常不会因为这样的错误终止
	LOG_LEVEL_ERROR = 4
	// panic 异常错误日志，一般用于预期外的错误，程序的当前Goroutine会终止并输出堆栈信息
	LOG_LEVEL_PANIC = 5
	// fatal 致命错误日志，程序会马上终止
	LOG_LEVEL_FATAL = 6
```

## 日志输出调用接口

| 函数名 | 日志级别 | 参数列表 | 说明 |
| --- | --- | --- | --- |
| Print | LOG_LEVEL_DEBUG | v ...interface{} | Print在zcgolog中处理为与Debug一致 |
| Printf | LOG_LEVEL_DEBUG | msg string, params ...interface{} | Printf在zcgolog中处理为与Debugf一致 |
| Println | LOG_LEVEL_DEBUG | v ...interface{} | Println在zcgolog中处理为与Debugln一致 |
| Debug | LOG_LEVEL_DEBUG | v ...interface{} | 参数直接拼接，末尾换行 |
| Debugf | LOG_LEVEL_DEBUG | msg string, params ...interface{} | 参数按照msg中的format定义格式化拼接，末尾换行 |
| Debugln | LOG_LEVEL_DEBUG | v ...interface{} | 参数直接拼接，末尾换行 |
| Info | LOG_LEVEL_INFO | v ...interface{} | 参数直接拼接，末尾换行 |
| Infof | LOG_LEVEL_INFO | msg string, params ...interface{} | 参数按照msg中的format定义格式化拼接，末尾换行 |
| Infoln | LOG_LEVEL_INFO | v ...interface{} | 参数直接拼接，末尾换行 |
| Warn | LOG_LEVEL_WARNING | v ...interface{} | 参数直接拼接，末尾换行 |
| Warnf | LOG_LEVEL_WARNING | msg string, params ...interface{} | 参数按照msg中的format定义格式化拼接，末尾换行 |
| Warnln | LOG_LEVEL_WARNING | v ...interface{} | 参数直接拼接，末尾换行 |
| Error | LOG_LEVEL_ERROR | v ...interface{} | 参数直接拼接，末尾换行 |
| Errorf | LOG_LEVEL_ERROR | msg string, params ...interface{} | 参数按照msg中的format定义格式化拼接，末尾换行 |
| Errorln | LOG_LEVEL_ERROR | v ...interface{} | 参数直接拼接，末尾换行 |
| Panic | LOG_LEVEL_PANIC | v ...interface{} | 参数直接拼接，并输出堆栈信息，无视服务器模式直接输出日志并终止当前goroutine |
| Panicf | LOG_LEVEL_PANIC | msg string, params ...interface{} | 参数按照msg中的format定义格式化拼接，无视服务器模式直接输出日志并终止当前goroutine |
| Panicln | LOG_LEVEL_PANIC | v ...interface{} | 参数直接拼接，并输出堆栈信息，无视服务器模式直接输出日志并终止当前goroutine |
| Fatal | LOG_LEVEL_FATAL | v ...interface{} | 参数直接拼接，并输出堆栈信息，无视服务器模式直接输出日志并终止程序 |
| Fatalf | LOG_LEVEL_FATAL | msg string, params ...interface{} | 参数按照msg中的format定义格式化拼接，无视服务器模式直接输出日志并终止程序 |
| Fatalln | LOG_LEVEL_FATAL | v ...interface{} | 参数直接拼接，并输出堆栈信息，无视服务器模式直接输出日志并终止程序 |


# zcgolog性能基准测试
针对zcgolog的服务器模式，本地模式，以及golang原生`log`包做了性能基准测试。代码:`benchtest/log_benchmark_test.go`

zcgolog服务器模式的性能表现最好，相比直接使用golang原生`log`包，性能约提升了约1倍。
> 因为服务器模式采用了异步输出。

zcgolog本地模式性能表现最差，相比直接使用golang原生`log`包，性能下降了约1倍。
> 因为本地模式与golang原生`log`包一样是同步输出日志，同时每次输出日志时有一些额外操作，比如获取runtime代码文件位置以及包路径等。

具体数据如下:

```
goos: linux
goarch: amd64
pkg: gitee.com/zhaochuninhefei/zcgolog/benchtest
cpu: 12th Gen Intel(R) Core(TM) i7-12700H
BenchmarkLogServer-20    	 1960144	       587.0 ns/op	     441 B/op	       6 allocs/op
BenchmarkLogLocal-20     	  566115	      2121 ns/op	     896 B/op	       9 allocs/op
BenchmarkLogGolang-20    	 1000000	      1090 ns/op	      39 B/op	       1 allocs/op
```

> BenchmarkLogServer:服务器模式; BenchmarkLogLocal:本地模式; BenchmarkLogGolang:直接使用golang原生log包。

