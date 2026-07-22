package main

import (
	"context"
	"fmt"
	"time"

	"bolt-monitor/shared/checkexecution"
	"bolt-monitor/shared/dynamodbrecord"
	sharederrors "bolt-monitor/shared/errors"
	"bolt-monitor/shared/escalation"
	"bolt-monitor/shared/notifications"
)

type schedulerConfigReader interface {
	GetSchedulerConfig(context.Context, string) (dynamodbrecord.SchedulerConfigRecord, error)
}

type schedulerConfigWriter interface {
	UpdateSchedulerConfig(context.Context, string, checkexecution.SchedulerConfig, time.Time) (dynamodbrecord.SchedulerConfigRecord, error)
}

type getSchedulerConfigQuery struct{ store schedulerConfigReader }
type updateSchedulerConfigCommand struct {
	store schedulerConfigWriter
	now   commandClock
}

type schedulerOperations struct {
	get    getSchedulerConfigQuery
	update updateSchedulerConfigCommand
}

func newSchedulerOperations(get schedulerConfigReader, update schedulerConfigWriter, now commandClock) schedulerOperations {
	return schedulerOperations{get: getSchedulerConfigQuery{store: get}, update: updateSchedulerConfigCommand{store: update, now: now}}
}

func (q getSchedulerConfigQuery) Execute(ctx context.Context, tenantID string) (dynamodbrecord.SchedulerConfigRecord, error) {
	return q.store.GetSchedulerConfig(ctx, tenantID)
}

func (c updateSchedulerConfigCommand) Execute(ctx context.Context, tenantID string, config checkexecution.SchedulerConfig) (dynamodbrecord.SchedulerConfigRecord, error) {
	if err := config.Validate(); err != nil {
		return dynamodbrecord.SchedulerConfigRecord{}, err
	}
	return c.store.UpdateSchedulerConfig(ctx, tenantID, config, c.now())
}

type searchResourcesStore interface {
	SearchResources(context.Context, string, string, int, map[string]struct{}) ([]searchResult, error)
}

type searchResourcesQuery struct{ store searchResourcesStore }

func (q searchResourcesQuery) Execute(ctx context.Context, tenantID, query string, limit int, types map[string]struct{}) ([]searchResult, error) {
	return q.store.SearchResources(ctx, tenantID, query, limit, types)
}

type listEscalationPoliciesStore interface {
	ListEscalationPolicies(context.Context, string) ([]escalation.EscalationPolicy, error)
}

type getEscalationPolicyStore interface {
	GetEscalationPolicy(context.Context, string, string) (*escalation.EscalationPolicy, error)
}

type createEscalationPolicyStore interface {
	CreateEscalationPolicy(context.Context, escalation.EscalationPolicy) (escalation.EscalationPolicy, error)
}

type updateEscalationPolicyStore interface {
	getEscalationPolicyStore
	UpdateEscalationPolicy(context.Context, escalation.EscalationPolicy) (escalation.EscalationPolicy, error)
}

type deleteEscalationPolicyStore interface {
	getEscalationPolicyStore
	DeleteEscalationPolicy(context.Context, string, string) error
}

type escalationPolicyChannelLookup interface {
	GetNotificationChannel(context.Context, string, string) (*escalation.NotificationChannel, error)
}

type createEscalationPolicyCommand struct {
	store    createEscalationPolicyStore
	channels escalationPolicyChannelLookup
	now      commandClock
	ids      identifierGenerator
}
type updateEscalationPolicyCommand struct {
	store    updateEscalationPolicyStore
	channels escalationPolicyChannelLookup
}
type deleteEscalationPolicyCommand struct {
	store      deleteEscalationPolicyStore
	references servicePolicyReferenceQuery
}
type listEscalationPoliciesQuery struct{ store listEscalationPoliciesStore }
type getEscalationPolicyQuery struct{ store getEscalationPolicyStore }
type serviceEscalationPolicyQuery struct {
	services serviceLookup
	policies getEscalationPolicyStore
}

