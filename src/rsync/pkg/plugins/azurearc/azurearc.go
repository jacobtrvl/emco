// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2022 Intel Corporation

package azurearc

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"net/url"

	"encoding/json"
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

type PropertiesFlux struct {
	Scope          string             `json:"scope"`
	Namespace      string             `json:"namespace"`
	SourceKind     string             `json:"sourceKind"`
	Suspend        bool               `json:"suspend"`
	GitRepository  RepoProperties     `json:"gitRepository"`
	Kustomizations KustomizationsUnit `json:"kustomizations"`
}

type RepoProperties struct {
	Url           string  `json:"url"`
	RepositoryRef RepoRef `json:"repositoryRef"`
}

type RepoRef struct {
	Branch string `json:"branch"`
}

type KustomizationsUnit struct {
	FirstKustomization KustomizationProperties `json:"kustomization-1"`
}

type KustomizationProperties struct {
	Path                  string `json:"path"`
	TimeoutInSeconds      int    `json:"timeoutInSeconds"`
	SyncIntervalInSeconds int    `json:"syncIntervalInSeconds"`
	Prune                 bool   `json:"prune"`
	Force                 bool   `json:"force"`
}

type Requestbody struct {
	Properties Properties `json:"properties"`
}

type RequestbodyFlux struct {
	Properties PropertiesFlux `json:"properties"`
}

type FluxExtension struct {
	Identity   IndentityProp `json:"identity"`
	Properties ExtensionProp `json:"properties"`
}

type IndentityProp struct {
	Type string `json:"type"`
}

type ExtensionProp struct {
	ExtensionType           string `json:"extensionType"`
	AutoUpgradeMinorVersion bool   `json:"autoUpgradeMinorVersion"`
}

