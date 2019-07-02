# shino

[![Build Status](https://travis-ci.org/kuuland/shino.svg?branch=master)](https://travis-ci.org/kuuland/shino)

```sh
NAME:
   shino - CLI for Kuu

USAGE:
   shino [global options] command [command options] [arguments...]

VERSION:
   0.0.0

COMMANDS:
     up       startup project
     fano     CLI for FanoJS
     help, h  Shows a list of commands or help for one command

GLOBAL OPTIONS:
   --help, -h     show help
   --version, -v  print the version
```

## 本地开发

```sh
NAME:
   shino up - startup project

USAGE:
   shino up [command options] [arguments...]

OPTIONS:
   --base value     base project (default: "https://github.com/kuuland/ui.git") [$BASE]
   --install value  install command (default: "npm install") [$INSTALL]
   --start value    start command (default: "npm start") [$START]
   --sync value     sync dir (default: "/Users/yinfxs/gopath/src/github.com/kuuland/shino") [$SYNC]
```

命令行配置项：

- `base` - 基础项目地址，必填参数
- `install` - 项目安装命令，默认值“npm install”
- `start` - 项目启动命令，默认值“npm start”
- `sync` - 监听同步的代码目录，默认值“src”

同样也支持在`kuu.json`中配置：

```json
{
  "base": "https://github.com/kuuland/ui.git",
  "install": "npm install",
  "start": "npm start"
}
```

## Drone CI插件

```yaml
steps:
- name: prebuild  
  image: yinfxs/shino
  pull: true
```

## Fano代码生成

shino提供了基于**元数据**的代码生成功能

### 表格页面

```sh
NAME:
   shino fano table - generate table pages based on metadata

USAGE:
   shino fano table [command options] [arguments...]

OPTIONS:
   --meta value  metadata url
   --out value   output dir
```

```sh
shino fano table \
    --meta http://localhost:8080/api/meta?json=1 \
    --out out
```