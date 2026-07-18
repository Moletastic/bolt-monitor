package aws

import "context"

// PrimaryKey is a typed DynamoDB primary-key value. Constructors and helpers
// in this package use it so callers cannot accidentally mix up PK/SK order
// when assembling exact-key reads.
type PrimaryKey struct {
	PK string
	SK string
}

// AttributeMap returns the canonical {"PK": "...", "SK": "..."} map used by
// GetItem and DeleteItem input parameters.
func (k PrimaryKey) AttributeMap() map[string]AttributeValue {
	return map[string]AttributeValue{
		"PK": &AttributeValueMemberS{Value: k.PK},
		"SK": &AttributeValueMemberS{Value: k.SK},
	}
}

// NewPrimaryKey constructs a typed primary key from PK and SK strings. Inputs
// are passed through unchanged because the caller is the canonical source of
// key shape.
func NewPrimaryKey(pk, sk string) PrimaryKey {
	return PrimaryKey{PK: pk, SK: sk}
}

// PageOptions carries the explicit pagination inputs for a primary-index
// query. Limit is the maximum number of items per page. Forward controls the
// sort direction. Cursor carries the LastEvaluatedKey returned by a previous
// page; nil means "start from the beginning of the partition".
type PageOptions struct {
	Limit   int32
	Forward bool
	Cursor  map[string]AttributeValue
}

// Page is the explicit result of a primary-index read. Items contains the
// decoded attribute maps for the page. NextKey is the DynamoDB continuation
// state: nil means the partition was fully consumed by this page; non-nil
// means the caller must decide whether to continue or report incomplete
// evidence.
type Page struct {
	Items   []map[string]AttributeValue
	NextKey map[string]AttributeValue
}

// HasMore reports whether the page returned continuation state.
func (p Page) HasMore() bool {
	return len(p.NextKey) > 0
}

// Incomplete reports whether the page returned no continuation state but the
// caller issued a bounded read (Limit > 0). Helpers do not silently iterate;
// callers use this to surface incompleteness rather than truncate evidence.
func (p Page) Incomplete(limit int32) bool {
	if limit <= 0 {
		return false
	}
	count := int64(len(p.Items))
	return count >= int64(limit) && !p.HasMore()
}

// GetByPrimaryKey performs an exact-key GetItem through the canonical facade.
// The boolean return reports whether an item was present.
func GetByPrimaryKey(ctx context.Context, api DynamoDBAPI, tableName string, key PrimaryKey) (map[string]AttributeValue, bool, error) {
	out, err := api.GetItem(ctx, &DynamoDBGetItemInput{
		TableName: String(tableName),
		Key:       key.AttributeMap(),
	})
	if err != nil {
		return nil, false, err
	}
	if len(out.Item) == 0 {
		return nil, false, nil
	}
	return out.Item, true, nil
}

// QueryPrimaryPrefixPage performs a primary-index Query constrained by an
// exact PK and an optional SK prefix. The returned page always exposes
// continuation state; callers must choose between bounded traversal or
// explicit incomplete reporting.
func QueryPrimaryPrefixPage(ctx context.Context, api DynamoDBAPI, tableName, pk, prefix string, opts PageOptions) (Page, error) {
	input := &DynamoDBQueryInput{
		TableName:                 String(tableName),
		ExpressionAttributeValues: map[string]AttributeValue{},
	}
	if prefix == "" {
		input.KeyConditionExpression = String("PK = :pk")
		input.ExpressionAttributeValues[":pk"] = &AttributeValueMemberS{Value: pk}
	} else {
		input.KeyConditionExpression = String("PK = :pk AND begins_with(SK, :prefix)")
		input.ExpressionAttributeValues[":pk"] = &AttributeValueMemberS{Value: pk}
		input.ExpressionAttributeValues[":prefix"] = &AttributeValueMemberS{Value: prefix}
	}
	if opts.Limit > 0 {
		input.Limit = Int32(opts.Limit)
	}
	if !opts.Forward {
		input.ScanIndexForward = Bool(false)
	}
	if len(opts.Cursor) > 0 {
		input.ExclusiveStartKey = opts.Cursor
	}

	out, err := api.Query(ctx, input)
	if err != nil {
		return Page{}, err
	}
	return Page{Items: out.Items, NextKey: out.LastEvaluatedKey}, nil
}
