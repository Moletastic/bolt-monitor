package main

import (
	"context"
	"fmt"
	"net/url"
	"sort"
	"strings"
	"unicode"

	sharedaws "bolt-monitor/shared/aws"
	"bolt-monitor/shared/dynamodbrecord"
	"bolt-monitor/shared/dynamodbschema"
	sharederrors "bolt-monitor/shared/errors"
	"bolt-monitor/shared/escalation"
	"bolt-monitor/shared/monitorconfig"
)

const (
	searchResourceService = "service"
	searchResourceMonitor = "monitor"
	searchResourcePolicy  = "policy"
	searchResourceChannel = "channel"

	minSearchQueryLength = 2
	defaultSearchLimit   = 8
	maxSearchLimit       = 25
	searchReadLimit      = 75
)

type searchResult struct {
	Type        string `json:"type"`
	ID          string `json:"id"`
	ServiceID   string `json:"serviceId,omitempty"`
	Label       string `json:"label"`
	Description string `json:"description"`
	Href        string `json:"href"`
	IconKey     string `json:"iconKey"`
	MatchText   string `json:"matchText"`
}

type searchResponse struct {
	Results []searchResult `json:"results"`
}

func (r *dynamoMonitorRepository) SearchResources(ctx context.Context, tenantID, query string, limit int, types map[string]struct{}) ([]searchResult, error) {
	if err := r.requireTableName(); err != nil {
		return nil, err
	}
	normalized := normalizeSearchText(query)
	if len(normalized) < minSearchQueryLength {
		return []searchResult{}, nil
	}
	if limit <= 0 {
		limit = defaultSearchLimit
	}
	if limit > maxSearchLimit {
		limit = maxSearchLimit
	}
	out, err := r.client.Query(ctx, &sharedaws.DynamoDBQueryInput{
		TableName:              sharedaws.String(r.tableName),
		KeyConditionExpression: sharedaws.String("PK = :pk AND begins_with(SK, :prefix)"),
		ExpressionAttributeValues: map[string]sharedaws.AttributeValue{
			":pk":     &sharedaws.AttributeValueMemberS{Value: dynamodbschema.TenantPK(tenantID)},
			":prefix": &sharedaws.AttributeValueMemberS{Value: dynamodbschema.SearchIndexSKPrefix(normalized)},
		},
		Limit: sharedaws.Int32(searchReadLimit),
	})
	if err != nil {
		return nil, err
	}
	byResource := map[string]dynamodbrecord.SearchIndexRecord{}
	for _, item := range out.Items {
		var record dynamodbrecord.SearchIndexRecord
		if err := sharedaws.UnmarshalMap(item, &record); err != nil {
			return nil, err
		}
		if record.EntityType != dynamodbschema.EntitySearchIndex {
			continue
		}
		if len(types) > 0 {
			if _, ok := types[record.ResourceType]; !ok {
				continue
			}
		}
		key := record.ResourceType + "\x00" + record.ServiceID + "\x00" + record.ResourceID
		current, ok := byResource[key]
		if !ok || compareSearchRecord(record, current) < 0 {
			byResource[key] = record
		}
	}
	records := make([]dynamodbrecord.SearchIndexRecord, 0, len(byResource))
	for _, record := range byResource {
		records = append(records, record)
	}
	sort.Slice(records, func(i, j int) bool { return compareSearchRecord(records[i], records[j]) < 0 })
	if len(records) > limit {
		records = records[:limit]
	}
	results := make([]searchResult, 0, len(records))
	for _, record := range records {
		results = append(results, searchResult{Type: record.ResourceType, ID: record.ResourceID, ServiceID: record.ServiceID, Label: record.Label, Description: record.Description, Href: record.Href, IconKey: record.IconKey, MatchText: record.MatchText})
	}
	return results, nil
}

