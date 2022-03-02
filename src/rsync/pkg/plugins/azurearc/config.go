// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2022 Intel Corporation

package azurearc

import (
	"bytes"
	"context"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/url"
	"time"

	"github.com/fluxcd/go-git-providers/gitprovider"
	log "gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/infra/logutils"
	emcogit "gitlab.com/project-emco/core/emco-base/src/rsync/pkg/gitops/emcogit"
)

type Token struct {
	TokenType    string `json:"token_type"`
	ExpiresIn    string `json:"expires_in"`
	ExtExpiresIn string `json:"ext_expires_in"`
	ExpiresOn    string `json:"expires_on"`
	NotBefore    string `json:"not_before"`
	Resource     string `json:"resource"`
	AccessToken  string `json:"access_token"`
}

type Properties struct {
	RepositoryUrl         string `json:"repositoryUrl"`
	OperatorNamespace     string `json:"operatorNamespace"`
	OperatorInstanceName  string `json:"operatorInstanceName"`
	OperatorType          string `json:"operatorType"`
	OperatorParams        string `json:"operatorParams"`
	OperatorScope         string `json:"operatorScope"`
	SshKnownHostsContents string `json:"sshKnownHostsContents"`
}

type Requestbody struct {
	Properties Properties `json:"properties"`
}

