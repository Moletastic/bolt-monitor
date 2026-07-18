package aws

import (
	"context"
	"errors"
	"testing"
)

// fakeDynamo is a minimal DynamoDBAPI implementation for exercising the page
// helpers in isolation. It stores items in a map keyed by PK+SK and records
// the last Query input for assertions.
type fakeDynamo struct {
	items       map[string]map[string]AttributeValue
	lastQuery   *DynamoDBQueryInput
	lastGetItem *DynamoDBGetItemInput
	nextKeyEcho map[string]AttributeValue
}

func newFakeDynamo() *fakeDynamo {
	return &fakeDynamo{items: map[string]map[string]AttributeValue{}}
}

func (f *fakeDynamo) addItem(pk, sk string, attributes map[string]AttributeValue) {
	combined := map[string]AttributeValue{
		"PK": &AttributeValueMemberS{Value: pk},
		"SK": &AttributeValueMemberS{Value: sk},
	}
	for key, value := range attributes {
		combined[key] = value
	}
	f.items[pk+"|"+sk] = combined
}

func (f *fakeDynamo) GetItem(_ context.Context, in *DynamoDBGetItemInput) (*DynamoDBGetItemOutput, error) {
	f.lastGetItem = in
	pk := in.Key["PK"].(*AttributeValueMemberS).Value
	sk := in.Key["SK"].(*AttributeValueMemberS).Value
	if item, ok := f.items[pk+"|"+sk]; ok {
		return &DynamoDBGetItemOutput{Item: item}, nil
	}
	return &DynamoDBGetItemOutput{}, nil
}

func (f *fakeDynamo) PutItem(context.Context, *DynamoDBPutItemInput) (*DynamoDBPutItemOutput, error) {
	return &DynamoDBPutItemOutput{}, nil
}

func (f *fakeDynamo) UpdateItem(context.Context, *DynamoDBUpdateItemInput) (*DynamoDBUpdateItemOutput, error) {
	return &DynamoDBUpdateItemOutput{}, nil
}

func (f *fakeDynamo) DeleteItem(context.Context, *DynamoDBDeleteItemInput) (*DynamoDBDeleteItemOutput, error) {
	return &DynamoDBDeleteItemOutput{}, nil
}

func (f *fakeDynamo) Query(_ context.Context, in *DynamoDBQueryInput) (*DynamoDBQueryOutput, error) {
	f.lastQuery = in
	pk := in.ExpressionAttributeValues[":pk"].(*AttributeValueMemberS).Value
	out := &DynamoDBQueryOutput{}
	for key, item := range f.items {
		combined := key
		if len(combined) < len(pk) || combined[:len(pk)] != pk {
			continue
		}
		if prefix, ok := in.ExpressionAttributeValues[":prefix"]; ok {
			want := prefix.(*AttributeValueMemberS).Value
			rest := combined[len(pk)+1:]
			if len(rest) < len(want) || rest[:len(want)] != want {
				continue
			}
		}
		out.Items = append(out.Items, item)
	}
	if f.nextKeyEcho != nil {
		out.LastEvaluatedKey = f.nextKeyEcho
	}
	return out, nil
}

func (f *fakeDynamo) Scan(context.Context, *DynamoDBScanInput) (*DynamoDBScanOutput, error) {
	return &DynamoDBScanOutput{}, nil
}

func (f *fakeDynamo) TransactWriteItems(context.Context, *DynamoDBTransactWriteItemsInput) (*DynamoDBTransactWriteItemsOutput, error) {
	return &DynamoDBTransactWriteItemsOutput{}, nil
}

func TestPrimaryKeyAttributeMapShape(t *testing.T) {
	key := NewPrimaryKey("SERVICE#DEFAULT#AUTH", "META")
	m := key.AttributeMap()
	if m["PK"].(*AttributeValueMemberS).Value != "SERVICE#DEFAULT#AUTH" {
		t.Fatalf("PK = %v", m["PK"])
	}
	if m["SK"].(*AttributeValueMemberS).Value != "META" {
		t.Fatalf("SK = %v", m["SK"])
	}
}

