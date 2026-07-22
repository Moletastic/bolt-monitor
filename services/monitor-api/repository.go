package main

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"time"

	sharedaws "bolt-monitor/shared/aws"
	"bolt-monitor/shared/checkexecution"
	"bolt-monitor/shared/domainvalues"
	"bolt-monitor/shared/dynamodbrecord"
	"bolt-monitor/shared/dynamodbschema"
	sharederrors "bolt-monitor/shared/errors"
	"bolt-monitor/shared/escalation"
	"bolt-monitor/shared/monitorconfig"
	"bolt-monitor/shared/resultstatus"
)

var errMissingTableName = sharederrors.New(sharederrors.CodeInternal, nil)
var errIncidentNotActionable = sharederrors.New(sharederrors.CodeIncidentNotActionable, nil)
var errServiceAlreadyExists = sharederrors.New(sharederrors.CodeServiceAlreadyExists, nil)
var errMonitorAlreadyExists = sharederrors.New(sharederrors.CodeMonitorAlreadyExists, nil)

var errCannotDeleteActiveService = sharederrors.New(sharederrors.CodeServiceActive, nil)
var errCannotDeleteLastMonitorFromActiveService = sharederrors.New(sharederrors.CodeLastMonitor, nil)

const (
	entityIncidentRef     = "IncidentRef"
	entitySchedulerConfig = "SchedulerConfig"

	rollupDraft    = "draft"
	rollupArchived = "archived"
	rollupPaused   = "paused"
	rollupUnknown  = "unknown"
	rollupUp       = "up"
	rollupDown     = "down"
	rollupDegraded = "degraded"

	incidentStatusOpen         = "open"
	incidentStatusAcknowledged = "acknowledged"
	incidentStatusResolved     = "resolved"
)

type dynamoAPI = sharedaws.DynamoDBAPI

// dynamoMonitorRepository is the single DynamoDB-backed implementation of
// every narrow repository interface in this package. The handler-facing
// interfaces are split by domain; this struct satisfies them all because
// domain transactions share one connection.
type dynamoMonitorRepository struct {
	client    dynamoAPI
	tableName string
	now       func() time.Time
}

// historyPage is the explicit pagination result exposed by every paginated
// read in this package.
type historyPage[T any] struct {
	Items   []T
	NextKey map[string]sharedaws.AttributeValue
}

func newDynamoMonitorRepository(client dynamoAPI, tableName string) *dynamoMonitorRepository {
	return &dynamoMonitorRepository{client: client, tableName: tableName, now: time.Now}
}

// requireTableName ensures the repository was constructed with a deployment-
// injected table name.
func (r *dynamoMonitorRepository) requireTableName() error {
	if strings.TrimSpace(r.tableName) == "" {
		return errMissingTableName
	}
	return nil
}

// queryPartition is a thin wrapper that hides the empty-Limit case while
// keeping the call sites short. Callers that need explicit pagination must
// use sharedaws.QueryPrimaryPrefixPage directly.
func (r *dynamoMonitorRepository) queryPartition(ctx context.Context, pk, prefix string) ([]map[string]sharedaws.AttributeValue, error) {
	page, err := sharedaws.QueryPrimaryPrefixPage(ctx, r.client, r.tableName, pk, prefix, sharedaws.PageOptions{})
	if err != nil {
		return nil, err
	}
	return page.Items, nil
}

// writeTransaction sends a TransactWriteItems request only when there is at
// least one item. An empty slice is a no-op so callers can compose items
// unconditionally.
func (r *dynamoMonitorRepository) writeTransaction(ctx context.Context, items []sharedaws.TransactWriteItem) error {
	if len(items) == 0 {
		return nil
	}
	_, err := r.client.TransactWriteItems(ctx, &sharedaws.DynamoDBTransactWriteItemsInput{TransactItems: items})
	return err
}