/*
	Function to get the access token for azure arc
	params: clientId, ClientSecret, tenantIdValue
	return: Token, error
*/
func GetAccessToken(clientId string, clientSecret string, tenantIdValue string) (string, error) {
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
		return "", err
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded; param=value")

	res, err := client.Do(req)
	if err != nil {
		return "", err
	}

	responseData, err := ioutil.ReadAll(res.Body)
	if err != nil {
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
func CreateGitConfiguration(accessToken string, repositoryUrl string, gitConfiguration string, operatorScopeType string, subscriptionIdValue string, arcClusterResourceGroupName string, arcClusterName string, gitbranch string, gitpath string) (string, error) {
	// PUT request for creating git configuration
	// PUT request body
	client := http.Client{}

	flags := "--git-branch=" + gitbranch + " --sync-garbage-collection --git-path=" + gitpath

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
		return "", err
	}
	// Add request header
	authorizationString := "Bearer " + accessToken
	reqPut.Header.Set("Content-Type", "application/json; charset=UTF-8")
	reqPut.Header.Add("Authorization", authorizationString)

	resPut, err := client.Do(reqPut)
	if err != nil {
		//Handle Error
		return "", err
	}

	responseDataPut, err := ioutil.ReadAll(resPut.Body)
	if err != nil {
		//Handle Error
		return "", err
	}

	return string(responseDataPut), nil

}

/*
	Function to create a FluxV2 configuration for the mentioned user repo
	params : Access Token, RepositoryUrl, Flux configuration name, scope of flux, Subscription Id
			Arc Resource group Name, Arc Cluster Name, git branch and git path, timeOut(seconds), SyncInterval(seconds)
	return : response, error
*/
func CreateFluxConfiguration(accessToken string, repositoryUrl string, gitConfiguration string, operatorScopeType string, subscriptionIdValue string, arcClusterResourceGroupName string, arcClusterName string, gitbranch string, gitpath string, timeOut int, syncInterval int) (string, error) {
	// PUT request for creating git configuration
	// PUT request body
	client := http.Client{}
	if gitpath != "" {
		gitpath = "./" + gitpath
	}

	properties := RequestbodyFlux{
		PropertiesFlux{
			Scope:      operatorScopeType,
			Namespace:  gitConfiguration,
			SourceKind: "GitRepository",
			Suspend:    false,
			GitRepository: RepoProperties{
				Url: repositoryUrl,
				RepositoryRef: RepoRef{
					Branch: gitbranch}},
			Kustomizations: KustomizationsUnit{
				FirstKustomization: KustomizationProperties{
					Path:                  gitpath,
					TimeoutInSeconds:      timeOut,
					SyncIntervalInSeconds: syncInterval,
					Prune:                 true,
					Force:                 false}}}}

	dataProperties, err := json.Marshal(properties)
	if err != nil {
		return "", err
	}

	urlPut := "https://management.azure.com/subscriptions/" + subscriptionIdValue + "/resourceGroups/" + arcClusterResourceGroupName + "/providers/Microsoft.Kubernetes/connectedClusters/" + arcClusterName + "/providers/Microsoft.KubernetesConfiguration/fluxConfigurations/" + gitConfiguration + "?api-version=2022-01-01-preview"

	reqPut, err := http.NewRequest(http.MethodPut, urlPut, bytes.NewBuffer(dataProperties))
	if err != nil {
		//Handle Error
		return "", err
	}
	// Add request header
	authorizationString := "Bearer " + accessToken
	reqPut.Header.Set("Content-Type", "application/json; charset=UTF-8")
	reqPut.Header.Add("Authorization", authorizationString)

	resPut, err := client.Do(reqPut)
	if err != nil {
		//Handle Error
		return "", err
	}
	responseDataPut, err := ioutil.ReadAll(resPut.Body)
	if err != nil {
		return "", err
	}

	return string(responseDataPut), nil

}

/*
	Function to install Microsoft.flux extension
	params: Access token, Subscription Id Value, Arc Cluster Resource Group Name, Arc Cluster Name
	return: response, error
*/
func InstallFluxExtension(accessToken string, subscriptionIdValue string, arcClusterResourceGroupName string, arcClusterName string) (string, error) {
	// PUT request for installing microsoft.flux extension
	// PUT request body
	client := http.Client{}
	properties := FluxExtension{IndentityProp{"SystemAssigned"}, ExtensionProp{"microsoft.flux", true}}
	dataProperties, err := json.Marshal(properties)
	if err != nil {
		return "", err
	}

	urlPut := "https://management.azure.com/subscriptions/" + subscriptionIdValue + "/resourceGroups/" + arcClusterResourceGroupName + "/providers/Microsoft.Kubernetes/connectedClusters/" + arcClusterName + "/providers/Microsoft.KubernetesConfiguration/extensions/flux?api-version=2021-09-01"

	reqPut, err := http.NewRequest(http.MethodPut, urlPut, bytes.NewBuffer(dataProperties))
	if err != nil {
		//Handle Error
		return "", err
	}
	// Add request header
	authorizationString := "Bearer " + accessToken
	reqPut.Header.Set("Content-Type", "application/json; charset=UTF-8")
	reqPut.Header.Add("Authorization", authorizationString)

	resPut, err := client.Do(reqPut)
	if err != nil {
		//Handle Error
		return "", err
	}
	responseDataPut, err := ioutil.ReadAll(resPut.Body)
	if err != nil {
		return "", err
	}

	return string(responseDataPut), nil
}

/*
	Function to Delete Flux configuration
	params : Access Token, Subscription Id, Arc Cluster ResourceName, Arc Cluster Name, Flux Configuration name
	return : Response, error

*/
func DeleteFluxConfiguration(accessToken string, subscriptionIdValue string, arcClusterResourceGroupName string, arcClusterName string, gitConfiguration string) (string, error) {
	// Create client
	client := &http.Client{}
	// Create request
	urlDelete := "https://management.azure.com/subscriptions/" + subscriptionIdValue + "/resourceGroups/" + arcClusterResourceGroupName + "/providers/Microsoft.Kubernetes/connectedClusters/" + arcClusterName + "/providers/Microsoft.KubernetesConfiguration/fluxConfigurations/" + gitConfiguration + "?api-version=2021-11-01-preview"
	reqDelete, err := http.NewRequest("DELETE", urlDelete, nil)

	if err != nil {
		//Handle Error
		return "", err
	}
	// Add request header
	authorizationString := "Bearer " + accessToken
	reqDelete.Header.Set("Content-Type", "application/json; charset=UTF-8")
	reqDelete.Header.Add("Authorization", authorizationString)

	resPut, err := client.Do(reqDelete)
	if err != nil {
		//Handle Error
		return "", err
	}
	responseDataPut, err := ioutil.ReadAll(resPut.Body)
	if err != nil {
		return "", err
	}

	return string(responseDataPut), nil
}

/*
	Function to Delete Git configuration
	params : Access Token, Subscription Id, Arc Cluster ResourceName, Arc Cluster Name, Flux Configuration name
	return : Response, error

*/
func DeleteGitConfiguration(accessToken string, subscriptionIdValue string, arcClusterResourceGroupName string, arcClusterName string, gitConfiguration string) (string, error) {
	// Create client
	client := &http.Client{}
	// Create request
	urlDelete := "https://management.azure.com/subscriptions/" + subscriptionIdValue + "/resourceGroups/" + arcClusterResourceGroupName + "/providers/Microsoft.Kubernetes/connectedClusters/" + arcClusterName + "/providers/Microsoft.KubernetesConfiguration/sourceControlConfigurations/" + gitConfiguration + "?api-version=2021-03-01"

	reqDelete, err := http.NewRequest("DELETE", urlDelete, nil)
	if err != nil {
		//Handle Error
		return "", err
	}
	// Add request header
	authorizationString := "Bearer " + accessToken
	reqDelete.Header.Set("Content-Type", "application/json; charset=UTF-8")
	reqDelete.Header.Add("Authorization", authorizationString)

	resPut, err := client.Do(reqDelete)
	if err != nil {
		//Handle Error
		return "", err
	}
	responseDataPut, err := ioutil.ReadAll(resPut.Body)
	if err != nil {
		return "", err
	}

	return string(responseDataPut), nil
}