func TestGetByPrimaryKeyReturnsItemAndFound(t *testing.T) {
	api := newFakeDynamo()
	api.addItem("SERVICE#DEFAULT#AUTH", "META", map[string]AttributeValue{
		"Name": &AttributeValueMemberS{Value: "Auth"},
	})

	item, found, err := GetByPrimaryKey(context.Background(), api, "table-name", NewPrimaryKey("SERVICE#DEFAULT#AUTH", "META"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !found {
		t.Fatal("found = false, want true")
	}
	if got := item["Name"].(*AttributeValueMemberS).Value; got != "Auth" {
		t.Fatalf("Name = %q", got)
	}
}

func TestGetByPrimaryKeyMissingReturnsNotFound(t *testing.T) {
	api := newFakeDynamo()
	_, found, err := GetByPrimaryKey(context.Background(), api, "table-name", NewPrimaryKey("SERVICE#DEFAULT#AUTH", "META"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if found {
		t.Fatal("found = true, want false")
	}
}

func TestQueryPrimaryPrefixPageWithPrefixAndOrdering(t *testing.T) {
	api := newFakeDynamo()
	api.addItem("MONITOR#DEFAULT#AUTH#X", "RUN#2026-05-17T22:00:00Z#RUN_A", nil)
	api.addItem("MONITOR#DEFAULT#AUTH#X", "RUN#2026-05-17T22:05:00Z#RUN_B", nil)
	api.addItem("MONITOR#DEFAULT#AUTH#X", "STATUS", nil)

	page, err := QueryPrimaryPrefixPage(context.Background(), api, "table-name", "MONITOR#DEFAULT#AUTH#X", "RUN#", PageOptions{
		Limit:   10,
		Forward: false,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(page.Items) != 2 {
		t.Fatalf("items = %d, want 2", len(page.Items))
	}
	if api.lastQuery.KeyConditionExpression == nil ||
		*api.lastQuery.KeyConditionExpression != "PK = :pk AND begins_with(SK, :prefix)" {
		t.Fatalf("KCE = %v", api.lastQuery.KeyConditionExpression)
	}
	if api.lastQuery.ScanIndexForward == nil || *api.lastQuery.ScanIndexForward != false {
		t.Fatal("ScanIndexForward should be false (reverse)")
	}
}

func TestQueryPrimaryPrefixPageWithoutPrefix(t *testing.T) {
	api := newFakeDynamo()
	api.addItem("TENANT#DEFAULT", "META", nil)
	api.addItem("TENANT#DEFAULT", "OTHER", nil)

	page, err := QueryPrimaryPrefixPage(context.Background(), api, "table-name", "TENANT#DEFAULT", "", PageOptions{Limit: 5})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(page.Items) != 2 {
		t.Fatalf("items = %d, want 2", len(page.Items))
	}
	if api.lastQuery.KeyConditionExpression == nil ||
		*api.lastQuery.KeyConditionExpression != "PK = :pk" {
		t.Fatalf("KCE = %v", api.lastQuery.KeyConditionExpression)
	}
}

func TestQueryPrimaryPrefixPagePropagatesCursorAndLimit(t *testing.T) {
	api := newFakeDynamo()
	cursor := map[string]AttributeValue{
		"PK": &AttributeValueMemberS{Value: "TENANT#DEFAULT"},
		"SK": &AttributeValueMemberS{Value: "SERVICE#AUTH"},
	}
	api.nextKeyEcho = cursor
	api.addItem("TENANT#DEFAULT", "SERVICE#AUTH", nil)

	page, err := QueryPrimaryPrefixPage(context.Background(), api, "table-name", "TENANT#DEFAULT", "SERVICE#", PageOptions{
		Limit:  1,
		Cursor: cursor,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !page.HasMore() {
		t.Fatal("HasMore = false, want true (cursor echoed)")
	}
	if api.lastQuery.ExclusiveStartKey == nil {
		t.Fatal("ExclusiveStartKey not propagated")
	}
	if api.lastQuery.Limit == nil || *api.lastQuery.Limit != 1 {
		t.Fatalf("Limit = %v", api.lastQuery.Limit)
	}
}

func TestPageIncompleteReportsBoundedEvidence(t *testing.T) {
	page := Page{Items: []map[string]AttributeValue{{}, {}, {}}, NextKey: nil}
	if !page.Incomplete(3) {
		t.Fatal("Incomplete(3) should be true when limit reached and no continuation")
	}
	if page.Incomplete(10) {
		t.Fatal("Incomplete(10) should be false when items < limit")
	}
	if page.Incomplete(0) {
		t.Fatal("Incomplete(0) should be false (unbounded read)")
	}
}

// errDynamo returns an error on Query to exercise the helper's error path.
type errDynamo struct{}

func (errDynamo) GetItem(context.Context, *DynamoDBGetItemInput) (*DynamoDBGetItemOutput, error) {
	return nil, errors.New("unused")
}
func (errDynamo) PutItem(context.Context, *DynamoDBPutItemInput) (*DynamoDBPutItemOutput, error) {
	return nil, errors.New("unused")
}
func (errDynamo) UpdateItem(context.Context, *DynamoDBUpdateItemInput) (*DynamoDBUpdateItemOutput, error) {
	return nil, errors.New("unused")
}
func (errDynamo) DeleteItem(context.Context, *DynamoDBDeleteItemInput) (*DynamoDBDeleteItemOutput, error) {
	return nil, errors.New("unused")
}
func (errDynamo) Query(context.Context, *DynamoDBQueryInput) (*DynamoDBQueryOutput, error) {
	return nil, errors.New("query failure")
}
func (errDynamo) Scan(context.Context, *DynamoDBScanInput) (*DynamoDBScanOutput, error) {
	return nil, errors.New("unused")
}
func (errDynamo) TransactWriteItems(context.Context, *DynamoDBTransactWriteItemsInput) (*DynamoDBTransactWriteItemsOutput, error) {
	return nil, errors.New("unused")
}

func TestQueryPrimaryPrefixPagePropagatesError(t *testing.T) {
	if _, err := QueryPrimaryPrefixPage(context.Background(), errDynamo{}, "t", "pk", "", PageOptions{}); err == nil {
		t.Fatal("expected error from underlying Query")
	}
}
