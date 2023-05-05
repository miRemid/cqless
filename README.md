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

请参考[Wiki页面](https://github.com/miRemid/cqless/wiki)

## License

[Apache-2.0](https://www.apache.org/licenses/LICENSE-2.0)

