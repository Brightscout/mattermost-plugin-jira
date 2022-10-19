package main

import (
	"github.com/mattermost/mattermost-server/v6/model"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"

	"github.com/mattermost/mattermost-plugin-jira/server/utils/types"
)

const (
	settingOn  = "on"
	settingOff = "off"

	errStoreNewSettings = "Could not store new settings. Please contact your system administrator. Error: %v"
	errConnectToJira    = "Your account is not connected to Jira. Please type `/jira connect`. %v"

	notificationRoleAssignee = "assignee"
	notificationRoleMention  = "mention"
	notificationRoleReporter = "reporter"
	notificationRoleWatching = "watching"
)

func (connection *Connection) updateRolesForDMNotification(role string, hasNotification bool) bool {
	if role != notificationRoleAssignee && role != notificationRoleMention && role != notificationRoleReporter && role != notificationRoleWatching {
		return false
	}
	if connection.Settings.RolesForDMNotification == nil {
		connection.Settings.RolesForDMNotification = make(map[string]bool)
	}
	connection.Settings.RolesForDMNotification[role] = hasNotification
	return true
}
func (p *Plugin) settingsNotifications(header *model.CommandArgs, instanceID, mattermostUserID types.ID, connection *Connection, args []string) *model.CommandResponse {
	const helpText = "`/jira settings notifications [assignee|mention|reporter|watching] [value]`\n* Invalid value. Accepted values are: `on` or `off`."

	if len(args) != 3 {
		return p.responsef(header, helpText)
	}

	var value bool
	switch args[2] {
	case settingOn:
		value = true
	case settingOff:
		value = false
	default:
		return p.responsef(header, helpText)
	}

	if connection.Settings == nil {
		connection.Settings = &ConnectionSettings{}
	}
	if !connection.updateRolesForDMNotification(args[1], value) {
		return p.responsef(header, helpText)
	}

	if err := p.userStore.StoreConnection(instanceID, mattermostUserID, connection); err != nil {
		p.errorf("settingsNotifications, err: %v", err)
		p.responsef(header, errStoreNewSettings, err)
	}

	// send back the actual value
	updatedConnection, err := p.userStore.LoadConnection(instanceID, mattermostUserID)
	if err != nil {
		return p.responsef(header, errConnectToJira, err)
	}
	notifications := settingOff
	if updatedConnection.Settings.RolesForDMNotification[args[1]] {
		notifications = settingOn
	}

	return p.responsef(header, "Settings updated.\n\t%s notifications %s.", cases.Title(language.Und, cases.NoLower).String(args[1]), notifications)
}
