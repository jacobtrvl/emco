// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2022 Intel Corporation

package emcogithub

import (
	"context"
	"fmt"
	"math/rand"
	"strings"
	"time"

	"github.com/fluxcd/go-git-providers/github"
	"github.com/fluxcd/go-git-providers/gitprovider"
	gogithub "github.com/google/go-github/v41/github"
	log "gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/infra/logutils"
	// git "github.com/google/go-github/v41/github"
)

const (
	githubDomain = "github.com"
	maxrand      = 0x7fffffffffffffff
)

/*
	Function to create gitprovider githubClient
	params : github token
	return : gitprovider github client, error
*/
func CreateClient(githubToken string) (gitprovider.Client, error) {
	c, err := github.NewClient(gitprovider.WithOAuth2Token(githubToken), gitprovider.WithDestructiveAPICalls(true))
	if err != nil {
		return nil, err
	}
	return c, nil
}

/*
	Function to create go githubClient
	params : userName, github token
	return : go github client, error
*/
func CreateGoGitClient(userName, githubToken string) (*gogithub.Client, error) {

	tp := gogithub.BasicAuthTransport{
		Username: userName,
		Password: githubToken,
	}
	c := gogithub.NewClient(tp.Client())

	return c, nil
}

/*
	Function to create a new Repo in github
	params : context, github client, Repository Name, User Name, description
	return : nil/error
*/
func CreateRepo(ctx context.Context, c gitprovider.Client, repoName string, userName string, desc string) error {

	// create repo reference
	userRepoRef := getRepoRef(userName, repoName)

	// Create repoinfo reference
	userRepoInfo := gitprovider.RepositoryInfo{
		Description: &desc,
		Visibility:  gitprovider.RepositoryVisibilityVar(gitprovider.RepositoryVisibilityPublic),
	}

	// Create the repository
	_, err := c.UserRepositories().Create(ctx, userRepoRef, userRepoInfo, &gitprovider.RepositoryCreateOptions{
		AutoInit:        gitprovider.BoolVar(true),
		LicenseTemplate: gitprovider.LicenseTemplateVar(gitprovider.LicenseTemplateApache2),
	})

	if err != nil {
		return err
	}
	log.Info("Repo Created", log.Fields{})

	return nil
}

/*
	Function to commit multiple files to the github repo
	params : context, github client, User Name, Repo Name, Branch Name, Commit Message, files ([]gitprovider.CommitFile)
	return : nil/error
*/
func CommitFiles(ctx context.Context, c gitprovider.Client, userName, repoName, branch, commitMessage, appName string, files []gitprovider.CommitFile) error {

	// create a new go github client
	githubToken := "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"

	tp := gogithub.BasicAuthTransport{
		Username: userName,
		Password: githubToken,
	}

	client := gogithub.NewClient(tp.Client())

	n := 0
	for {
		// obtain the sha key for main
		//obtain sha
		latestSHA, err := GetLatestCommitSHA(ctx, client, userName, repoName, branch, "")
		if err != nil {
			return err
		}
		//create a new branch from main
		ra := rand.New(rand.NewSource(time.Now().UnixNano()))
		rn := ra.Int63n(maxrand)
		id := fmt.Sprintf("%v", rn)

		mergeBranch := appName + "-" + id
		err = createBranch(ctx, client, latestSHA, userName, repoName, mergeBranch)

		// defer deletion of the created branch
		//delete the branch
		defer DeleteBranch(ctx, client, userName, repoName, mergeBranch)

		if err != nil {
			return err
		}
		// commit the files to this new branch
		// create repo reference
		log.Info("Creating Repo Reference. ", log.Fields{})
		userRepoRef := getRepoRef(userName, repoName)
		log.Info("UserRepoRef:", log.Fields{"UserRepoRef": userRepoRef})

		log.Info("Obtaining user repo. ", log.Fields{})
		userRepo, err := c.UserRepositories().Get(ctx, userRepoRef)
		if err != nil {
			return err
		}
		log.Info("UserRepo:", log.Fields{"UserRepo": userRepo})

		log.Info("Commiting Files:", log.Fields{"files": files})
		//Commit file to this repo
		resp, err := userRepo.Commits().Create(ctx, mergeBranch, commitMessage, files)
		if err != nil {
			log.Error("Error in commiting the files", log.Fields{"err": err, "mergeBranch": mergeBranch, "commitMessage": commitMessage, "files": files})
			return err
		}
		log.Info("CommitResponse for userRepo:", log.Fields{"resp": resp})

		// merge the branch to the main
		//merge the branch
		err = mergeBranchToMain(ctx, client, userName, repoName, branch, mergeBranch)

		if err != nil {
			// check error for merge conflict "409 Merge conflict"
			if strings.Contains(err.Error(), "409 Merge conflict") && n < 3 {
				// Merge conflict flag
				n++
				log.Error("Merge Conflict, trying again!", log.Fields{"err": err})
				continue
			} else {
				return err
			}
		}
		return nil
	}
	return nil
}

/*
	Function to delete repo
	params : context, gitprovider client , user name, repo name
	return : nil/error
*/
func DeleteRepo(ctx context.Context, c gitprovider.Client, userName string, repoName string) error {

	// create repo reference
	userRepoRef := getRepoRef(userName, repoName)
	// get the reference of the repo to be deleted
	userRepo, err := c.UserRepositories().Get(ctx, userRepoRef)

	if err != nil {
		return err
	}
	//delete repo
	err = userRepo.Delete(ctx)

	if err != nil {
		return err
	}

	return nil
}

