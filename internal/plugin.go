package internal

import (
	"fmt"
	"github.com/urfave/cli"
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

func RunDrone() {
	app := cli.NewApp()
	app.Name = "shino"
	app.Usage = "shino plugin"
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
			Value:  baseURL,
			EnvVar: "PLUGIN_BASE",
		},
		cli.StringFlag{
			Name:   "sync",
			Usage:  "sync dir",
			Value:  syncDir,
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
				Base: c.String("config.base"),
				Sync: c.String("config.sync"),
			},
		}

		if err := plugin.exec(); err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}

// Exec 执行插件
func (p Plugin) exec() error {
	// 1.备份一次当前目录到/tmp/backup
	backupDir := path.Join(os.TempDir(), "backup")
	ensureDir(backupDir)
	syncDir := syncDir
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
