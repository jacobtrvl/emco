// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2022 Intel Corporation

package controllers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"sync"

	"github.com/fluxcd/go-git-providers/github"
	"github.com/fluxcd/go-git-providers/gitprovider"
	gogithub "github.com/google/go-github/v41/github"
	k8spluginv1alpha1 "gitlab.com/project-emco/core/emco-base/src/monitor/pkg/apis/k8splugin/v1alpha1"
	log "gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/infra/logutils"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	githubDomain = "github.com"
	// maxrand      = 0x7fffffffffffffff
)

type GithubAccessClient struct {
	cl           gitprovider.Client
	gitUser      string
	gitRepo      string
	cluster      string
	githubDomain string
}

var GitHubClient GithubAccessClient

func SetupGitHubClient() error {
	var err error
	GitHubClient, err = NewGitHubClient()
	return err
}

func NewGitHubClient() (GithubAccessClient, error) {

	githubDomain := "github.com"
	gitUser := os.Getenv("GIT_USERNAME")
	gitToken := os.Getenv("GIT_TOKEN")
	gitRepo := os.Getenv("GIT_REPO")
	clusterName := os.Getenv("GIT_CLUSTERNAME")

	// If any value is not provided then can't store in Git location
	if len(gitRepo) <= 0 || len(gitToken) <= 0 || len(gitUser) <= 0 || len(clusterName) <= 0 {
		log.Info("Github information not found:: Skipping Github storage", log.Fields{})
		return GithubAccessClient{}, nil
	}
	log.Info("GitHub Info found", log.Fields{"gitRepo::": gitRepo, "cluster::": clusterName})

	cl, err := github.NewClient(gitprovider.WithOAuth2Token(gitToken), gitprovider.WithDestructiveAPICalls(true))
	if err != nil {
		return GithubAccessClient{}, err
	}
	return GithubAccessClient{
		cl:           cl,
		gitUser:      gitUser,
		gitRepo:      gitRepo,
		githubDomain: githubDomain,
		cluster:      clusterName,
	}, nil
}

func CommitCR(c client.Client, cr *k8spluginv1alpha1.ResourceBundleState, org *k8spluginv1alpha1.ResourceBundleStateStatus) error {

	// Compare status and update if status changed
	resBytesCr, err := json.Marshal(cr.Status)
	if err != nil {
		log.Info("json Marshal error for resource::", log.Fields{"cr": cr, "err": err})
		return err
	}
	resBytesOrg, err := json.Marshal(org)
	if err != nil {
		log.Info("json Marshal error for resource::", log.Fields{"cr": cr, "err": err})
		return err
	}
	// If the status is not changed no need to update CR
	if bytes.Compare(resBytesCr, resBytesOrg) == 0 {
		return nil
	}
	err = c.Status().Update(context.TODO(), cr)
	if err != nil {
		if k8serrors.IsConflict(err) {
			return err
		} else {
			log.Info("CR Update Error::", log.Fields{"err": err})
			return err
		}
	}
	resBytes, err := json.Marshal(cr)
	if err != nil {
		log.Info("json Marshal error for resource::", log.Fields{"cr": cr, "err": err})
		return err
	}
	// Check if GIT Info is provided if so store the information in the Git Repo also
	err = GitHubClient.CommitCRToGitHub(resBytes, cr.GetLabels())
	if err != nil {
		log.Info("Error commiting status to Github", log.Fields{"err": err})
	}
	return nil
}

// var mutex = sync.Mutex{}