/*
	Internal function to create a repo refercnce
	params : user name, repo name
	return : repo reference
*/
func getRepoRef(userName string, repoName string) gitprovider.UserRepositoryRef {
	// Create the user reference
	userRef := gitprovider.UserRef{
		Domain:    githubDomain,
		UserLogin: userName,
	}

	// Create the repo reference
	userRepoRef := gitprovider.UserRepositoryRef{
		UserRef:        userRef,
		RepositoryName: repoName,
	}

	return userRepoRef
}

/*
	Function to Add file to the commit
	params : path , content, files (gitprovider commitfile array)
	return : files (gitprovider commitfile array)
*/
func Add(path string, content string, files []gitprovider.CommitFile) []gitprovider.CommitFile {
	files = append(files, gitprovider.CommitFile{
		Path:    &path,
		Content: &content,
	})

	return files
}

/*
	Function to Delete file from the commit
	params : path, files (gitprovider commitfile array)
	return : files (gitprovider commitfile array)
*/
func Delete(path string, files []gitprovider.CommitFile) []gitprovider.CommitFile {
	// check if the file exists for this path
	files = append(files, gitprovider.CommitFile{
		Path:    &path,
		Content: nil,
	})

	return files
}

/*
	Function to get files to the github repo
	params : context, github client, User Name, Repo Name, Branch Name, path)
	return : []*gitprovider.CommitFile, nil/error
*/
func GetFiles(ctx context.Context, c gitprovider.Client, userName string, repoName string, branch string, path string) ([]*gitprovider.CommitFile, error) {

	// create repo reference
	userRepoRef := getRepoRef(userName, repoName)
	userRepo, err := c.UserRepositories().Get(ctx, userRepoRef)
	if err != nil {
		return nil, err
	}
	// Read the files
	cf, err := userRepo.Files().Get(ctx, path, branch)
	if err != nil {
		return nil, err
	}
	return cf, nil
}

/*
	Function to obtaion the SHA of latest commit
	params : context, go github client, User Name, Repo Name, Branch, Path
	return : LatestCommit string, error
*/
func GetLatestCommitSHA(ctx context.Context, c *gogithub.Client, userName, repoName, branch, path string) (string, error) {

	perPage := 1
	page := 1

	lcOpts := &gogithub.CommitsListOptions{
		ListOptions: gogithub.ListOptions{
			// func getMainSHA(ctx context.Context, c *git.Client, userName, repoName, branch string) (string, error) {
			// 	// obtain latet sha of main branch
			// 	perPage := 1
			// 	page := 1

			// lcOpts := &git.CommitsListOptions{
			// 	ListOptions: git.ListOptions{
			PerPage: perPage,
			Page:    page,
		},
		SHA:  branch,
		Path: path,
	}
	//Get the latest SHA
	resp, _, err := c.Repositories.ListCommits(ctx, userName, repoName, lcOpts)
	if err != nil {
		log.Error("Error in obtaining the list of commits", log.Fields{"err": err})
		return "", err
	}
	if len(resp) == 0 {
		log.Info("File not created yet.", log.Fields{"Latest Commit Array": resp})
		return "", nil
	}
	latestCommitSHA := *resp[0].SHA

	return latestCommitSHA, nil
}

func createBranch(ctx context.Context, c *gogithub.Client, latestCommitSHA, userName, repoName, branch string) error {
	// create a new branch
	ref, _, err := c.Git.CreateRef(ctx, userName, repoName, &gogithub.Reference{
		Ref: gogithub.String("refs/heads/" + branch),
		Object: &gogithub.GitObject{
			SHA: gogithub.String(latestCommitSHA),
		},
	})
	if err != nil {
		log.Error("Git.CreateRef returned error:", log.Fields{"err": err})
		return err

	}
	log.Info("Branch Created: ", log.Fields{"ref": ref})
	return nil
}

//function to merge the branch to main
func mergeBranchToMain(ctx context.Context, c *gogithub.Client, userName, repoName, branch, mergeBranch string) error {
	// merge the branch
	input := &gogithub.RepositoryMergeRequest{
		Base:          gogithub.String(branch),
		Head:          gogithub.String(mergeBranch),
		CommitMessage: gogithub.String("merging " + mergeBranch + " to " + branch),
	}

	commit, _, err := c.Repositories.Merge(ctx, userName, repoName, input)
	if err != nil {
		log.Error("Error occured while Merging", log.Fields{"err": err})
		return err
	}

	log.Info("Branch Merged, Merge response:", log.Fields{"commit": commit})

	return nil

}

// Function to delete the branch
func DeleteBranch(ctx context.Context, c *gogithub.Client, userName, repoName, mergeBranch string) error {

	// Delete the Git branch
	_, err := c.Git.DeleteRef(ctx, userName, repoName, "refs/heads/"+mergeBranch)
	if err != nil {
		log.Error("Git.DeleteRef returned error: ", log.Fields{"err": err})
		return err
	}
	log.Info("Branch Deleted", log.Fields{"mergeBranch": mergeBranch})
	return nil
}

// Check if file exists
func CheckIfFileExists(ctx context.Context, c *gogithub.Client, userName, repoName, branch, path string) (bool, error) {
	latestSHA, err := GetLatestCommitSHA(ctx, c, userName, repoName, branch, path)
	if err != nil {
		return false, err
	}

	if latestSHA == "" {
		return false, nil
	}

	return true, nil

}
