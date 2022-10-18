// Copyright (c) 2017-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package main

import (
	"fmt"
	"net/http"

	jira "github.com/andygrunwald/go-jira"
	"github.com/hashicorp/go-version"
	"github.com/pkg/errors"
)

const (
	ServerInfoApiEndpoint = "rest/api/2/serverInfo"
	CreateMetaAPIEndpoint = "rest/api/2/issue/createmeta/"
	PivotVersion          = "8.4.0"
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

type IssueInfo struct {
	Values []*jira.MetaIssueType `json:"values,omitempty"`
}

type FieldInfo struct {
	Values []interface{} `json:"values,omitempty"`
}

type FieldValues struct {
	FieldID string `json:"fieldId,omitempty"`
}

type FieldID struct {
	Values []FieldValues `json:"values,omitempty"`
}

type Version struct {
	VersionInfo string `json:"version,omitempty"`
}

// GetCreateMeta returns the metadata needed to implement the UI and validation of
// creating new Jira issues.
func (client jiraServerClient) GetCreateMeta(options *jira.GetQueryOptions) (*jira.CreateMetaInfo, error) {
	v := new(Version)
	req, err := client.Jira.NewRequest(http.MethodGet, ServerInfoApiEndpoint, nil)
	if err != nil {
		return nil, err
	}

	if _, err = client.Jira.Do(req, v); err != nil {
		return nil, err
	}

	currentVersion, err := version.NewVersion(v.VersionInfo)
	if err != nil {
		return nil, err
	}

	pivotVersion, err := version.NewVersion(PivotVersion)
	if err != nil {
		return nil, err
	}

	var info *jira.CreateMetaInfo
	var resp *jira.Response
	if currentVersion.LessThan(pivotVersion) {
		info, resp, err = client.Jira.Issue.GetCreateMetaWithOptions(options)
	} else {
		cd, response, apiErr := client.Jira.Project.ListWithOptions(options)
		meta := new(jira.CreateMetaInfo)

		if apiErr == nil {
			for i := 0; i < len(*cd); i++ {
				meta.Expand = (*cd)[i].Expand
				apiEndpoint := fmt.Sprintf("%s%s/issuetypes", CreateMetaAPIEndpoint, (*cd)[i].ID)
				req, err = client.Jira.NewRequest(http.MethodGet, apiEndpoint, nil)
				if err != nil {
					break
				}

				issues := new(IssueInfo)
				response, err = client.Jira.Do(req, issues)
				if err != nil {
					break
				}

				project := &jira.MetaProject{
					Expand:     (*cd)[i].Expand,
					Self:       (*cd)[i].Self,
					Id:         (*cd)[i].ID,
					Key:        (*cd)[i].Key,
					Name:       (*cd)[i].Name,
					IssueTypes: issues.Values,
				}

				for _, issue := range project.IssueTypes {
					apiEndpoint := fmt.Sprintf("%s%s/issuetypes/%s", CreateMetaAPIEndpoint, (*cd)[i].ID, issue.Id)
					req, err = client.Jira.NewRequest(http.MethodGet, apiEndpoint, nil)
					if err != nil {
						break
					}

					field := new(FieldInfo)
					response, err = client.Jira.Do(req, field)
					if err != nil {
						break
					}

					fieldID := new(FieldID)
					response, err = client.Jira.Do(req, fieldID)
					if err != nil {
						break
					}

					fieldMap := make(map[string]interface{})
					for f, fieldValue := range field.Values {
						fieldMap[fieldID.Values[f].FieldID] = fieldValue
					}
					issue.Fields = fieldMap
				}
				meta.Projects = append(meta.Projects, project)
			}
		}
		info = meta
		resp = response
		if apiErr != nil {
			err = apiErr
		}
	}

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