type escalationPolicyOperations struct {
	create  createEscalationPolicyCommand
	update  updateEscalationPolicyCommand
	delete  deleteEscalationPolicyCommand
	list    listEscalationPoliciesQuery
	get     getEscalationPolicyQuery
	service serviceEscalationPolicyQuery
}

func newEscalationPolicyOperations(create createEscalationPolicyStore, update updateEscalationPolicyStore, delete deleteEscalationPolicyStore, list listEscalationPoliciesStore, get getEscalationPolicyStore, channels escalationPolicyChannelLookup, references servicePolicyReferenceQuery, services serviceLookup, now commandClock, ids identifierGenerator) escalationPolicyOperations {
	return escalationPolicyOperations{
		create: createEscalationPolicyCommand{store: create, channels: channels, now: now, ids: ids}, update: updateEscalationPolicyCommand{store: update, channels: channels},
		delete: deleteEscalationPolicyCommand{store: delete, references: references}, list: listEscalationPoliciesQuery{store: list}, get: getEscalationPolicyQuery{store: get},
		service: serviceEscalationPolicyQuery{services: services, policies: get},
	}
}

func (c createEscalationPolicyCommand) Execute(ctx context.Context, policy escalation.EscalationPolicy) (escalation.EscalationPolicy, error) {
	policy.PolicyID = c.ids.newEscalationPolicyID(c.now())
	if err := validateEscalationPolicyChannels(ctx, c.channels, policy); err != nil {
		return escalation.EscalationPolicy{}, err
	}
	return c.store.CreateEscalationPolicy(ctx, policy)
}

func (c updateEscalationPolicyCommand) Execute(ctx context.Context, policy escalation.EscalationPolicy) (escalation.EscalationPolicy, error) {
	if existing, err := c.store.GetEscalationPolicy(ctx, policy.TenantID, policy.PolicyID); err != nil {
		return escalation.EscalationPolicy{}, err
	} else if existing == nil {
		return escalation.EscalationPolicy{}, sharederrors.New(sharederrors.CodePolicyNotFound, nil)
	}
	if err := validateEscalationPolicyChannels(ctx, c.channels, policy); err != nil {
		return escalation.EscalationPolicy{}, err
	}
	return c.store.UpdateEscalationPolicy(ctx, policy)
}

func (c deleteEscalationPolicyCommand) Execute(ctx context.Context, tenantID, policyID string) error {
	if policy, err := c.store.GetEscalationPolicy(ctx, tenantID, policyID); err != nil {
		return err
	} else if policy == nil {
		return sharederrors.New(sharederrors.CodePolicyNotFound, nil)
	}
	if referenced, err := c.references.ServiceReferencesEscalationPolicy(ctx, tenantID, policyID); err != nil {
		return err
	} else if referenced {
		return sharederrors.New(sharederrors.CodePolicyReferenced, nil)
	}
	return c.store.DeleteEscalationPolicy(ctx, tenantID, policyID)
}

func (q listEscalationPoliciesQuery) Execute(ctx context.Context, tenantID string) ([]escalation.EscalationPolicy, error) {
	return q.store.ListEscalationPolicies(ctx, tenantID)
}
func (q getEscalationPolicyQuery) Execute(ctx context.Context, tenantID, policyID string) (*escalation.EscalationPolicy, error) {
	return q.store.GetEscalationPolicy(ctx, tenantID, policyID)
}
func (q serviceEscalationPolicyQuery) Execute(ctx context.Context, tenantID, serviceID string) (*escalation.EscalationPolicy, error) {
	service, found, err := q.services.GetService(ctx, tenantID, serviceID)
	if err != nil {
		return nil, err
	}
	if !found {
		return nil, sharederrors.New(sharederrors.CodeServiceNotFound, nil)
	}
	if service.EscalationPolicyID == "" {
		return nil, sharederrors.New(sharederrors.CodeServiceHasNoPolicy, nil)
	}
	policy, err := q.policies.GetEscalationPolicy(ctx, tenantID, service.EscalationPolicyID)
	if err != nil {
		return nil, err
	}
	if policy == nil {
		return nil, sharederrors.New(sharederrors.CodePolicyNotFound, nil)
	}
	return policy, nil
}

