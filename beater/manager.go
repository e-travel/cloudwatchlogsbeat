package beater

import (
	"strings"
	"time"

	"github.com/e-travel/cloudwatchlogsbeat/config"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudwatchlogs"
	"github.com/elastic/beats/libbeat/logp"
)

type GroupManager struct {
	prospectors []config.Prospector
	beat        *Cloudwatchlogsbeat
	groups      map[string]*Group
}

func NewGroupManager(beat *Cloudwatchlogsbeat) *GroupManager {
	return &GroupManager{
		prospectors: beat.Config.Prospectors,
		beat:        beat,
		groups:      make(map[string]*Group),
	}
}

func (manager *GroupManager) refreshGroups() {
	for _, prospector := range manager.prospectors {
		prospector := prospector
		for _, groupName := range prospector.GroupNames {
			groupName := groupName
			// If input group name doesn't end with a star, then consider it a
			// normal group name
			if !strings.HasSuffix(groupName, "*") {
				if _, ok := manager.groups[groupName]; !ok {
					manager.addNewGroup(groupName, &prospector)
				}
				continue
			}
			// If the input group name ends with a star, then consider it a prefix and
			// find all group names with that prefix
			err := manager.beat.AWSClient.DescribeLogGroupsPages(
				&cloudwatchlogs.DescribeLogGroupsInput{
					LogGroupNamePrefix: aws.String(groupName[:len(groupName)-1]),
				},
				func(page *cloudwatchlogs.DescribeLogGroupsOutput, lastPage bool) bool {
					for _, logGroup := range page.LogGroups {
						groupName := aws.StringValue(logGroup.LogGroupName)
						if _, ok := manager.groups[groupName]; !ok {
							manager.addNewGroup(groupName, &prospector)
						}
					}
					return true
				},
			)
			if err != nil {
				logp.Warn("manager: Failed to describe log group %s [%s]", groupName, err.Error())
			}
		}
	}
}

func (manager *GroupManager) addNewGroup(name string, prospector *config.Prospector) {
	group := NewGroup(name, prospector, manager.beat)
	manager.groups[group.Name] = group
	go group.Monitor()
}

func (manager *GroupManager) Monitor() {
	ticker := time.NewTicker(manager.beat.Config.GroupRefreshFrequency)
	defer ticker.Stop()
	reportTicker := time.NewTicker(reportFrequency)
	defer reportTicker.Stop()
	for {
		select {
		case <-ticker.C:
			manager.refreshGroups()
		case <-reportTicker.C:
			manager.report()
		}
	}
}

func (manager *GroupManager) report() {
	logp.Info("report[manager] %d %d", len(manager.prospectors), len(manager.groups))
}
