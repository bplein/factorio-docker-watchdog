package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"sort"
	"strings"

	"github.com/sirupsen/logrus"
)

const (
	templateTagsPrefix = "<!-- start autogeneration tags -->"
	templateTagsSuffix = "<!-- end autogeneration tags -->"
)

func gitSetupCredentials() error {
	cmd := exec.Command("git", "config", "--global", "user.email", os.Getenv("GIT_EMAIL"))
	output, err := cmd.CombinedOutput()
	if err != nil {
		return errors.New(string(output))
	}

	cmd = exec.Command("git", "config", "--global", "user.name", os.Getenv("GIT_NAME"))
	output, err = cmd.CombinedOutput()
	if err != nil {
		return errors.New(string(output))
	}

	return nil
}

func gitCloneRepo(path string) error {
	logrus.Debugln("git clone", fmt.Sprintf("https://github.com/%s/%s.git", githubRepoOwner, githubRepoName))
	cmd := exec.Command("git", "clone", fmt.Sprintf("https://%s:%s@github.com/%s/%s.git", githubUser, githubToken, githubRepoOwner, githubRepoName), path)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return errors.New(string(output))
	}
	return nil
}

func gitCreateCommit(path string, commitMessage string) error {
	logrus.Debugln("git add")
	cmd := exec.Command("git", "add", ".")
	cmd.Dir = path
	output, err := cmd.CombinedOutput()
	if err != nil {
		return errors.New(string(output))
	}

	logrus.Debugln("git commit", commitMessage)
	cmd = exec.Command("git", "commit", "-am", commitMessage)
	cmd.Dir = path
	output, err = cmd.CombinedOutput()
	if err != nil {
		return errors.New(string(output))
	}
	return nil
}

func gitPush(path string, branch string) error {
	args := []string{"push"}
	if branch != "" {
		args = append(args, "--set-upstream", "origin", branch)
	}
	cmd := exec.Command("git", args...)
	cmd.Dir = path
	output, err := cmd.CombinedOutput()
	if err != nil {
		return errors.New(string(output))
	}
	return nil
}

func editReadme(path string, version BuildInfo) error {
	file, err := ioutil.ReadFile(path + "/README.md")
	if err != nil {
		return err
	}

	startIndex := bytes.Index(file, []byte(templateTagsPrefix))
	endIndex := bytes.Index(file, []byte(templateTagsSuffix))
	if startIndex == -1 || endIndex == -1 {
		return errors.New("unable to find start or end tags in README.md")
	}

	var versions []string
	for k := range version.Versions {
		versions = append(versions, k)
	}
	sort.Sort(sort.Reverse(sort.StringSlice(versions)))

	generatedTags := ""
	for _, v := range versions {
		tags := make([]string, len(version.Versions[v].Tags))
		for i, tag := range version.Versions[v].Tags {
			tags[i] = fmt.Sprintf("`%s`", tag)
		}

		generatedTags += fmt.Sprintf("* %s\n", strings.Join(tags, ", "))
	}

	content := string(file[:startIndex]) + templateTagsPrefix + "\n" + generatedTags + string(file[endIndex:])

	err = ioutil.WriteFile(path+"/README.md", []byte(content), 0666)
	if err != nil {
		return err
	}

	return nil
}

func editBuildinfo(path string, buildinfo BuildInfo) error {
	f, err := os.OpenFile(path+"/buildinfo.json", os.O_CREATE|os.O_RDWR|os.O_TRUNC, 0644)
	if err != nil {
		return err
	}

	encoder := json.NewEncoder(f)
	encoder.SetIndent("", "  ")
	return encoder.Encode(buildinfo.Versions)
}
