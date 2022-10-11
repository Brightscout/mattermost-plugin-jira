// Copyright (c) 2017-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package main

import (
	"net/http"

	jira "github.com/andygrunwald/go-jira"
	"github.com/hashicorp/go-version"
	"github.com/pkg/errors"
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
	Id string `json:"fieldId,omitempty"`
}

type FieldId struct {
	Values []FieldValues `json:"values,omitempty"`
}
type Version struct {
	VersionInfo string `json:"version,omitempty"`
}

// GetCreateMeta returns the metadata needed to implement the UI and validation of
// creating new Jira issues.
func (client jiraServerClient) GetCreateMeta(options *jira.GetQueryOptions) (*jira.CreateMetaInfo, error) {
	apiEndpoint := "rest/api/2/serverInfo"
	v := new(Version)
	req, err := client.Jira.NewRequest("GET", apiEndpoint, nil)
	if err != nil {
		return nil, err
	}
	_, err = client.Jira.Do(req, v)
	if err != nil {
		return nil, err
	}
	v1, err := version.NewVersion(v.VersionInfo)
	if err != nil {
		return nil, err
	}
	v2, err := version.NewVersion("8.4.0")
	if err != nil {
		return nil, err
	}
	if v1.LessThan(v2) {
		cimd, resp, err := client.Jira.Issue.GetCreateMetaWithOptions(options)
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
		return cimd, nil
	}
	cd, resp, err := client.Jira.Project.ListWithOptions(options)
	meta := new(jira.CreateMetaInfo)
	for i := 0; i < len(*cd); i++ {
		meta.Expand = (*cd)[i].Expand
		project := new(jira.MetaProject)
		project.Self = (*cd)[i].Self
		project.Id = (*cd)[i].ID
		project.Key = (*cd)[i].Key
		project.Name = (*cd)[i].Name
		apiEndpoint := "rest/api/2/issue/createmeta/" + (*cd)[i].ID + "/issuetypes"
		req, err := client.Jira.NewRequest("GET", apiEndpoint, nil)
		if err != nil {
			break
		}
		issues := new(IssueInfo)
		resp, err = client.Jira.Do(req, issues)
		if err != nil {
			break
		}
		project.IssueTypes = issues.Values
		for j := 0; j < len(project.IssueTypes); j++ {
			apiEndpoint := "rest/api/2/issue/createmeta/" + (*cd)[i].ID + "/issuetypes/" + project.IssueTypes[j].Id
			req, err := client.Jira.NewRequest("GET", apiEndpoint, nil)
			if err != nil {
				break
			}
			field := new(FieldInfo)
			resp, err = client.Jira.Do(req, field)
			if err != nil {
				break
			}
			fieldId := new(FieldId)
			resp, err = client.Jira.Do(req, fieldId)
			if err != nil {
				break
			}
			newMap := make(map[string]interface{})
			for k := 0; k < len(field.Values); k++ {
				newMap[fieldId.Values[k].Id] = field.Values[k]
			}
			project.IssueTypes[j].Fields = newMap
		}
		proj := meta.Projects
		proj = append(proj, project)
		meta.Projects = proj
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
	return meta, nil
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