/*
	Function to get the access token for azure arc
	params: clientId, ClientSecret, tenantIdValue
	return: Token, error
*/
func (p *AzureArcProvider) getAccessToken(clientId string, clientSecret string, tenantIdValue string) (string, error) {

	log.Info("Inside getAccessToken", log.Fields{})

	//Rest api to get the access token
	client := http.Client{}

	data := url.Values{}
	data.Set("grant_type", "client_credentials")
	data.Add("client_id", clientId)
	data.Add("resource", "https://management.core.windows.net/")
	data.Add("client_secret", clientSecret)

	urlPost := "https://login.microsoftonline.com/" + tenantIdValue + "/oauth2/token"

	req, err := http.NewRequest("POST", urlPost, bytes.NewBufferString(data.Encode()))
	if err != nil {
		//Handle Error
		log.Error("Couldn't create Azure Access Token request", log.Fields{"err": err, "req": req})
		return "", err
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded; param=value")

	res, err := client.Do(req)
	if err != nil {
		log.Error(" Azure Access Token response error", log.Fields{"err": err, "res": res})
		return "", err
	}

	responseData, err := ioutil.ReadAll(res.Body)
	if err != nil {
		log.Error(" Azure Access Token response marshall error", log.Fields{"err": err, "responseData": responseData})
		return "", err
	}

	// Unmarshall the response body into json and get token value
	newToken := Token{}
	json.Unmarshal(responseData, &newToken)

	return newToken.AccessToken, nil
}

/*
	Function to create a git configuration for the mentioned user repo
	params: accessToken, repositoryUrl, gitConfiguration, operatorScopeType, subscriptionId, Arc Cluster ResourceGroup, Arc ClusterName
			git branch, git path
	return: response, error
*/
func (p *AzureArcProvider) createGitConfiguration(accessToken string, repositoryUrl string, gitConfiguration string, operatorScopeType string, subscriptionIdValue string, arcClusterResourceGroupName string, arcClusterName string, gitbranch string, gitpath string) (string, error) {
	// PUT request for creating git configuration
	// PUT request body
	log.Info("Inside CreateGitConfiguration Azure Arc", log.Fields{})
	client := http.Client{}

	flags := "--git-branch=" + gitbranch + " --git-poll-interval=1s --sync-garbage-collection --git-path=" + gitpath

	properties := Requestbody{
		Properties{
			RepositoryUrl:         repositoryUrl,
			OperatorNamespace:     gitConfiguration,
			OperatorInstanceName:  gitConfiguration,
			OperatorType:          "flux",
			OperatorParams:        flags,
			OperatorScope:         operatorScopeType,
			SshKnownHostsContents: ""}}

	dataProperties, err := json.Marshal(properties)
	if err != nil {
		return "", err
	}

	urlPut := "https://management.azure.com/subscriptions/" + subscriptionIdValue + "/resourceGroups/" + arcClusterResourceGroupName + "/providers/Microsoft.Kubernetes/connectedClusters/" + arcClusterName + "/providers/Microsoft.KubernetesConfiguration/sourceControlConfigurations/" + gitConfiguration + "?api-version=2021-03-01"

	reqPut, err := http.NewRequest(http.MethodPut, urlPut, bytes.NewBuffer(dataProperties))
	if err != nil {
		//Handle Error
		log.Error("Error in creating request for git configuration", log.Fields{"err": err})
		return "", err
	}
	// Add request header
	authorizationString := "Bearer " + accessToken
	reqPut.Header.Set("Content-Type", "application/json; charset=UTF-8")
	reqPut.Header.Add("Authorization", authorizationString)

	resPut, err := client.Do(reqPut)
	if err != nil {
		//Handle Error
		log.Error("Error in response from creating git configuration", log.Fields{"err": err})
		return "", err
	}

	responseDataPut, err := ioutil.ReadAll(resPut.Body)
	if err != nil {
		//Handle Error
		log.Error("Error in parsing response from creating git configuration", log.Fields{"err": err})
		return "", err
	}

	log.Info("Response for GitConfiguration creation:", log.Fields{"ResponseData": string(responseDataPut)})
	return string(responseDataPut), nil

}

/*
	Function to add dummy file to prevent the path getting deleted
	params:
	return

*/
func (p *AzureArcProvider) addDummyFile(ctx context.Context) error {
	c, err := emcogit.CreateClient(p.gitProvider.GitToken, p.gitProvider.GitType)
	if err != nil {
		return err
	}
	files := []gitprovider.CommitFile{}

	path := "clusters/" + p.gitProvider.Cluster + "/context/" + p.gitProvider.Cid
	files = emcogit.Add(path+"/DoNotDelete", "Dummy file", files, p.gitProvider.GitType).([]gitprovider.CommitFile)

	response := emcogit.CommitFiles(ctx, c, p.gitProvider.UserName, p.gitProvider.RepoName, "main", "New Commit", files, p.gitProvider.GitType)
	if response != nil {
		log.Error("Couldn't commit the file", log.Fields{"response": response})
		return err
	}

	return nil
}

/*
	Function to Delete Git configuration
	params : Access Token, Subscription Id, Arc Cluster ResourceName, Arc Cluster Name, Flux Configuration name
	return : Response, error

*/
func (p *AzureArcProvider) deleteGitConfiguration(accessToken string, subscriptionIdValue string, arcClusterResourceGroupName string, arcClusterName string, gitConfiguration string) (string, error) {

	log.Info("Inside DeleteGitConfiguration Azure Arc", log.Fields{})
	// Create client
	client := &http.Client{}
	// Create request
	urlDelete := "https://management.azure.com/subscriptions/" + subscriptionIdValue + "/resourceGroups/" + arcClusterResourceGroupName + "/providers/Microsoft.Kubernetes/connectedClusters/" + arcClusterName + "/providers/Microsoft.KubernetesConfiguration/sourceControlConfigurations/" + gitConfiguration + "?api-version=2021-03-01"

	reqDelete, err := http.NewRequest("DELETE", urlDelete, nil)
	if err != nil {
		//Handle Error
		log.Error("Error in request of delete configuration", log.Fields{"Response": reqDelete, "err": err})
		return "", err
	}
	// Add request header
	authorizationString := "Bearer " + accessToken
	reqDelete.Header.Set("Content-Type", "application/json; charset=UTF-8")
	reqDelete.Header.Add("Authorization", authorizationString)

	resPut, err := client.Do(reqDelete)
	if err != nil {
		//Handle Error
		log.Error("Error in response of delete configuration", log.Fields{"Response": resPut, "err": err})
		return "", err
	}
	responseDataPut, err := ioutil.ReadAll(resPut.Body)
	if err != nil {
		log.Error("Error in parsing response of delete configuration", log.Fields{"Response": responseDataPut, "err": err})
		return "", err
	}

	return string(responseDataPut), nil
}

// create gitconfiguration of fluxv1 type in azure
func (p *AzureArcProvider) ApplyConfig(ctx context.Context, config interface{}) error {
	//get accesstoken for azure
	log.Info("Inside ApplyConfig AzureArc", log.Fields{})

	//Add dummy file to git
	resp := p.addDummyFile(ctx)
	if resp != nil {
		log.Error("Couldn't add dummy file", log.Fields{"err": resp})
		return resp
	}
	accessToken, err := p.getAccessToken(p.clientID, p.clientSecret, p.tenantID)

	log.Info("Obtained AccessToken: ", log.Fields{"accessToken": accessToken})

	if err != nil {
		log.Error("Couldn't obtain access token", log.Fields{"err": err, "accessToken": accessToken})
		return err
	}

	//gitConfiguration := "config-" + p.gitProvider.Cluster + "-" + p.gitProvider.Cid
	gitConfiguration := "config-" + p.gitProvider.Cid // what should be the gitconfiguration name? Something unique like config-p.cluster-p.cid
	operatorScope := "cluster"
	gitPath := "clusters/" + p.gitProvider.Cluster + "/context/" + p.gitProvider.Cid //what should be the git path?
	gitBranch := "main"

	_, err = p.createGitConfiguration(accessToken, p.gitProvider.Url, gitConfiguration, operatorScope, p.subscriptionID,
		p.arcResourceGroup, p.arcCluster, gitBranch, gitPath)

	if err != nil {
		log.Error("Error in creating git configuration", log.Fields{"err": err})
		return err
	}

	return nil

}

// Delete the git configuration
func (p *AzureArcProvider) DeleteConfig(ctx context.Context, config interface{}) error {

	//Delete the files from the path?
	// Maybe delete repo ("clusters/" + p.cluster + "/context/" + p.cid)
	log.Info("Inside DeleteConfig Azure Arc", log.Fields{})
	//delete configuration
	//get accesstoken for azure
	accessToken, err := p.getAccessToken(p.clientID, p.clientSecret, p.tenantID)

	if err != nil {
		return err
	}

	// gitConfiguration := "config-" + p.gitProvider.Cluster + "-" + p.gitProvider.Cid // what should be the gitconfiguration name? Something unique like config-p.cluster-p.cid
	gitConfiguration := "config-" + p.gitProvider.Cid

	time.Sleep(1 * time.Second)

	_, err = p.deleteGitConfiguration(accessToken, p.subscriptionID, p.arcResourceGroup, p.arcCluster, gitConfiguration)

	if err != nil {
		return err
	}

	return nil
}
