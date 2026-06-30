package aws

import "testing"

type testRecord struct {
	PK   string `dynamodbav:"PK"`
	SK   string `dynamodbav:"SK"`
	Name string `dynamodbav:"Name"`
}

func TestDynamoDBTranslationHelpersRoundTripDomainRecord(t *testing.T) {
	record := testRecord{PK: "SERVICE#DEFAULT#AUTH", SK: "META", Name: "Auth"}

	item, err := MarshalMap(record)
	if err != nil {
		t.Fatalf("MarshalMap returned error: %v", err)
	}
	if item["PK"].(*AttributeValueMemberS).Value != record.PK {
		t.Fatalf("PK attribute = %v, want %q", item["PK"], record.PK)
	}

	var roundTrip testRecord
	if err := UnmarshalMap(item, &roundTrip); err != nil {
		t.Fatalf("UnmarshalMap returned error: %v", err)
	}
	if roundTrip != record {
		t.Fatalf("roundTrip = %+v, want %+v", roundTrip, record)
	}
}
