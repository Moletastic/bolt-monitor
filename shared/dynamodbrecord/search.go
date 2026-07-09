package dynamodbrecord

import "bolt-monitor/shared/dynamodbschema"

type SearchIndexRecord struct {
	PK           string `dynamodbav:"PK"`
	SK           string `dynamodbav:"SK"`
	EntityType   string `dynamodbav:"EntityType"`
	TenantID     string `dynamodbav:"TenantID"`
	ResourceType string `dynamodbav:"ResourceType"`
	ResourceID   string `dynamodbav:"ResourceID"`
	ServiceID    string `dynamodbav:"ServiceID,omitempty"`
	Label        string `dynamodbav:"Label"`
	Description  string `dynamodbav:"Description,omitempty"`
	Href         string `dynamodbav:"Href"`
	IconKey      string `dynamodbav:"IconKey"`
	MatchText    string `dynamodbav:"MatchText"`
	Rank         int    `dynamodbav:"Rank"`
	UpdatedAt    string `dynamodbav:"UpdatedAt,omitempty"`
}

func NewSearchIndexRecord(tenantID, prefix, resourceType, resourceID, serviceID, label, description, href, iconKey, matchText string, rank int, updatedAt string) SearchIndexRecord {
	resourceKey := resourceID
	if serviceID != "" {
		resourceKey = serviceID + "#" + resourceID
	}
	item := dynamodbschema.SearchIndexItem(tenantID, prefix, resourceType, resourceKey)
	return SearchIndexRecord{
		PK:           item.PK,
		SK:           item.SK,
		EntityType:   item.EntityType,
		TenantID:     dynamodbschema.NormalizeToken(tenantID),
		ResourceType: resourceType,
		ResourceID:   resourceID,
		ServiceID:    serviceID,
		Label:        label,
		Description:  description,
		Href:         href,
		IconKey:      iconKey,
		MatchText:    matchText,
		Rank:         rank,
		UpdatedAt:    updatedAt,
	}
}