// deleteKeysAndPut walks the supplied delete keys in 25-item chunks because
// DynamoDB TransactWriteItems caps each call at 100 items total (puts + deletes).
func (r *dynamoMonitorRepository) deleteKeysAndPut(ctx context.Context, keys []ddbKey, puts []sharedaws.TransactWriteItem) error {
	if len(puts) > 25 {
		return fmt.Errorf("too many put items for transaction: %d", len(puts))
	}
	if len(keys) == 0 {
		return r.writeTransaction(ctx, puts)
	}
	maxDeletesPerTransaction := 25
	if len(puts) > 0 {
		maxDeletesPerTransaction = 25 - len(puts)
	}
	if maxDeletesPerTransaction <= 0 {
		return fmt.Errorf("too many put items for transaction with deletes: %d", len(puts))
	}
	for start := 0; start < len(keys); start += maxDeletesPerTransaction {
		end := start + maxDeletesPerTransaction
		if end > len(keys) {
			end = len(keys)
		}
		items := make([]sharedaws.TransactWriteItem, 0, end-start+1)
		for _, key := range keys[start:end] {
			items = append(items, sharedaws.TransactWriteItem{Delete: &sharedaws.Delete{
				TableName: sharedaws.String(r.tableName),
				Key: map[string]sharedaws.AttributeValue{
					"PK": &sharedaws.AttributeValueMemberS{Value: key.PK},
					"SK": &sharedaws.AttributeValueMemberS{Value: key.SK},
				},
			}})
		}
		if end == len(keys) {
			items = append(items, puts...)
		}
		if err := r.writeTransaction(ctx, items); err != nil {
			return err
		}
	}
	return nil
}

// marshalPutItems marshals each record into a TransactWriteItems Put so
// callers can compose multi-item transactions.
func marshalPutItems(tableName string, records ...any) ([]sharedaws.TransactWriteItem, error) {
	items := make([]sharedaws.TransactWriteItem, 0, len(records))
	for _, record := range records {
		item, err := sharedaws.MarshalMap(record)
		if err != nil {
			return nil, err
		}
		items = append(items, sharedaws.TransactWriteItem{Put: &sharedaws.Put{TableName: sharedaws.String(tableName), Item: item}})
	}
	return items, nil
}

// ddbKey is the canonical primary-key value used to drive transactional
// delete operations. It matches sharedaws.PrimaryKey in spirit but stays
// local so callers do not accidentally mix it with the facade helper.
type ddbKey struct {
	PK string
	SK string
}

type deleteKeySet struct {
	seen map[string]ddbKey
}

func newDeleteKeySet() *deleteKeySet {
	return &deleteKeySet{seen: map[string]ddbKey{}}
}

func (s *deleteKeySet) add(pk, sk string) {
	if strings.TrimSpace(pk) == "" || strings.TrimSpace(sk) == "" {
		return
	}
	s.seen[pk+"\x00"+sk] = ddbKey{PK: pk, SK: sk}
}

func (s *deleteKeySet) addItems(items []map[string]sharedaws.AttributeValue) {
	for _, item := range items {
		s.addItem(item)
	}
}

func (s *deleteKeySet) addItem(item map[string]sharedaws.AttributeValue) {
	pk, ok1 := item["PK"].(*sharedaws.AttributeValueMemberS)
	sk, ok2 := item["SK"].(*sharedaws.AttributeValueMemberS)
	if ok1 && ok2 {
		s.add(pk.Value, sk.Value)
	}
}

func (s *deleteKeySet) list() []ddbKey {
	out := make([]ddbKey, 0, len(s.seen))
	for _, key := range s.seen {
		out = append(out, key)
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].PK == out[j].PK {
			return out[i].SK < out[j].SK
		}
		return out[i].PK < out[j].PK
	})
	return out
}