// BackfillSearchIndex rebuilds search-index records for existing tenant resources.
// New writes maintain search records inline; this helper is reserved for one-off
// migrations when resources predate index maintenance.
func (r *dynamoMonitorRepository) BackfillSearchIndex(ctx context.Context, tenantID string) error {
	services, err := r.ListServices(ctx, tenantID)
	if err != nil {
		return err
	}
	for _, service := range services {
		if err := r.replaceSearchIndex(ctx, service.TenantID, searchResourceService, service.ServiceID, "", buildServiceSearchRecords(service)); err != nil {
			return err
		}
		monitors, err := r.ListMonitors(ctx, tenantID, service.ServiceID)
		if err != nil {
			return err
		}
		for _, monitor := range monitors {
			if err := r.replaceSearchIndex(ctx, monitor.TenantID, searchResourceMonitor, monitor.MonitorID, monitor.ServiceID, buildMonitorSearchRecords(monitor, service.Name)); err != nil {
				return err
			}
		}
	}
	policies, err := r.ListEscalationPolicies(ctx, tenantID)
	if err != nil {
		return err
	}
	for _, policy := range policies {
		if err := r.replaceSearchIndex(ctx, policy.TenantID, searchResourcePolicy, policy.PolicyID, "", buildPolicySearchRecords(policy)); err != nil {
			return err
		}
	}
	channels, err := r.ListNotificationChannels(ctx, tenantID)
	if err != nil {
		return err
	}
	for _, channel := range channels {
		if err := r.replaceSearchIndex(ctx, channel.TenantID, searchResourceChannel, channel.ChannelID, "", buildChannelSearchRecords(channel)); err != nil {
			return err
		}
	}
	return nil
}

func (r *dynamoMonitorRepository) replaceSearchIndex(ctx context.Context, tenantID, resourceType, resourceID, serviceID string, records []dynamodbrecord.SearchIndexRecord) error {
	deleteKeys, err := r.searchIndexDeleteKeys(ctx, tenantID, resourceType, resourceID, serviceID)
	if err != nil {
		return err
	}
	for _, key := range deleteKeys {
		if _, err := r.client.DeleteItem(ctx, &sharedaws.DynamoDBDeleteItemInput{TableName: sharedaws.String(r.tableName), Key: map[string]sharedaws.AttributeValue{"PK": &sharedaws.AttributeValueMemberS{Value: key.PK}, "SK": &sharedaws.AttributeValueMemberS{Value: key.SK}}}); err != nil {
			return err
		}
	}
	for _, record := range records {
		item, err := sharedaws.MarshalMap(record)
		if err != nil {
			return err
		}
		if _, err := r.client.PutItem(ctx, &sharedaws.DynamoDBPutItemInput{TableName: sharedaws.String(r.tableName), Item: item}); err != nil {
			return err
		}
	}
	return nil
}

func (r *dynamoMonitorRepository) deleteSearchIndex(ctx context.Context, tenantID, resourceType, resourceID, serviceID string) error {
	deleteKeys, err := r.searchIndexDeleteKeys(ctx, tenantID, resourceType, resourceID, serviceID)
	if err != nil {
		return err
	}
	for _, key := range deleteKeys {
		if _, err := r.client.DeleteItem(ctx, &sharedaws.DynamoDBDeleteItemInput{TableName: sharedaws.String(r.tableName), Key: map[string]sharedaws.AttributeValue{"PK": &sharedaws.AttributeValueMemberS{Value: key.PK}, "SK": &sharedaws.AttributeValueMemberS{Value: key.SK}}}); err != nil {
			return err
		}
	}
	return nil
}

func (r *dynamoMonitorRepository) searchIndexDeleteKeys(ctx context.Context, tenantID, resourceType, resourceID, serviceID string) ([]ddbKey, error) {
	items, err := r.queryPartition(ctx, dynamodbschema.TenantPK(tenantID), "SEARCH#")
	if err != nil {
		return nil, err
	}
	keys := make([]ddbKey, 0)
	for _, item := range items {
		var record dynamodbrecord.SearchIndexRecord
		if err := sharedaws.UnmarshalMap(item, &record); err != nil {
			return nil, err
		}
		if record.EntityType != dynamodbschema.EntitySearchIndex || record.ResourceType != resourceType || !strings.EqualFold(record.ResourceID, resourceID) {
			continue
		}
		if serviceID != "" && !strings.EqualFold(record.ServiceID, serviceID) {
			continue
		}
		keys = append(keys, ddbKey{PK: record.PK, SK: record.SK})
	}
	return keys, nil
}

