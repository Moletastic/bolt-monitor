package dynamodbrecord

import (
	"fmt"
	"strings"

	"bolt-monitor/shared/dynamodbschema"
)

const DispatchPendingBucketFormat = "2006010215"

type DispatchPendingRecord struct {
	PK         string `dynamodbav:"PK"`
	SK         string `dynamodbav:"SK"`
	EntityType string `dynamodbav:"EntityType"`
	TenantID   string `dynamodbav:"TenantID"`
	RunID      string `dynamodbav:"RunID"`
	Bucket     string `dynamodbav:"Bucket"`
	Shard      string `dynamodbav:"Shard"`
}

func NewDispatchPendingRecord(tenantID, eventID, bucket, shard string) DispatchPendingRecord {
	return DispatchPendingRecord{
		PK:         dispatchPendingPK(tenantID, bucket, shard),
		SK:         strings.ToUpper(strings.TrimSpace(eventID)),
		EntityType: dynamodbschema.EntityDispatchPending,
		TenantID:   tenantID,
		RunID:      eventID,
		Bucket:     bucket,
		Shard:      shard,
	}
}

func dispatchPendingPK(tenantID, bucket, shard string) string {
	return fmt.Sprintf("DISPATCH_PENDING#%s#%s#%s", strings.ToUpper(strings.TrimSpace(tenantID)), strings.ToUpper(strings.TrimSpace(bucket)), strings.ToUpper(strings.TrimSpace(shard)))
}