func validateEscalationPolicyChannels(ctx context.Context, channels escalationPolicyChannelLookup, policy escalation.EscalationPolicy) error {
	for _, path := range []struct {
		name string
		path escalation.EscalationPath
	}{{name: "businessHoursPath", path: policy.BusinessHoursPath}, {name: "offHoursPath", path: policy.OffHoursPath}} {
		for index, step := range path.path.Steps {
			channelID := step.ChannelID
			if channelID == "" {
				return sharederrors.New(sharederrors.CodeValidationFailed, map[string]any{"field": fmt.Sprintf("%s.steps[%d].channelId", path.name, index), "reason": "required"})
			}
			channel, err := channels.GetNotificationChannel(ctx, policy.TenantID, channelID)
			if err != nil {
				return err
			}
			if channel == nil {
				return sharederrors.New(sharederrors.CodeValidationFailed, map[string]any{"field": fmt.Sprintf("%s.steps[%d].channelId", path.name, index), "reason": "channel not found"})
			}
		}
	}
	return nil
}

type listNotificationChannelsStore interface {
	ListNotificationChannels(context.Context, string) ([]escalation.NotificationChannel, error)
}
type getNotificationChannelStore interface {
	GetNotificationChannel(context.Context, string, string) (*escalation.NotificationChannel, error)
}
type createNotificationChannelStore interface {
	CreateNotificationChannel(context.Context, escalation.NotificationChannel) (escalation.NotificationChannel, error)
}
type updateNotificationChannelStore interface {
	UpdateNotificationChannel(context.Context, escalation.NotificationChannel) (escalation.NotificationChannel, error)
}
type deleteNotificationChannelStore interface {
	DeleteNotificationChannel(context.Context, string, string) error
	ChannelsReferencedByRoutes(context.Context, string, string) ([]routeReference, error)
}
type notificationChannelAuditStore interface {
	RecordNotificationChannelTestAudit(context.Context, string, string, string, string, string, time.Time) error
}

type createNotificationChannelCommand struct {
	store createNotificationChannelStore
	now   commandClock
	ids   identifierGenerator
}
type updateNotificationChannelCommand struct {
	store updateNotificationChannelStore
}
type deleteNotificationChannelCommand struct {
	channels getNotificationChannelStore
	store    deleteNotificationChannelStore
}
type listNotificationChannelsQuery struct{ store listNotificationChannelsStore }
type getNotificationChannelQuery struct{ store getNotificationChannelStore }
type notificationChannelTestCommand struct {
	channels getNotificationChannelStore
	audit    notificationChannelAuditStore
	senders  notifications.SenderRegistry
	now      commandClock
}

type notificationChannelOperations struct {
	create createNotificationChannelCommand
	update updateNotificationChannelCommand
	delete deleteNotificationChannelCommand
	list   listNotificationChannelsQuery
	get    getNotificationChannelQuery
	test   notificationChannelTestCommand
}

func newNotificationChannelOperations(create createNotificationChannelStore, update updateNotificationChannelStore, delete deleteNotificationChannelStore, list listNotificationChannelsStore, get getNotificationChannelStore, audit notificationChannelAuditStore, senders notifications.SenderRegistry, now commandClock, ids identifierGenerator) notificationChannelOperations {
	return notificationChannelOperations{create: createNotificationChannelCommand{store: create, now: now, ids: ids}, update: updateNotificationChannelCommand{store: update}, delete: deleteNotificationChannelCommand{channels: get, store: delete}, list: listNotificationChannelsQuery{store: list}, get: getNotificationChannelQuery{store: get}, test: notificationChannelTestCommand{channels: get, audit: audit, senders: senders, now: now}}
}

