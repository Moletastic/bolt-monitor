package dynamodbrecord

import (
	"time"

	"bolt-monitor/shared/checkexecution"
	"bolt-monitor/shared/dynamodbschema"
)

const (
	ExecutionMarkerPublication = "publication"
	ExecutionMarkerRecovery    = "recovery"
)

// ExecutionMarkerRecord is a query-only recovery index. Canonical work stays
// authoritative, so stale marker reads are validated with a point read.
type ExecutionMarkerRecord struct {
	PK         string `dynamodbav:"PK"`
	SK         string `dynamodbav:"SK"`
	EntityType string `dynamodbav:"EntityType"`
	TTL        int64  `dynamodbav:"TTL"`
	TenantID   string `dynamodbav:"TenantID"`
	RunID      string `dynamodbav:"RunID"`
	Kind       string `dynamodbav:"Kind"`
	Bucket     string `dynamodbav:"Bucket"`
	Shard      string `dynamodbav:"Shard"`
}

func NewExecutionMarkerRecord(work checkexecution.ExecutionWork, kind, bucket, shard string) ExecutionMarkerRecord {
	acceptedAt := work.AcceptedAt
	if acceptedAt.IsZero() {
		acceptedAt = work.RequestedAt
	}
	ttl := acceptedAt.UTC().Add(checkexecution.DefaultExecutionWorkRetentionDays * 24 * time.Hour).Unix()
	item := dynamodbschema.ExecutionMarkerItem(work.TenantID, kind, bucket, shard, work.RunID, ttl)
	return ExecutionMarkerRecord{PK: item.PK, SK: item.SK, EntityType: item.EntityType, TTL: ttl, TenantID: item.TenantID, RunID: item.RunID, Kind: kind, Bucket: bucket, Shard: shard}
}