// matchesIncidentFilter maps the API filter string to the canonical incident
// status set.
func matchesIncidentFilter(incidentStatus, filter string) bool {
	switch strings.ToLower(filter) {
	case "", "all":
		return true
	case "open":
		return incidentStatus == incidentStatusOpen || incidentStatus == incidentStatusAcknowledged
	case "closed":
		return incidentStatus == incidentStatusResolved
	default:
		return true
	}
}

// deriveServiceRollup aggregates enabled monitor statuses into the rollup
// string persisted on the service status item.
func deriveServiceRollup(lifecycle monitorconfig.ServiceLifecycle, summaries []monitorconfig.MonitorSummary) string {
	enabled := make([]monitorconfig.MonitorSummary, 0)
	for _, summary := range summaries {
		if summary.Enabled {
			enabled = append(enabled, summary)
		}
	}
	if len(enabled) == 0 {
		return rollupPaused
	}
	upCount := 0
	downCount := 0
	unknownCount := 0
	for _, summary := range enabled {
		switch strings.ToLower(strings.TrimSpace(summary.CurrentStatus)) {
		case "up":
			upCount++
		case "down":
			downCount++
		default:
			unknownCount++
		}
	}
	if unknownCount == len(enabled) {
		return rollupUnknown
	}
	if upCount == len(enabled) {
		return rollupUp
	}
	if downCount == len(enabled) {
		return rollupDown
	}
	return rollupDegraded
}

// buildMonitorSummaries materializes the per-monitor summary list combining
// monitor config and current status.
func buildMonitorSummaries(monitors []monitorconfig.Monitor, statuses map[string]resultstatus.MonitorStatus) []monitorconfig.MonitorSummary {
	summaries := make([]monitorconfig.MonitorSummary, 0, len(monitors))
	for _, monitor := range monitors {
		summary := monitorconfig.MonitorSummary{
			TenantID:        monitor.TenantID,
			ServiceID:       monitor.ServiceID,
			MonitorID:       monitor.MonitorID,
			Name:            monitor.Name,
			Type:            monitor.Type,
			Enabled:         monitor.Enabled,
			IntervalSeconds: monitor.IntervalSeconds,
		}
		if status, ok := statuses[monitorStatusMapKey(monitor)]; ok {
			summary.CurrentStatus = strings.ToLower(status.CurrentStatus)
			summary.LastCheckedAt = status.LastCheckedAt.UTC().Format(time.RFC3339)
			summary.LastDurationMs = status.LastDurationMs
			summary.LastError = status.LastError
			summary.UpdatedAt = summary.LastCheckedAt
		}
		summaries = append(summaries, summary)
	}
	sort.Slice(summaries, func(i, j int) bool { return summaries[i].MonitorID < summaries[j].MonitorID })
	return summaries
}

// buildServiceCardMetrics composes the dashboard service-card payload from
// monitors, statuses, and recent run samples.
func buildServiceCardMetrics(monitors []monitorconfig.Monitor, statuses map[string]resultstatus.MonitorStatus, runsByMonitor map[string][]resultstatus.CheckRun) serviceCardMetricsResponse {
	metrics := serviceCardMetricsResponse{
		MonitorCount: len(monitors),
		Trend:        []serviceCardTrendPoint{},
	}
	if len(monitors) == 0 {
		metrics.State = serviceCardMetricStateNoMonitors
		return metrics
	}
	for _, status := range statuses {
		if strings.EqualFold(status.CurrentStatus, string(resultstatus.MonitorStateUp)) {
			metrics.UpMonitorCount++
		}
	}
	successDurations := []int64{}
	for _, monitor := range monitors {
		for _, run := range runsByMonitor[monitor.MonitorID] {
			metrics.SampleCount++
			success := run.Outcome == checkexecution.OutcomeSuccess
			if success {
				metrics.SuccessCount++
				successDurations = append(successDurations, run.DurationMs)
			}
			metrics.Trend = append(metrics.Trend, serviceCardTrendPoint{
				MonitorID:  run.MonitorID,
				StartedAt:  run.StartedAt.UTC().Format(time.RFC3339),
				DurationMs: run.DurationMs,
				Outcome:    string(run.Outcome),
				Success:    success,
			})
		}
	}
	if metrics.SampleCount == 0 {
		metrics.State = serviceCardMetricStateNoData
		return metrics
	}
	metrics.State = serviceCardMetricStateReady
	uptime := float64(metrics.SuccessCount) / float64(metrics.SampleCount) * 100
	metrics.RecentUptimePct = &uptime
	if len(successDurations) > 0 {
		avg := averageInt64(successDurations)
		p99 := percentileNearestRank(successDurations, 99)
		metrics.AvgLatencyMs = &avg
		metrics.P99LatencyMs = &p99
	}
	sort.Slice(metrics.Trend, func(i, j int) bool { return metrics.Trend[i].StartedAt < metrics.Trend[j].StartedAt })
	return metrics
}

