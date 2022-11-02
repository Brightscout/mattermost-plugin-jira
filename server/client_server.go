// Copyright (c) 2017-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package main

import (
	"fmt"
	"net/http"

	jira "github.com/andygrunwald/go-jira"
	"github.com/blang/semver/v4"
	"github.com/pkg/errors"
)

const (
	APIEndpointGetServerInfo                     = "rest/api/2/serverInfo"
	APIEndpointCreateIssueMeta                   = "rest/api/2/issue/createmeta/"
	JiraVersionWithOldCreateMetaBreakingEndPoint = "9.0.0"
)

type jiraServerClient struct {
	JiraClient
}

func newServerClient(jiraClient *jira.Client) Client {
	return &jiraServerClient{
		JiraClient: JiraClient{
			Jira: jiraClient,
		},
	}
}

type ProjectIssueInfo struct {
	Values []*jira.MetaIssueType `json:"values"`
}

type FieldInfo struct {
	Values []map[string]interface{} `json:"values"`
}

type FieldValues struct {
	FieldID string `json:"fieldId"`
}

type FieldID struct {
	Values []FieldValues `json:"values"`
}

type JiraServerVersionInfo struct {
	Version string `json:"version"`
}

// GetIssueInfo returns the issues information based on project id.
func (client jiraServerClient) GetIssueInfo(projectID string) (*ProjectIssueInfo, *jira.Response, error) {
	apiEndpoint := fmt.Sprintf("%s%s/issuetypes", APIEndpointCreateIssueMeta, projectID)
	req, err := client.Jira.NewRequest(http.MethodGet, apiEndpoint, nil)
	if err != nil {
		return nil, nil, err
	}

	issues := ProjectIssueInfo{}
	response, err := client.Jira.Do(req, &issues)
	return &issues, response, err
}

func (client jiraServerClient) GetProjectInfo(currentVersion semver.Version, pivotVersion semver.Version, options *jira.GetQueryOptions) (*jira.CreateMetaInfo, *jira.Response, error) {
	var info *jira.CreateMetaInfo
	var resp *jira.Response
	var req *http.Request
	var issueInfo *ProjectIssueInfo
	var projectList *jira.ProjectList
	var err error

	if currentVersion.LT(pivotVersion) {
		info, resp, err = client.Jira.Issue.GetCreateMetaWithOptions(options)
	} else {
		projectList, resp, err = client.Jira.Project.ListWithOptions(options)
		meta := new(jira.CreateMetaInfo)

		if err == nil {
			for _, proj := range *projectList {
				meta.Expand = proj.Expand
				issueInfo, resp, err = client.GetIssueInfo(proj.ID)
				if err != nil {
					break
				}

				for _, issueType := range issueInfo.Values {
					apiEndpoint := fmt.Sprintf("%s%s/issuetypes/%s", APIEndpointCreateIssueMeta, proj.ID, issueType.Id)
					req, err = client.Jira.NewRequest(http.MethodGet, apiEndpoint, nil)
					if err != nil {
						break
					}

					fieldInfo := FieldInfo{}
					resp, err = client.Jira.Do(req, &fieldInfo)
					if err != nil {
						break
					}

					fieldID := new(FieldID)
					resp, err = client.Jira.Do(req, fieldID)
					if err != nil {
						break
					}

					fieldMap := make(map[string]interface{})
					for index, fieldValue := range fieldInfo.Values {
						fieldMap[fieldID.Values[index].FieldID] = fieldValue
					}
					issueType.Fields = fieldMap
				}
				project := &jira.MetaProject{
					Expand:     proj.Expand,
					Self:       proj.Self,
					Id:         proj.ID,
					Key:        proj.Key,
					Name:       proj.Name,
					IssueTypes: issueInfo.Values,
				}

				meta.Projects = append(meta.Projects, project)
			}
		}
		info = meta
	}
	return info, resp, err
}

// GetCreateMeta returns the metadata needed to implement the UI and validation of
// creating new Jira issues.
func (client jiraServerClient) GetCreateMeta(options *jira.GetQueryOptions) (*jira.CreateMetaInfo, error) {
	v := new(JiraServerVersionInfo)
	req, err := client.Jira.NewRequest(http.MethodGet, APIEndpointGetServerInfo, nil)
	if err != nil {
		return nil, err
	}

	if _, err = client.Jira.Do(req, v); err != nil {
		return nil, errors.Wrap(err, "failed to fetch Jira server version")
	}

	currentVersion, err := semver.Make(v.Version)
	if err != nil {
		return nil, errors.Wrap(err, "error while parsing version")
	}

	pivotVersion, err := semver.Make(JiraVersionWithOldCreateMetaBreakingEndPoint)
	if err != nil {
		return nil, errors.Wrap(err, "error while parsing version")
	}

	info, resp, err := client.GetProjectInfo(currentVersion, pivotVersion, options)

	if err != nil {
		if resp == nil {
			return nil, err
		}
		resp.Body.Close()
		if resp.StatusCode == http.StatusForbidden || resp.StatusCode == http.StatusUnauthorized {
			err = errors.New("not authorized to create issues")
		}
		return nil, RESTError{err, resp.StatusCode}
	}
	return info, nil
}

// SearchUsersAssignableToIssue finds all users that can be assigned to an issue.
func (client jiraServerClient) SearchUsersAssignableToIssue(issueKey, query string, maxResults int) ([]jira.User, error) {
	return SearchUsersAssignableToIssue(client, issueKey, "username", query, maxResults)
}

// SearchUsersAssignableInProject finds all users that can be assigned to some issue in a given project.
func (client jiraServerClient) SearchUsersAssignableInProject(projectKey, query string, maxResults int) ([]jira.User, error) {
	return SearchUsersAssignableInProject(client, projectKey, "username", query, maxResults)
}

// GetUserGroups returns the list of groups that a user belongs to.
func (client jiraServerClient) GetUserGroups(connection *Connection) ([]*jira.UserGroup, error) {
	var result struct {
		Groups struct {
			Items []*jira.UserGroup
		}
	}
	err := client.RESTGet("2/myself", map[string]string{"expand": "groups"}, &result)
	if err != nil {
		return nil, err
	}
	return result.Groups.Items, nil
}

func (client jiraServerClient) ListProjects(query string, limit int) (jira.ProjectList, error) {
	plist, resp, err := client.Jira.Project.GetList()
	if err != nil {
		return nil, userFriendlyJiraError(resp, err)
	}
	if plist == nil {
		return jira.ProjectList{}, nil
	}
	result := *plist
	if limit > 0 && len(result) > limit {
		result = result[:limit]
	}
	return result, nil
}
