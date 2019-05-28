### 基本功能

自动提交问卷星问卷，目前支持：

> - 指定答案模板，程序将以`10:1`的比例提交期望答案和非期望答案
> - 可通过接口提供代理ip，使用代理ip提交答案
> - 只支持单选以及提交时不需要进行验证码验证的问卷
> - 详细的提交日志，供用户在后期对答案提交进行复盘

### 食用方法

以 `Linux/OSX` 进行说明，Windows系统请自行脑补。

#### 安装golang

如果没有安装golang需要安装golang开发环境，具体安装步骤 => [安装Go](https://github.com/astaxie/build-web-application-with-golang/blob/master/zh/01.1.md)

#### 下载项目

> go get -u github.com/Jungzhang/wjx

#### 编译项目

> cd $GOPATH/src/github.com/Jungzhang/wjx && go build wjx.go

#### 配置答案模板

配置一个答案模板，用来记录期望答案等信息。模板格式为：题号 期望答案 总答案个数。例如：
```
1 AB 2
2 ABCDEF 6
3 ABCD 4
4 ABC 4
5 ACD 5
6 CDE 5   => 表示期望 C、D、E答案的出现，该题共有 ABCDE 5个答案选项
7 BCD 5
8 BCD 5
9 BCDE 5
10 BCD 5
```

#### 启动

推荐启动时将输出日志重定向至文件中，方便后续分析。

> ./wjx 提交数 试卷id 答案模板 代理接口地址(可选)

例如，需要为id为 12345678 的问卷自动提交1000份调查结果，答案模板文件为answer.txt，且使用`http://127.0.0.1:8080/api/v1/proxy/ip`接口随机获取代理ip进行答案提交。则启动命令为：

> ./wjx 1000 12345678 ./answer.txt http://127.0.0.1:8080/api/v1/proxy/ip >> submit.log 2>&1

重点说明：**当前问卷星对ip提交有限制，如果同一个ip提交次数过多，需要输入验证码，且同一个ip刷票过多对答案的仿真度也不够高，推荐启用ip代理功能。**

#### IP代理配置

当使用ip代理功能时需要提供一个ip代理池接口，通过该接口可以随机获取一个代理ip`（如需ip代理池服务可邮件博主或提issue留下联系方式，博主免费提供接口服务）`，要求该接口中必须包含ip字段，如：

```
要求代理ip接口返回至少包含如下字段

{
    "ip": "192.168.1.109:8080"
}
```

#### 日志分离

分离提交成功的日志

> grep "提交成功" submit.log >> success_linux.log

统计期望结果的提交总次数

> grep "使用期望" success_linux.log | wc -l

统计非期望结果的提交次数

> grep "使用非期望" success_linux.log | wc -l 

