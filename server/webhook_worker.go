// Copyright (c) 2017-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package main

import (
	jira "github.com/andygrunwald/go-jira"

	"github.com/mattermost/mattermost-plugin-jira/server/utils/types"
)

type webhookWorker struct {
	id        int
	p         *Plugin
	workQueue <-chan *webhookMessage
}

type webhookMessage struct {
	InstanceID types.ID
	Data       []byte
}

func (ww webhookWorker) work() {
	for msg := range ww.workQueue {
		err := ww.process(msg)
		if err != nil {
			ww.p.errorf("WebhookWorker id: %d, error processing, err: %v", ww.id, err)
		}
	}
}

func (ww webhookWorker) process(msg *webhookMessage) (err error) {
	defer func() {
		if err == ErrWebhookIgnored {
			// ignore ErrWebhookIgnored - from here up it's a success
			err = nil
		}
	}()

	wh, err := ParseWebhook(msg.Data)
	if err != nil {
		return err
	}

	if _, _, err = wh.PostNotifications(ww.p, msg.InstanceID); err != nil {
		ww.p.errorf("WebhookWorker id: %d, error posting notifications, err: %v", ww.id, err)
	}

	v := wh.(*webhook)
	if err = v.JiraWebhook.expandIssue(ww.p, msg.InstanceID); err != nil {
		return err
	}
	// We will only send webhook events if we have a connected instance.
	instance, err := ww.p.instanceStore.LoadInstance(msg.InstanceID)
	if err != nil {
		// This isn't an internal server error. There's just no instance installed.
		return err
	}

	comments := v.JiraWebhook.Issue.Fields.Comments.Comments
	mattermostUserID, err := ww.p.userStore.LoadMattermostUserID(instance.GetID(), comments[len(comments)-1].Author.AccountID)
	if err != nil {
		return err
	}

	// Check if the user has permissions.
	c, err := ww.p.userStore.LoadConnection(instance.GetID(), mattermostUserID)
	if err != nil {
		// Not connected to Jira, so can't check permissions
		return err
	}

	client, err := instance.GetClient(c)
	if err != nil {
		return err
	}

	// If this is a comment-related webhook, we need to check if they have permissions to read that.
	// Otherwise, check if they can view the issue.
	isCommentEvent := wh.Events().Intersection(commentEvents).Len() > 0
	visibilityAttribute := ""
	if isCommentEvent {
		visibility := jira.Comment{}
		err = client.RESTGet(v.JiraWebhook.Comment.Self, nil, &visibility)
		if err != nil {
			return err
		}

		visibilityAttribute = visibility.Visibility.Value
	}

	channelsSubscribed, err := ww.p.getChannelsSubscribed(v, msg.InstanceID, visibilityAttribute)
	if err != nil {
		return err
	}

	botUserID := ww.p.getUserID()
	for _, channelSubscribed := range channelsSubscribed {
		if _, _, err1 := wh.PostToChannel(ww.p, msg.InstanceID, channelSubscribed.ChannelID, botUserID, channelSubscribed.Name); err1 != nil {
			ww.p.errorf("WebhookWorker id: %d, error posting to channel, err: %v", ww.id, err1)
		}
	}

	return nil
}
