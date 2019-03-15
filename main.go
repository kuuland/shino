package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"os/user"
	"path"
	"path/filepath"
	"strings"
	"syscall"

	"github.com/fatih/color"
	"github.com/fsnotify/fsnotify"
	"github.com/urfave/cli"
)

var (
	base    = ""
	install = "npm install"
	start   = "npm start"
	build   = "npm run build"
	sync    = cwd()

	wsPath       = path.Join(cwd(), ".shino")
	wsBasePath   = path.Join(wsPath, "base")
	wsMergedPath = path.Join(wsPath, "merged")

	version string

	outputFlag = "[SHINO]"
)

func main() {
	if os.Getenv("DRONE_REPO") != "" {
		runDrone()
	} else {
		runLocal()
	}
}

func runDrone() {
	app := cli.NewApp()
	app.Name = "shino"
	app.Usage = "shino plugin"
	app.Version = version
	app.Flags = []cli.Flag{
		//
		// repo args
		//
		cli.StringFlag{
			Name:   "repo.fullname",
			Usage:  "repository full name",
			EnvVar: "DRONE_REPO",
		},
		cli.StringFlag{
			Name:   "repo.owner",
			Usage:  "repository owner",
			EnvVar: "DRONE_REPO_OWNER",
		},
		cli.StringFlag{
			Name:   "repo.name",
			Usage:  "repository name",
			EnvVar: "DRONE_REPO_NAME",
		},
		cli.StringFlag{
			Name:   "repo.link",
			Usage:  "repository link",
			EnvVar: "DRONE_REPO_LINK",
		},
		cli.StringFlag{
			Name:   "repo.avatar",
			Usage:  "repository avatar",
			EnvVar: "DRONE_REPO_AVATAR",
		},
		cli.StringFlag{
			Name:   "repo.branch",
			Usage:  "repository default branch",
			EnvVar: "DRONE_REPO_BRANCH",
		},
		cli.BoolFlag{
			Name:   "repo.private",
			Usage:  "repository is private",
			EnvVar: "DRONE_REPO_PRIVATE",
		},
		cli.BoolFlag{
			Name:   "repo.trusted",
			Usage:  "repository is trusted",
			EnvVar: "DRONE_REPO_TRUSTED",
		},
		//
		// commit args
		//
		cli.StringFlag{
			Name:   "remote.url",
			Usage:  "git remote url",
			EnvVar: "DRONE_REMOTE_URL",
		},
		cli.StringFlag{
			Name:   "commit.sha",
			Usage:  "git commit sha",
			EnvVar: "DRONE_COMMIT_SHA",
		},
		cli.StringFlag{
			Name:   "commit.ref",
			Value:  "refs/heads/master",
			Usage:  "git commit ref",
			EnvVar: "DRONE_COMMIT_REF",
		},
		cli.StringFlag{
			Name:   "commit.branch",
			Value:  "master",
			Usage:  "git commit branch",
			EnvVar: "DRONE_COMMIT_BRANCH",
		},
		cli.StringFlag{
			Name:   "commit.message",
			Usage:  "git commit message",
			EnvVar: "DRONE_COMMIT_MESSAGE",
		},
		cli.StringFlag{
			Name:   "commit.link",
			Usage:  "git commit link",
			EnvVar: "DRONE_COMMIT_LINK",
		},
		cli.StringFlag{
			Name:   "commit.author.name",
			Usage:  "git author name",
			EnvVar: "DRONE_COMMIT_AUTHOR",
		},
		cli.StringFlag{
			Name:   "commit.author.email",
			Usage:  "git author email",
			EnvVar: "DRONE_COMMIT_AUTHOR_EMAIL",
		},
		cli.StringFlag{
			Name:   "commit.author.avatar",
			Usage:  "git author avatar",
			EnvVar: "DRONE_COMMIT_AUTHOR_AVATAR",
		},
		//
		// build args
		//
		cli.StringFlag{
			Name:   "build.event",
			Value:  "push",
			Usage:  "build event",
			EnvVar: "DRONE_BUILD_EVENT",
		},
		cli.IntFlag{
			Name:   "build.number",
			Usage:  "build number",
			EnvVar: "DRONE_BUILD_NUMBER",
		},
		cli.IntFlag{
			Name:   "build.created",
			Usage:  "build created",
			EnvVar: "DRONE_BUILD_CREATED",
		},
		cli.IntFlag{
			Name:   "build.started",
			Usage:  "build started",
			EnvVar: "DRONE_BUILD_STARTED",
		},
		cli.IntFlag{
			Name:   "build.finished",
			Usage:  "build finished",
			EnvVar: "DRONE_BUILD_FINISHED",
		},
		cli.StringFlag{
			Name:   "build.status",
			Usage:  "build status",
			Value:  "success",
			EnvVar: "DRONE_BUILD_STATUS",
		},
		cli.StringFlag{
			Name:   "build.link",
			Usage:  "build link",
			EnvVar: "DRONE_BUILD_LINK",
		},
		cli.StringFlag{
			Name:   "build.deploy",
			Usage:  "build deployment target",
			EnvVar: "DRONE_DEPLOY_TO",
		},
		cli.BoolFlag{
			Name:   "yaml.verified",
			Usage:  "build yaml is verified",
			EnvVar: "DRONE_YAML_VERIFIED",
		},
		cli.BoolFlag{
			Name:   "yaml.signed",
			Usage:  "build yaml is signed",
			EnvVar: "DRONE_YAML_SIGNED",
		},
		//
		// prev build args
		//
		cli.IntFlag{
			Name:   "prev.build.number",
			Usage:  "previous build number",
			EnvVar: "DRONE_PREV_BUILD_NUMBER",
		},
		cli.StringFlag{
			Name:   "prev.build.status",
			Usage:  "previous build status",
			EnvVar: "DRONE_PREV_BUILD_STATUS",
		},
		cli.StringFlag{
			Name:   "prev.commit.sha",
			Usage:  "previous build sha",
			EnvVar: "DRONE_PREV_COMMIT_SHA",
		},
		//
		// config args
		//
		cli.StringFlag{
			Name:   "config.base",
			Usage:  "base project",
			EnvVar: "PLUGIN_BASE",
		},
		cli.StringFlag{
			Name:   "config.install",
			Usage:  "install command",
			EnvVar: "PLUGIN_INSTALL",
		},
		cli.StringFlag{
			Name:   "config.build",
			Usage:  "build command",
			EnvVar: "PLUGIN_BUILD",
		},
		cli.StringFlag{
			Name:   "sync",
			Usage:  "sync dir",
			Value:  "config.sync",
			EnvVar: "PLUGIN_SYNC",
		},
	}
	app.Action = func(c *cli.Context) {
		plugin := Plugin{
			Repo: Repo{
				Owner:   c.String("repo.owner"),
				Name:    c.String("repo.name"),
				Link:    c.String("repo.link"),
				Avatar:  c.String("repo.avatar"),
				Branch:  c.String("repo.branch"),
				Private: c.Bool("repo.private"),
				Trusted: c.Bool("repo.trusted"),
			},
			Build: Build{
				Number:   c.Int("build.number"),
				Event:    c.String("build.event"),
				Status:   c.String("build.status"),
				Deploy:   c.String("build.deploy"),
				Created:  int64(c.Int("build.created")),
				Started:  int64(c.Int("build.started")),
				Finished: int64(c.Int("build.finished")),
				Link:     c.String("build.link"),
			},
			Commit: Commit{
				Remote:  c.String("remote.url"),
				Sha:     c.String("commit.sha"),
				Ref:     c.String("commit.sha"),
				Link:    c.String("commit.link"),
				Branch:  c.String("commit.branch"),
				Message: c.String("commit.message"),
				Author: Author{
					Name:   c.String("commit.author.name"),
					Email:  c.String("commit.author.email"),
					Avatar: c.String("commit.author.avatar"),
				},
			},
			Config: Config{
				Base:    c.String("config.base"),
				Install: c.String("config.install"),
				Build:   c.String("config.build"),
				Sync:    c.String("config.sync"),
			},
		}

		if err := plugin.Exec(); err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}

func runLocal() {
	app := cli.NewApp()
	app.Name = "shino"
	app.Usage = "a command line tool for Kuu"
	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:   "base",
			Usage:  "base project",
			EnvVar: "BASE",
		},
		cli.StringFlag{
			Name:   "install",
			Usage:  "install command",
			Value:  "npm install",
			EnvVar: "INSTALL",
		},
		cli.StringFlag{
			Name:   "start",
			Usage:  "start command",
			Value:  "npm start",
			EnvVar: "START",
		},
		cli.StringFlag{
			Name:   "sync",
			Usage:  "sync dir",
			Value:  cwd(),
			EnvVar: "SYNC",
		},
	}
	app.Action = func(c *cli.Context) {
		baseVal := c.String("base")
		installVal := c.String("install")
		startVal := c.String("start")
		syncVal := c.String("sync")

		if baseVal != "" {
			base = strings.TrimSpace(baseVal)
		}
		if installVal != "" {
			install = strings.TrimSpace(installVal)
		}
		if startVal != "" {
			start = strings.TrimSpace(startVal)
		}
		if syncVal != "" {
			sync = strings.TrimSpace(syncVal)
		}

		setup()
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}

