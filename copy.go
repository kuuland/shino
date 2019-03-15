package main

import (
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"path"
	"path/filepath"
	"strings"
)

func copyDir(srcPath string, destPath string) error {
	cwd := cwd()
	if !path.IsAbs(srcPath) {
		srcPath = path.Join(cwd, srcPath)
	}
	if !path.IsAbs(destPath) {
		destPath = path.Join(cwd, destPath)
	}
	//检测目录正确性
	if srcInfo, err := os.Stat(srcPath); err != nil {
		log.Fatal(err)
	} else {
		if !srcInfo.IsDir() {
			log.Fatal(errors.New("srcPath不是一个正确的目录！"))
		}
	}
	if destInfo, err := os.Stat(destPath); err != nil {
		log.Fatal(err)
	} else {
		if !destInfo.IsDir() {
			log.Fatal(errors.New("destInfo不是一个正确的目录！"))
		}
	}

	err := filepath.Walk(srcPath, func(path string, f os.FileInfo, err error) error {
		if f == nil {
			return err
		}
		if !f.IsDir() {
			path := strings.Replace(path, "\\", "/", -1)
			destNewPath := strings.Replace(path, srcPath, destPath, -1)
			copyFile(path, destNewPath)
		} else {
			if err := isIgnoreDir(path); err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		log.Fatal(err)
	}
	return err
}

func isIgnoreDir(path string) error {
	ignoreDirs := []string{".shino", "node_modules", ".git", ".idea", ".vscode", ".history"}
	for _, dir := range ignoreDirs {
		if strings.HasSuffix(path, dir) {
			return filepath.SkipDir
		}
	}
	return nil
}

func copyFile(src, dest string) (w int64, err error) {
	srcFile, err := os.Open(src)
	if err != nil {
		log.Fatal(err)
	}
	defer srcFile.Close()
	//分割path目录
	destSplitPathDirs := strings.Split(dest, "/")

	//检测时候存在目录
	destSplitPath := ""
	for index, dir := range destSplitPathDirs {
		if index < len(destSplitPathDirs)-1 {
			destSplitPath = destSplitPath + dir + "/"
			b, _ := pathExists(destSplitPath)
			if b == false {
				//创建目录
				err := os.Mkdir(destSplitPath, os.ModePerm)
				if err != nil {
					log.Fatal(err)
				}
			}
		}
	}
	dstFile, err := os.Create(dest)
	if err != nil {
		log.Fatal(err)
	}
	defer dstFile.Close()

	return io.Copy(dstFile, srcFile)
}

//检测文件夹路径时候存在
func pathExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}
