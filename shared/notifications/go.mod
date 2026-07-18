module bolt-monitor/shared/notifications

go 1.26.0

require (
	bolt-monitor/shared/dynamodbschema v0.0.0
	bolt-monitor/shared/outboundhttp v0.0.0
	github.com/aws/aws-sdk-go-v2 v1.42.0
	github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue v1.20.13
	github.com/aws/aws-sdk-go-v2/service/dynamodb v1.51.1
)

require (
	github.com/aws/aws-sdk-go-v2/internal/configsources v1.4.29 // indirect
	github.com/aws/aws-sdk-go-v2/internal/endpoints/v2 v2.7.29 // indirect
	github.com/aws/aws-sdk-go-v2/service/dynamodbstreams v1.31.0 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/accept-encoding v1.13.2 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/endpoint-discovery v1.11.10 // indirect
	github.com/aws/smithy-go v1.27.1 // indirect
)

replace bolt-monitor/shared/dynamodbschema => ../dynamodbschema

replace bolt-monitor/shared/outboundhttp => ../outboundhttp