func (c *GithubAccessClient) CommitCRToGitHub(resBytes []byte, l map[string]string) error {

	// Check if Github Client is available
	if c.cl == nil {
		return nil
	}
	// Get cid and app id
	v, ok := l["emco/deployment-id"]
	if !ok {
		return fmt.Errorf("Unexpected error:: Inconsistent labels %v", l)
	}
	result := strings.SplitN(v, "-", 2)
	if len(result) != 2 {
		return fmt.Errorf("Unexpected error:: Inconsistent labels %v", l)
	}
	app := result[1]
	cid := result[0]
	path := "clusters/" + c.cluster + "/status/" + cid + "/app/" + app + "/" + v

	// userRef := gitprovider.UserRef{
	// 	Domain:    c.githubDomain,
	// 	UserLogin: c.gitUser,
	// }
	// // Create the repo reference
	// userRepoRef := gitprovider.UserRepositoryRef{
	// 	UserRef:        userRef,
	// 	RepositoryName: c.gitRepo,
	// }
	s := string(resBytes)
	var files []gitprovider.CommitFile
	files = append(files, gitprovider.CommitFile{
		Path:    &path,
		Content: &s,
	})
	commitMessage := "Adding Status for " + path

	appName := c.cluster + "-" + cid + "-" + app
	// commitfiles
	err := c.CommitFiles(context.Background(), "main", commitMessage, appName, files)
	// // Only one process to commit to Github location to avoid conflicts
	// mutex.Lock()
	// defer mutex.Unlock()
	// userRepo, err := c.cl.UserRepositories().Get(context.Background(), userRepoRef)
	// if err != nil {
	// 	return err
	// }
	// //Commit file to this repo to a branch status
	// _, err = userRepo.Commits().Create(context.Background(), "main", commitMessage, files)
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

var mutex = sync.Mutex{}

/*
	Function to commit multiple files to the github repo
	params : context, github client, User Name, Repo Name, Branch Name, Commit Message, files ([]gitprovider.CommitFile)
	return : nil/error
*/
func (c *GithubAccessClient) CommitFiles(ctx context.Context, branch, commitMessage, appName string, files []gitprovider.CommitFile) error {

	// create a new go github client
	githubToken := "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"

	tp := gogithub.BasicAuthTransport{
		Username: c.gitUser,
		Password: githubToken,
	}

	client := gogithub.NewClient(tp.Client())

	// obtain the sha key for main
	//obtain shaake
	latestSHA, err := GetLatestCommitSHA(ctx, client, c.gitUser, c.gitRepo, branch, "")
	if err != nil {
		return err
	}
	// //create a new branch from main
	// ra := rand.New(rand.NewSource(time.Now().UnixNano()))
	// rn := ra.Int63n(maxrand)
	// id := fmt.Sprintf("%v", rn)

	mergeBranch := appName
	err = createBranch(ctx, client, latestSHA, c.gitUser, c.gitRepo, mergeBranch)

	// defer deletion of the created branch
	// //delete the branch
	// defer deleteBranch(ctx, client, c.gitUser, c.gitRepo, mergeBranch)

	if err != nil {
		if !strings.Contains(err.Error(), "422 Reference already exists") {
			return err
		}
		fmt.Println("Branch Already exists: ", err)
	}
	// Only one process to commit to Github location to avoid conflicts
	mutex.Lock()
	defer mutex.Unlock()

	// commit the files to this new branch
	// create repo reference
	log.Info("Creating Repo Reference. ", log.Fields{})
	userRepoRef := getRepoRef(c.gitUser, c.gitRepo)
	log.Info("UserRepoRef:", log.Fields{"UserRepoRef": userRepoRef})

	log.Info("Obtaining user repo. ", log.Fields{})
	userRepo, err := c.cl.UserRepositories().Get(ctx, userRepoRef)
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

	// // merge the branch to the main
	// //merge the branch
	// err = mergeBranchToMain(ctx, client, c.gitUser, c.gitRepo, branch, mergeBranch)

	// 	if err != nil {
	// 		// check error for merge conflict "409 Merge conflict"
	// 		if strings.Contains(err.Error(), "409 Merge conflict") && n < 3 {
	// 			// Merge conflict flag
	// 			n++
	// 			log.Error("Merge Conflict, trying again!", log.Fields{"err": err})
	// 			continue
	// 		} else {
	// 			return err
	// 		}

	// }
	return nil
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
func deleteBranch(ctx context.Context, c *gogithub.Client, userName, repoName, mergeBranch string) error {

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
