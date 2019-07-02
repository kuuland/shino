package internal

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/fatih/color"
	"github.com/fsnotify/fsnotify"
	"github.com/urfave/cli"
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
)

func RunLocal() {
	app := cli.NewApp()
	app.Name = "shino"
	app.Usage = "CLI for Kuu"
	app.Commands = cli.Commands{
		{
			Name:  "up",
			Usage: "startup project",
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:   "base",
					Usage:  "base project",
					Value:  baseURL,
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
			},
			Action: func(c *cli.Context) {
				baseVal := c.String("base")
				installVal := c.String("install")
				startVal := c.String("start")
				syncVal := c.String("sync")

				if baseVal != "" {
					baseURL = strings.TrimSpace(baseVal)
				}
				if installVal != "" {
					installCmd = strings.TrimSpace(installVal)
				}
				if startVal != "" {
					startCmd = strings.TrimSpace(startVal)
				}
				if syncVal != "" {
					syncDir = strings.TrimSpace(syncVal)
				}

				localSetup()
			},
		},
		{
			Name:  "fano",
			Usage: "CLI for FanoJS",
			Subcommands: []cli.Command{
				{
					Name:  "table",
					Usage: "generate table pages based on metadata",
					Flags: []cli.Flag{
						cli.StringFlag{
							Name:  "meta",
							Usage: "metadata url",
						},
						cli.StringFlag{
							Name:  "out",
							Usage: "output dir",
						},
					},
					Action: func(c *cli.Context) error {
						fanoTable(c.String("meta"), c.String("out"))
						return nil
					},
				},
			},
		},
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
			if err := json.Unmarshal(data, &cfg); err != nil {
				log.Println(err)
			}
		}
	} else {
		return
	}
	if cfg != nil {
		if v, ok := cfg["base"]; ok {
			baseURL = strings.TrimSpace(v)
		}
		if v, ok := cfg["install"]; ok {
			installCmd = strings.TrimSpace(v)
		}
		if v, ok := cfg["start"]; ok {
			startCmd = strings.TrimSpace(v)
		}
		if v, ok := cfg["sync"]; ok {
			syncDir = strings.TrimSpace(v)
		}
	}
}

func localSetup() {
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
	if _, err := os.Stat(workBaseDir); os.IsNotExist(err) {
		ensureDir(workBaseDir)
		// 执行clone命令
		cloneCmd := clone(workDir, workBaseDir)
		execCmd(cloneCmd)
	}
	// 执行合并：.shino/base + sync = .shino/merged
	if _, err := os.Stat(workMergedDir); os.IsNotExist(err) {
		execMerge()
		// 执行install命令
		if installCmd != "" {
			installCmd := mergedCmd(ctx, installCmd)
			execCmd(installCmd)
		}
	}
	// 执行start命令
	startCmd := mergedCmd(ctx, startCmd)
	go execCmd(startCmd)
	// 启动监听器
	registerWatcher()
}
func mergedCmd(ctx context.Context, execStr string) *exec.Cmd {
	args := strings.Split(execStr, " ")
	cmd := exec.CommandContext(ctx, args[0])
	if len(args) > 1 {
		cmd.Args = args
	}
	cmd.Dir = workMergedDir
	return cmd
}

func errorPrint(format string, a ...interface{}) {
	if _, err := color.New(color.FgHiRed, color.Bold).Printf(format, a...); err != nil {
		log.Println(err)
	}
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
		if workMergedDir == dir {
			log.Fatal(fmt.Errorf("Fatal merged dir: %s", workMergedDir))
		}
	}
	ensureDir(workMergedDir)
	// 复制base目录
	if err := copyDir(workBaseDir, workMergedDir); err != nil {
		log.Fatal(err)
	}
	// 复制sync目录
	destPath := destSrcCase(syncDir, workMergedDir)
	if err := copyDir(syncDir, destPath); err != nil {
		log.Fatal(err)
	}
}

func consumeEvent(watcher *fsnotify.Watcher, event fsnotify.Event) {
	changedPath := event.Name
	replacePath := strings.Replace(changedPath, syncDir, "", 1)
	wsRealMergedPath := destSrcCase(syncDir, workMergedDir)
	destPath := path.Join(wsRealMergedPath, replacePath)

	switch event.Op {
	case fsnotify.Create:
		successPrint("%s create: %s => %s\n", outputFlag, changedPath, destPath)
		if err := watcher.Add(event.Name); err != nil {
			log.Println(err)
		}
		if stat, err := os.Stat(changedPath); err == nil {
			if stat.IsDir() {
				ensureDir(destPath)
			} else {
				ensureDir(path.Dir(destPath))
				if _, err := copyFile(changedPath, destPath); err != nil {
					log.Println(err)
				}
			}
		}
	case fsnotify.Rename, fsnotify.Remove, fsnotify.Remove | fsnotify.Rename:
		successPrint("%s remove: %s => %s\n", outputFlag, changedPath, destPath)
		if err := watcher.Remove(event.Name); err != nil {
			log.Println(err)
		}
		if err := os.RemoveAll(destPath); err != nil {
			log.Println(err)
		}
	case fsnotify.Write:
		successPrint("%s write: %s => %s\n", outputFlag, changedPath, destPath)
		if _, err := copyFile(changedPath, destPath); err != nil {
			log.Println(err)
		}
	}
}

func registerWatcher() {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal(err)
	}
	defer func() {
		if err := watcher.Close(); err != nil {
			log.Println(err)
		}
	}()

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

	err = watcher.Add(syncDir)
	if err != nil {
		log.Fatal(err)
	}
	err = filepath.Walk(syncDir, func(path string, info os.FileInfo, err error) error {
		if strings.Contains(path, "node_modules") {
			return nil
		}
		if info.IsDir() {
			if err := IsIgnoreDir(path); err != nil {
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
