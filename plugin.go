package main

import (
	"log"
	"os"
	"os/exec"
	"path"
	"strings"
)

type (
	// Repo 代码库信息
	Repo struct {
		Owner   string
		Name    string
		Link    string
		Avatar  string
		Branch  string
		Private bool
		Trusted bool
	}
	// Build 编译信息
	Build struct {
		Number   int
		Event    string
		Status   string
		Deploy   string
		Created  int64
		Started  int64
		Finished int64
		Link     string
	}
	// Commit 提交信息
	Commit struct {
		Remote  string
		Sha     string
		Ref     string
		Link    string
		Branch  string
		Message string
		Author  Author
	}
	// Author 提交人信息
	Author struct {
		Name   string
		Email  string
		Avatar string
	}
	// Config 插件配置
	Config struct {
		// plugin-specific parameters and secrets
		Base    string
		Install string
		Build   string
		Sync    string
	}
	// Plugin 插件
	Plugin struct {
		Repo   Repo
		Build  Build
		Commit Commit
		Config Config
	}
)

// Exec 执行插件
func (p Plugin) Exec() error {
	// 1.备份一次当前目录到系统临时目录
	backupDir := path.Join(os.TempDir(), "backup")
	syncDir := sync
	if !path.IsAbs(syncDir) {
		syncDir = path.Join(cwd(), syncDir)
	}
	if stat, err := os.Stat(syncDir); err != nil || !stat.IsDir() {
		log.Fatal(err)
	}
	if err := copyDir(syncDir, backupDir); err != nil {
		log.Fatal(err)
	}
	os.RemoveAll(syncDir)
	// 2.克隆base到当前目录
	cloneCmd := clone(p.Config.Base, backupDir)
	execCmd(cloneCmd)
	// 3.复制备份目录到当前目录
	destPath := destSrcCase(backupDir, cwd())
	if err := copyDir(backupDir, destPath); err != nil {
		log.Fatal(err)
	}
	// 4.执行install命令
	p.install()
	// 5.执行build命令
	p.build()
	return nil
}

func (p Plugin) install() {
	if p.Config.Install != "" {
		args := strings.Split(p.Config.Install, " ")
		cmd := exec.Command(args[0])
		if len(args) > 1 {
			cmd.Args = args
		}
		execCmd(cmd)
	}
}

func (p Plugin) build() {
	args := strings.Split(p.Config.Build, " ")
	cmd := exec.Command(args[0])
	if len(args) > 1 {
		cmd.Args = args
	}
	execCmd(cmd)
}
