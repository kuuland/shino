package internal

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/fatih/color"
	"io"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"
)

var (
	baseURL    = "https://github.com/kuuland/ui.git"
	installCmd = "npm install"
	startCmd   = "npm start"
	syncDir    = cwd()

	workDir       = path.Join(".shino")
	workBaseDir   = path.Join(workDir, "base")
	workMergedDir = path.Join(workDir, "merged")

	outputFlag = "[SHINO]"
)

func copyDir(srcPath string, destPath string) error {
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
			p := strings.Replace(path, "\\", "/", -1)
			destNewPath := strings.Replace(p, srcPath, destPath, -1)
			if _, err := copyFile(p, destNewPath); err != nil {
				log.Println(err)
			}
		} else {
			if err := IsIgnoreDir(path); err != nil {
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

func ensureDir(dir string) {
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		if err := os.MkdirAll(dir, 0755); err != nil {
			log.Println(err)
		}
	}
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

func execCmd(cmd *exec.Cmd) {
	buf := new(bytes.Buffer)
	cmd.Stdout = io.MultiWriter(os.Stdout, buf)
	cmd.Stderr = io.MultiWriter(os.Stderr, buf)
	logArgs(cmd.Args)
	if err := cmd.Run(); err != nil {
		log.Fatal(err)
	}
}

func logArgs(args []string) {
	output := outputFlag
	for _, arg := range args {
		output = fmt.Sprintf("%s %s", output, arg)
	}
	successPrint(fmt.Sprintf("%s\n", output))
}

func successPrint(format string, a ...interface{}) {
	if _, err := color.New(color.FgHiGreen, color.Bold).Printf(format, a...); err != nil {
		log.Println(err)
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

func IsIgnoreDir(path string) error {
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
	defer func() {
		if err := srcFile.Close(); err != nil {
			log.Println(err)
		}
	}()
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
	defer func() {
		if err := dstFile.Close(); err != nil {
			log.Println(err)
		}
	}()

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

func isEmptyDir(dirname string) bool {
	dir, _ := ioutil.ReadDir(dirname)
	return len(dir) == 0
}