func buildServiceSearchRecords(service monitorconfig.Service) []dynamodbrecord.SearchIndexRecord {
	description := fmt.Sprintf("Service · %s · %d monitors", service.RollupStatus, service.MonitorCount)
	text := strings.Join([]string{service.Name, service.ServiceID, service.Description, string(service.ServiceCategory), string(service.LifecycleState), service.RollupStatus}, " ")
	return buildSearchRecords(service.TenantID, searchResourceService, service.ServiceID, "", service.Name, description, "/services/"+service.ServiceID, "service:"+firstNonEmpty(string(service.ServiceCategory), "server"), text, service.UpdatedAt)
}

func buildMonitorSearchRecords(monitor monitorconfig.Monitor, serviceName string) []dynamodbrecord.SearchIndexRecord {
	target := safeURLDisplay("")
	if monitor.HTTP != nil {
		target = safeURLDisplay(monitor.HTTP.Target)
	}
	description := fmt.Sprintf("Monitor · %s · %s", firstNonEmpty(serviceName, monitor.ServiceID), target)
	state := "disabled"
	if monitor.Enabled {
		state = "enabled"
	}
	text := strings.Join([]string{monitor.Name, monitor.MonitorID, serviceName, monitor.ServiceID, target, string(monitor.Type), state}, " ")
	return buildSearchRecords(monitor.TenantID, searchResourceMonitor, monitor.MonitorID, monitor.ServiceID, monitor.Name, description, "/services/"+monitor.ServiceID+"/monitors/"+monitor.MonitorID, "monitor:"+string(monitor.Type), text, "")
}

func buildPolicySearchRecords(policy escalation.EscalationPolicy) []dynamodbrecord.SearchIndexRecord {
	businessSteps := len(policy.BusinessHoursPath.Steps)
	offHoursSteps := len(policy.OffHoursPath.Steps)
	description := fmt.Sprintf("Notification route · %d business-hours steps · %d off-hours steps", businessSteps, offHoursSteps)
	channelIDs := make([]string, 0, businessSteps+offHoursSteps)
	for _, path := range []escalation.EscalationPath{policy.BusinessHoursPath, policy.OffHoursPath} {
		for _, step := range path.Steps {
			channelIDs = append(channelIDs, step.ChannelID)
		}
	}
	text := strings.Join([]string{policy.Name, policy.PolicyID, policy.Description, strings.Join(channelIDs, " "), description}, " ")
	return buildSearchRecords(policy.TenantID, searchResourcePolicy, policy.PolicyID, "", policy.Name, description, "/policies/"+policy.PolicyID, "policy", text, policy.UpdatedAt)
}

func buildChannelSearchRecords(channel escalation.NotificationChannel) []dynamodbrecord.SearchIndexRecord {
	target := safeChannelTarget(channel)
	description := fmt.Sprintf("Channel · %s · %s", channel.Type, target)
	text := strings.Join([]string{channel.Name, channel.ChannelID, string(channel.Type), target}, " ")
	return buildSearchRecords(channel.TenantID, searchResourceChannel, channel.ChannelID, "", channel.Name, description, "/integrations/channels/"+channel.ChannelID, "channel:"+string(channel.Type), text, channel.UpdatedAt)
}