func averageInt64(values []int64) int64 {
	if len(values) == 0 {
		return 0
	}
	var total int64
	for _, value := range values {
		total += value
	}
	return total / int64(len(values))
}

func percentileNearestRank(values []int64, percentile int) int64 {
	if len(values) == 0 {
		return 0
	}
	sorted := append([]int64(nil), values...)
	sort.Slice(sorted, func(i, j int) bool { return sorted[i] < sorted[j] })
	index := (percentile*len(sorted)+99)/100 - 1
	if index < 0 {
		index = 0
	}
	if index >= len(sorted) {
		index = len(sorted) - 1
	}
	return sorted[index]
}

func countEnabledMonitors(monitors []monitorconfig.Monitor) int {
	count := 0
	for _, monitor := range monitors {
		if monitor.Enabled {
			count++
		}
	}
	return count
}

// monitorStatusMapKey is the canonical "<service>/<monitor>" key used by
// status-map caches. Replacing the previous primitive helper with this method
// locks the format in one place.
func monitorStatusMapKey(monitor monitorconfig.Monitor) string {
	return domainvalues.MustMonitorRef(
		domainvalues.TenantID(monitor.TenantID),
		domainvalues.ServiceID(monitor.ServiceID),
		domainvalues.MonitorID(monitor.MonitorID),
	).StatusMapKey()
}

func mustParseTime(value string) time.Time {
	parsed, _ := time.Parse(time.RFC3339, value)
	return parsed
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return value
		}
	}
	return ""
}

func boolString(v bool) string {
	if v {
		return "true"
	}
	return "false"
}

// narrowInterfaces is a compile-time guard that dynamoMonitorRepository
// satisfies every domain-local interface declared in this package. Failures
// here point directly at the missing method.
var (
	_ MonitorStore    = (*dynamoMonitorRepository)(nil)
	_ SchedulerStore  = (*dynamoMonitorRepository)(nil)
	_ AuditStore      = (*dynamoMonitorRepository)(nil)
	_ EscalationStore = (*dynamoMonitorRepository)(nil)
	_ SearchStore     = (*dynamoMonitorRepository)(nil)
	_ executionStore  = (*dynamoMonitorRepository)(nil)
)

// Domain-local interfaces are declared in their respective files:
//   - MonitorStore        : repository_monitor.go
//   - SchedulerStore      : repository_scheduler.go
//   - AuditStore          : repository_audit.go
//   - EscalationStore     : repository_escalation.go
//   - SearchStore         : repository_search.go (search.go)
//   - executionStore      : repository_monitor.go (RecordExecutionResult)
//
// The shared searchStore interface lives in search.go.

// _ references to imported packages to keep the import set stable across the
// domain files. These aliases are no-ops; they exist so gofmt does not strip
// unused-import errors if a future split removes a local usage.
var (
	_ = dynamodbschema.WorkspaceItem
	_ = dynamodbrecord.NewMonitorItemRecord
	_ = escalation.NotificationChannel{}
)
