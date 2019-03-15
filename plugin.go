package main

import (
	"log"
	"os"
	"path"
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
	// 1.备份一次当前目录到/tmp/backup
	backupDir := path.Join(os.TempDir(), "backup")
	ensureDir(backupDir)
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
	// 2.克隆base到/tmp/base
	baseDir := path.Join(os.TempDir(), "base")
	cloneCmd := clone(p.Config.Base, baseDir)
	execCmd(cloneCmd)
	// 3.复制base目录到当前目录
	if err := copyDir(baseDir, cwd()); err != nil {
		log.Fatal(err)
	}
	// 4.复制备份目录到当前目录
	destPath := destSrcCase(backupDir, cwd())
	if err := copyDir(backupDir, destPath); err != nil {
		log.Fatal(err)
	}
	return nil
}