func buildSearchRecords(tenantID, resourceType, resourceID, serviceID, label, description, href, iconKey, text, updatedAt string) []dynamodbrecord.SearchIndexRecord {
	labelText := normalizeSearchText(label)
	matchText := normalizeSearchText(text)
	prefixes := map[string]int{}
	addPrefixes(prefixes, labelText, 1, 24)
	for _, token := range strings.Fields(labelText) {
		addPrefixes(prefixes, token, 2, 16)
	}
	for _, token := range strings.Fields(matchText) {
		addPrefixes(prefixes, token, 4, 16)
	}
	records := make([]dynamodbrecord.SearchIndexRecord, 0, len(prefixes))
	for prefix, rank := range prefixes {
		records = append(records, dynamodbrecord.NewSearchIndexRecord(tenantID, prefix, resourceType, resourceID, serviceID, strings.TrimSpace(label), strings.TrimSpace(description), href, iconKey, matchText, rank, updatedAt))
	}
	sort.Slice(records, func(i, j int) bool {
		if records[i].Rank != records[j].Rank {
			return records[i].Rank < records[j].Rank
		}
		if len(records[i].SK) != len(records[j].SK) {
			return len(records[i].SK) < len(records[j].SK)
		}
		return records[i].SK < records[j].SK
	})
	if len(records) > 24 {
		records = records[:24]
	}
	sort.Slice(records, func(i, j int) bool { return records[i].SK < records[j].SK })
	return records
}

func addPrefixes(prefixes map[string]int, text string, rank, maxLen int) {
	if len(text) < minSearchQueryLength {
		return
	}
	if len(text) > maxLen {
		text = text[:maxLen]
	}
	for i := minSearchQueryLength; i <= len(text); i++ {
		prefix := strings.TrimSpace(text[:i])
		if len(prefix) < minSearchQueryLength {
			continue
		}
		if current, ok := prefixes[prefix]; !ok || rank < current {
			prefixes[prefix] = rank
		}
	}
}

func normalizeSearchText(value string) string {
	var b strings.Builder
	lastSpace := true
	for _, r := range strings.ToLower(strings.TrimSpace(value)) {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			b.WriteRune(r)
			lastSpace = false
			continue
		}
		if !lastSpace {
			b.WriteByte(' ')
			lastSpace = true
		}
	}
	return strings.TrimSpace(b.String())
}

func safeURLDisplay(raw string) string {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return "target unavailable"
	}
	parsed, err := url.Parse(trimmed)
	if err != nil || parsed.Host == "" {
		return trimmed
	}
	return parsed.Host + parsed.EscapedPath()
}

func safeChannelTarget(channel escalation.NotificationChannel) string {
	target := strings.TrimSpace(channel.Target)
	if target == "" {
		return "target configured"
	}
	switch channel.Type {
	case escalation.ChannelTypeWebhook:
		return safeURLDisplay(target)
	case escalation.ChannelTypeEmail:
		parts := strings.Split(target, "@")
		if len(parts) == 2 {
			return "@" + parts[1]
		}
	case escalation.ChannelTypeSMS:
		if len(target) > 4 {
			return "..." + target[len(target)-4:]
		}
	case escalation.ChannelTypePagerDuty, escalation.ChannelTypeTelegram:
		if len(target) > 6 {
			return target[:3] + "..." + target[len(target)-3:]
		}
	}
	return target
}

func compareSearchRecord(a, b dynamodbrecord.SearchIndexRecord) int {
	if a.Rank != b.Rank {
		return a.Rank - b.Rank
	}
	if priority := resourceTypePriority(a.ResourceType) - resourceTypePriority(b.ResourceType); priority != 0 {
		return priority
	}
	return strings.Compare(strings.ToLower(a.Label), strings.ToLower(b.Label))
}

func resourceTypePriority(resourceType string) int {
	switch resourceType {
	case searchResourceService:
		return 1
	case searchResourceMonitor:
		return 2
	case searchResourcePolicy:
		return 3
	case searchResourceChannel:
		return 4
	default:
		return 9
	}
}

func parseSearchTypes(raw string) (map[string]struct{}, error) {
	if strings.TrimSpace(raw) == "" {
		return nil, nil
	}
	out := map[string]struct{}{}
	for _, part := range strings.Split(raw, ",") {
		value := strings.ToLower(strings.TrimSpace(part))
		switch value {
		case searchResourceService, searchResourceMonitor, searchResourcePolicy, searchResourceChannel:
			out[value] = struct{}{}
		default:
			return nil, sharederrors.New(sharederrors.CodeValidationFailed, map[string]any{"field": "types", "reason": "unsupported resource type"})
		}
	}
	return out, nil
}
