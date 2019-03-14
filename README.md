# shino

## 本地开发

```sh
NAME:
   shino - a command line tool for Kuu

USAGE:
   shino [global options] command [command options] [arguments...]

VERSION:
   0.0.0

COMMANDS:
     help, h  Shows a list of commands or help for one command

GLOBAL OPTIONS:
   --base value     base project [$BASE]
   --install value  install command (default: "npm install") [$INSTALL]
   --start value    start command (default: "npm start") [$START]
   --sync value     sync dir (default: "sync") [$SYNC]
   --help, -h       show help
   --version, -v    print the version
```

命令行配置项：

- `base` - 基础项目地址，必填参数
- `install` - 项目安装命令，默认值“npm install”
- `start` - 项目启动命令，默认值“npm start”
- `sync` - 监听同步的代码目录，默认值“src”

同样也支持在`kuu.json`中配置：

```json
{
  "base": "https://github.com/fho/fho-admin.git",
  "install": "npm install",
  "start": "npm start",
  "sync": "src"
}
```

## Drone 插件

```yaml
pipeline:
  build:
    image: yinfxs/shino
    base: https://github.com/fho/fho-admin.git
```

支持的配置项：

- **base** - 项目安装命令，默认值为“npm install”
- **sync** - 本地代码目录，默认值为“src”
- **install** - 项目安装命令，默认值为“npm install”
- **build** - 项目构建命令，默认值为“npm run build”