func parseConfigFile() {
	var cfg map[string]string
	cfgFile := path.Join(cwd(), "kuu.json")
	if stat, err := os.Stat(cfgFile); err == nil && !stat.IsDir() {
		if data, err := ioutil.ReadFile(cfgFile); err == nil {
			json.Unmarshal(data, &cfg)
		}
	} else {
		return
	}
	if cfg != nil {
		if v, ok := cfg["base"]; ok {
			base = strings.TrimSpace(v)
		}
		if v, ok := cfg["install"]; ok {
			install = strings.TrimSpace(v)
		}
		if v, ok := cfg["start"]; ok {
			start = strings.TrimSpace(v)
		}
		if v, ok := cfg["sync"]; ok {
			sync = strings.TrimSpace(v)
		}
	}
}

func execCmd(cmd *exec.Cmd) {
	buf := new(bytes.Buffer)
	cmd.Stdout = io.MultiWriter(os.Stdout, buf)
	cmd.Stderr = io.MultiWriter(os.Stderr, buf)
	logArgs(cmd.Args)
	if err := cmd.Run(); err != nil {
		log.Fatal(err)
	}
}

func setup() {
	parseConfigFile()
	ctx, cancel := context.WithCancel(context.Background())
	//创建监听退出chan
	c := make(chan os.Signal)
	//监听指定信号 ctrl+c kill
	signal.Notify(c, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	go func() {
		for s := range c {
			switch s {
			case syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT:
				cancel()
				errorPrint("%s exit %v\n", outputFlag, s)
			default:
				errorPrint("%s other %v\n", outputFlag, s)
			}
		}
	}()
	// 检查是否存在.shino/base目录
	if _, err := os.Stat(wsBasePath); os.IsNotExist(err) {
		ensureDir(wsBasePath)
		// 执行clone命令
		cloneCmd := clone(base, wsBasePath)
		execCmd(cloneCmd)
	}
	// 执行合并：.shino/base + sync = .shino/merged
	if _, err := os.Stat(wsMergedPath); os.IsNotExist(err) {
		execMerge()
		// 执行install命令
		if install != "" {
			installCmd := mergedCmd(ctx, install)
			execCmd(installCmd)
		}
	}
	// 执行start命令
	startCmd := mergedCmd(ctx, start)
	go execCmd(startCmd)
	// 启动监听器
	registerWatcher()
}

func logArgs(args []string) {
	output := outputFlag
	for _, arg := range args {
		output = fmt.Sprintf("%s %s", output, arg)
	}
	successPrint(fmt.Sprintf("%s\n", output))
}

func clone(url, local string) *exec.Cmd {
	return exec.Command(
		"git",
		"clone",
		"--depth=1",
		url,
		local,
	)
}

func mergedCmd(ctx context.Context, execStr string) *exec.Cmd {
	args := strings.Split(execStr, " ")
	cmd := exec.CommandContext(ctx, args[0])
	if len(args) > 1 {
		cmd.Args = args
	}
	cmd.Dir = wsMergedPath
	return cmd
}

func successPrint(format string, a ...interface{}) {
	color.New(color.FgHiGreen, color.Bold).Printf(format, a...)
}

func errorPrint(format string, a ...interface{}) {
	color.New(color.FgHiRed, color.Bold).Printf(format, a...)
}

func execMerge() {
	successPrint("%s merge dirs\n", outputFlag)
	safeDirs := []string{"", ".", "/", "/usr", os.TempDir()}
	if v, err := os.UserCacheDir(); err == nil {
		safeDirs = append(safeDirs, v)
	}
	if usr, err := user.Current(); err == nil {
		safeDirs = append(safeDirs, usr.HomeDir)
	}
	for _, dir := range safeDirs {
		if wsMergedPath == dir {
			log.Fatal(fmt.Errorf("Fatal merged dir: %s", wsMergedPath))
		}
	}
	ensureDir(wsMergedPath)
	// 复制base目录
	if err := copyDir(wsBasePath, wsMergedPath); err != nil {
		log.Fatal(err)
	}
	// 复制sync目录
	destPath := destSrcCase(sync, wsMergedPath)
	if err := copyDir(sync, destPath); err != nil {
		log.Fatal(err)
	}
}

func destSrcCase(syncPath, destPath string) string {
	syncSrcPath := path.Join(syncPath, "src")
	destSrcPath := path.Join(destPath, "src")

	_, syncErr := os.Stat(syncSrcPath)
	destSrcStat, mergedErr := os.Stat(destSrcPath)

	if os.IsNotExist(syncErr) && (mergedErr == nil && destSrcStat.IsDir()) {
		destPath = destSrcPath
	}
	return destPath
}

func cwd() string {
	dir, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		log.Fatal(err)
	}
	return strings.Replace(dir, "\\", "/", -1)
}