func (c createNotificationChannelCommand) Execute(ctx context.Context, channel escalation.NotificationChannel) (escalation.NotificationChannel, error) {
	channel.ChannelID = c.ids.newNotificationChannelID(c.now())
	return c.store.CreateNotificationChannel(ctx, channel)
}
func (c updateNotificationChannelCommand) Execute(ctx context.Context, channel escalation.NotificationChannel) (escalation.NotificationChannel, error) {
	return c.store.UpdateNotificationChannel(ctx, channel)
}
func (c deleteNotificationChannelCommand) Execute(ctx context.Context, tenantID, channelID string) ([]routeReference, error) {
	channel, err := c.channels.GetNotificationChannel(ctx, tenantID, channelID)
	if err != nil {
		return nil, err
	}
	if channel == nil {
		return nil, sharederrors.New(sharederrors.CodeChannelNotFound, nil)
	}
	references, err := c.store.ChannelsReferencedByRoutes(ctx, tenantID, channelID)
	if err != nil || len(references) > 0 {
		return references, err
	}
	return nil, c.store.DeleteNotificationChannel(ctx, tenantID, channelID)
}
func (q listNotificationChannelsQuery) Execute(ctx context.Context, tenantID string) ([]escalation.NotificationChannel, error) {
	return q.store.ListNotificationChannels(ctx, tenantID)
}
func (q getNotificationChannelQuery) Execute(ctx context.Context, tenantID, channelID string) (*escalation.NotificationChannel, error) {
	return q.store.GetNotificationChannel(ctx, tenantID, channelID)
}

type notificationChannelTestResult struct {
	ChannelID string
	SentAt    time.Time
}

func (c notificationChannelTestCommand) Execute(ctx context.Context, tenantID, channelID string) (notificationChannelTestResult, error) {
	channel, err := c.channels.GetNotificationChannel(ctx, tenantID, channelID)
	if err != nil {
		return notificationChannelTestResult{}, err
	}
	if channel == nil {
		return notificationChannelTestResult{}, sharederrors.New(sharederrors.CodeChannelNotFound, nil)
	}
	now := c.now().UTC()
	record := func(outcome, reason string) error {
		return c.audit.RecordNotificationChannelTestAudit(ctx, tenantID, channel.ChannelID, string(channel.Type), outcome, reason, now)
	}
	sender, ok := c.senders.Get(string(channel.Type))
	if !ok {
		if err := record("failure", "notification delivery failed"); err != nil {
			return notificationChannelTestResult{}, err
		}
		return notificationChannelTestResult{}, sharederrors.New(sharederrors.CodeNotificationDelivery, map[string]any{"channelId": channel.ChannelID, "type": string(channel.Type), "reason": "notification delivery failed"})
	}
	config, err := mergeNotificationChannelTarget(*channel)
	if err != nil {
		if auditErr := record("failure", "notification delivery failed"); auditErr != nil {
			return notificationChannelTestResult{}, auditErr
		}
		return notificationChannelTestResult{}, sharederrors.New(sharederrors.CodeNotificationDelivery, map[string]any{"channelId": channel.ChannelID, "type": string(channel.Type), "reason": "notification delivery failed"})
	}
	notification := notifications.Notification{EventType: notifications.EventTypeIncidentDown, TenantID: tenantID, MonitorID: "notification-channel-test", ServiceID: "notification-channel-test", MonitorName: "Notification channel test", ServiceName: "Bolt Monitor", Timestamp: now, Message: "Bolt Monitor test notification\n\nChannel: " + channel.Name + "\nType: " + string(channel.Type) + "\nThis is a test message from the dashboard. No incident was created.", IncidentID: "notification-channel-test", Config: config}
	if _, err := sender.Send(ctx, notification); err != nil {
		if auditErr := record("failure", sanitizeNotificationDeliveryError(err, channel.Config)); auditErr != nil {
			return notificationChannelTestResult{}, auditErr
		}
		return notificationChannelTestResult{}, sharederrors.New(sharederrors.CodeNotificationDelivery, map[string]any{"channelId": channel.ChannelID, "type": string(channel.Type), "reason": sanitizeNotificationDeliveryError(err, channel.Config)})
	}
	if err := record("success", ""); err != nil {
		return notificationChannelTestResult{}, err
	}
	return notificationChannelTestResult{ChannelID: channel.ChannelID, SentAt: now}, nil
}
