package main

import (
	"os"
)

var (
	githubUser      = os.Getenv("GITHUB_USER")
	githubToken     = os.Getenv("GITHUB_TOKEN")
	githubRepoOwner = os.Getenv("GITHUB_REPO_OWNER") // current organization factoriotools
	githubRepoName  = os.Getenv("GITHUB_REPO_NAME")  // current repo factorio-docker
)
