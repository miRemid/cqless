# CQLESS
针对CQHTTP开发的Serverless平台

**目前还在开发阶段初期！现已支持基本的函数调用操作**

# 安装

本项目所需要的依赖有
- Docker
- CNI

## 从Release安装
在Github Release中直接下载安装即可（仅支持Linux）

## 从源码安装
首先克隆本项目至任意路径如：`$HOME/cqless`
```shell
cd $HOME
git clone https://github.com/miRemid/cqless && cd cqless
```

检查并安装所需依赖，需要安装Python3

```shell
chmod +x ./scripts/check_dep.py && ./scripts/check_dep.py
```

该脚本目前会检查系统中是否存在`go`和`docker`，关于`CNI`的安装请参考[官方文档](https://github.com/containernetworking/cni)

检查通过或安装完毕后，编译项目并安装至环境变量中

```
make build
```

编译好的可执行文件将位于`build/bin/cqless`

# 使用

## Gateway网关
Gateway网关是cqless关于函数操作的核心枢纽，也是和CQHTTP对应接口对接的中心

#### 生成配置文件

首先需要生成cqless所需的配置文件
```shell
cqless gateway init
```
该命令会在路径`$HOME/.local/share`中创建cqless所需的配置文件项目，完整的配置文件路径为`$HOME/.local/share/cqless/cqless.yaml`

请更改配置文件中对应的CNI配置选项，设置好CNI的可执行文件路径已经后续需要保存的文件路径，默认配置为`$HOME/.local/share/cqless/cni`

#### 启动命令
配置完全后，启动cqless
```shell
cqless gateway up
```

## 函数操作
cqless提供了CLI工具用于对函数进行操作，同时支持配置文件和命令行参数输入，推荐使用配置文件更好管理

在这我们提供了一个简单的函数镜像`kamir3mid/helloworld`，一个简单的函数配置文件如下：
```json
// examples/helloworld/helloworld.json
{
    "name": "helloworld",
    "image": "kamir3mid/helloworld:latest",

}
```
> 需要注意的是，目前所有函数对应的端口必须为8080

#### 创建

```shell
cqless func deploy -h
部署

Usage:
  cqless func deploy [flags]

Flags:
  -h, --help           help for deploy
  -i, --image string   容器镜像名称
  -n, --name string    函数名称

Global Flags:
  -c, --config string      函数部署配置文件路径，默认为空
  -g, --gateway string     网关地址，默认127.0.0.1 (default "127.0.0.1")
      --namespace string   函数所在命名空间(Docker无需关心) (default "cqless-fn-ns")
  -p, --port int           网关端口，默认8888 (default 8888)
  -t, --timeout int        执行超时时间，默认30s (default 30)
```
我们需要创建一个`helloworld`的函数，只需如下指令
```shell
cqless func deploy -c examples/helloworld/helloworld.json
```

#### 调用

```shell
cqless func invoke -h
调用一个函数接口

Usage:
  cqless func invoke [flags]

Flags:
      --fn string   需要调用的函数名称
  -h, --help        help for invoke

Global Flags:
  -c, --config string      函数部署配置文件路径，默认为空
  -g, --gateway string     网关地址，默认127.0.0.1 (default "127.0.0.1")
      --namespace string   函数所在命名空间(Docker无需关心) (default "cqless-fn-ns")
  -p, --port int           网关端口，默认8888 (default 8888)
  -t, --timeout int        执行超时时间，默认30s (default 30)
```
我们需要调用`helloworld`函数，只需如下指令
```shell
cqless func invoke -c examples/helloworld/helloworld.json
# output
{"status":0,"message":"","data":"hello!"}
```
#### 检查
```shell
cqless func inspect -h                                   
检查

Usage:
  cqless func inspect [flags]

Flags:
      --fn string   需要检查的函数名称
  -h, --help        help for inspect

Global Flags:
  -c, --config string      函数部署配置文件路径，默认为空
  -g, --gateway string     网关地址，默认127.0.0.1 (default "127.0.0.1")
      --namespace string   函数所在命名空间(Docker无需关心) (default "cqless-fn-ns")
  -p, --port int           网关端口，默认8888 (default 8888)
  -t, --timeout int        执行超时时间，默认30s (default 30)
```
```shell
cqless func inspect # 获取所有函数信息
cqless func inspect -c examples/helloworld/helloworld.json # 获取对应配置的函数信息
+------------+----------------------+------------------------------------------------------------------+------------+---------+
| NAME       | FULL NAME            | ID                                                               | IP ADDRESS | STATUS  |
+------------+----------------------+------------------------------------------------------------------+------------+---------+
| helloworld | /helloworld-ffb67948 | 87f8e766cb23292f963b0eb8ac222493ec3c9c818c7f1342e93759c8be38e51c | 10.72.0.48 | running |
+------------+----------------------+------------------------------------------------------------------+------------+---------+
```

#### 删除
```shell
cqless func remove -h                                     
删除目标函数

Usage:
  cqless func remove [flags]

Flags:
      --fn string   需要删除的函数名称
  -h, --help        help for remove

Global Flags:
  -c, --config string      函数部署配置文件路径，默认为空
  -g, --gateway string     网关地址，默认127.0.0.1 (default "127.0.0.1")
      --namespace string   函数所在命名空间(Docker无需关心) (default "cqless-fn-ns")
  -p, --port int           网关端口，默认8888 (default 8888)
  -t, --timeout int        执行超时时间，默认30s (default 30)
```
我们需要删除`helloworld`函数，只需如下指令
```shell
cqless func remove -c examples/helloworld/helloworld.json
```

## License

[Apache-2.0](https://www.apache.org/licenses/LICENSE-2.0)