func consumeEvent(watcher *fsnotify.Watcher, event fsnotify.Event) {
	changedPath := event.Name
	replacePath := strings.Replace(changedPath, sync, "", 1)
	wsRealMergedPath := destSrcCase(sync, wsMergedPath)
	destPath := path.Join(wsRealMergedPath, replacePath)

	switch event.Op {
	case fsnotify.Create:
		successPrint("%s create: %s => %s\n", outputFlag, changedPath, destPath)
		watcher.Add(event.Name)
		if stat, err := os.Stat(changedPath); err == nil {
			if stat.IsDir() {
				ensureDir(destPath)
			} else {
				ensureDir(path.Dir(destPath))
				copyFile(changedPath, destPath)
			}
		}
	case fsnotify.Rename, fsnotify.Remove, fsnotify.Remove | fsnotify.Rename:
		successPrint("%s remove: %s => %s\n", outputFlag, changedPath, destPath)
		watcher.Remove(event.Name)
		os.RemoveAll(destPath)
	case fsnotify.Write:
		successPrint("%s write: %s => %s\n", outputFlag, changedPath, destPath)
		copyFile(changedPath, destPath)
	}
}

func registerWatcher() {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal(err)
	}
	defer watcher.Close()

	done := make(chan bool)
	go func() {
		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					return
				}
				consumeEvent(watcher, event)
			case err, ok := <-watcher.Errors:
				if !ok {
					return
				}
				errorPrint("%s error:%s\n", outputFlag, err.Error())
			}
		}
	}()

	err = watcher.Add(sync)
	if err != nil {
		log.Fatal(err)
	}
	err = filepath.Walk(sync, func(path string, info os.FileInfo, err error) error {
		if strings.Contains(path, "node_modules") {
			return nil
		}
		if info.IsDir() {
			if err := isIgnoreDir(path); err != nil {
				return err
			}
		}
		err = watcher.Add(path)
		if err != nil {
			log.Fatal(err)
		}
		return nil
	})
	if err != nil {
		log.Fatal(err)
	}
	<-done
}

func ensureDir(dir string) {
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		os.MkdirAll(dir, 0755)
	}
}
