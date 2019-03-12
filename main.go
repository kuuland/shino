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

	"github.com/fsnotify/fsnotify"
)

var (
	base    = ""
	install = "npm install"
	start   = "npm start"
	watch   = "watch"

	wsPath       = ".shino"
	wsBasePath   = path.Join(wsPath, "base")
	wsMergedPath = path.Join(wsPath, "merged")
)

func init() {
	base = os.Getenv("BASE")
	install = os.Getenv("INSTALL")
	start = os.Getenv("START")
	watch = os.Getenv("WATCH")

	var cfg map[string]string
	cfgFile := "kuu.json"
	if _, err := os.Stat(cfgFile); err == nil {
		if data, err := ioutil.ReadFile(cfgFile); err == nil {
			json.Unmarshal(data, &cfg)
		}
	}
	if cfg != nil {
		if v, ok := cfg["base"]; ok {
			base = v
		}
		if v, ok := cfg["install"]; ok {
			install = v
		}
		if v, ok := cfg["start"]; ok {
			start = v
		}
		if v, ok := cfg["watch"]; ok {
			watch = v
		}
	}

	base = strings.TrimSpace(base)
	install = strings.TrimSpace(install)
	start = strings.TrimSpace(start)
	watch = strings.TrimSpace(watch)
}

func main() {
	setup()
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
	ctx, cancel := context.WithCancel(context.Background())
	//创建监听退出chan
	c := make(chan os.Signal)
	//监听指定信号 ctrl+c kill
	signal.Notify(c, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT, syscall.SIGUSR1, syscall.SIGUSR2)
	go func() {
		for s := range c {
			switch s {
			case syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT:
				fmt.Println("退出", s)
				cancel()
			case syscall.SIGUSR1:
				fmt.Println("usr1", s)
			case syscall.SIGUSR2:
				fmt.Println("usr2", s)
			default:
				fmt.Println("other", s)
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
	// 执行合并：.shino/base + watch = .shino/merged
	execMerge()
	// 执行install命令
	if install != "" {
		installCmd := mergedCmd(ctx, install)
		execCmd(installCmd)
	}
	// 执行start命令
	startCmd := mergedCmd(ctx, start)
	go execCmd(startCmd)
	// 启动监听器
	registerWatcher()
}

func logArgs(args []string) {
	output := "[SHINO]"
	for _, arg := range args {
		output = fmt.Sprintf("%s %s", output, arg)
	}
	fmt.Println(output)
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

func execMerge() {
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
	// 先删除merged
	if destInfo, err := os.Stat(wsMergedPath); err == nil && destInfo.IsDir() {
		// os.RemoveAll(wsMergedPath)
		return
	}
	ensureDir(wsMergedPath)
	// 复制base目录
	if err := copyDir(wsBasePath, wsMergedPath); err != nil {
		log.Fatal(err)
	}
	// 复制watch目录
	if err := copyDir(watch, wsMergedPath); err != nil {
		log.Fatal(err)
	}
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
	replacePath := strings.Replace(changedPath, watch, "", 1)
	destPath := path.Join(wsMergedPath, replacePath)

	switch event.Op {
	case fsnotify.Create:
		fmt.Printf("create: %s => %s\n", changedPath, destPath)
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
		log.Printf("remove: %s\n", destPath)
		watcher.Remove(event.Name)
		os.RemoveAll(destPath)
	case fsnotify.Write:
		log.Printf("write: %s => %s\n", changedPath, destPath)
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
				log.Println("error:", err)
			}
		}
	}()

	err = watcher.Add(watch)
	if err != nil {
		log.Fatal(err)
	}
	err = filepath.Walk(watch, func(path string, info os.FileInfo, err error) error {
		if strings.Contains(path, "node_modules") {
			return nil
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
