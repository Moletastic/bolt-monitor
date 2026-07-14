module bolt-monitor/services/check-runtime

go 1.26.0

require (
	bolt-monitor/shared/aws v0.0.0
	bolt-monitor/shared/checkexecution v0.0.0
	bolt-monitor/shared/dynamodbrecord v0.0.0
	bolt-monitor/shared/dynamodbschema v0.0.0
	bolt-monitor/shared/monitorconfig v0.0.0
	bolt-monitor/shared/resultstatus v0.0.0
	github.com/aws/aws-lambda-go v1.50.0
	github.com/oklog/ulid/v2 v2.1.1
)

require (
	bolt-monitor/shared/escalation v0.0.0 // indirect
	github.com/aws/aws-sdk-go-v2 v1.42.0 // indirect
	github.com/aws/aws-sdk-go-v2/config v1.32.0 // indirect
	github.com/aws/aws-sdk-go-v2/credentials v1.19.0 // indirect
	github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue v1.20.13 // indirect
	github.com/aws/aws-sdk-go-v2/feature/ec2/imds v1.18.14 // indirect
	github.com/aws/aws-sdk-go-v2/internal/configsources v1.4.29 // indirect
	github.com/aws/aws-sdk-go-v2/internal/endpoints/v2 v2.7.29 // indirect
	github.com/aws/aws-sdk-go-v2/internal/ini v1.8.4 // indirect
	github.com/aws/aws-sdk-go-v2/internal/v4a v1.4.30 // indirect
	github.com/aws/aws-sdk-go-v2/service/dynamodb v1.51.1 // indirect
	github.com/aws/aws-sdk-go-v2/service/dynamodbstreams v1.31.0 // indirect
	github.com/aws/aws-sdk-go-v2/service/eventbridge v1.46.6 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/accept-encoding v1.13.3 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/endpoint-discovery v1.11.10 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/presigned-url v1.13.14 // indirect
	github.com/aws/aws-sdk-go-v2/service/signin v1.0.1 // indirect
	github.com/aws/aws-sdk-go-v2/service/sqs v1.44.0 // indirect
	github.com/aws/aws-sdk-go-v2/service/sso v1.30.4 // indirect
	github.com/aws/aws-sdk-go-v2/service/ssooidc v1.35.8 // indirect
	github.com/aws/aws-sdk-go-v2/service/sts v1.41.1 // indirect
	github.com/aws/smithy-go v1.27.1 // indirect
	github.com/stretchr/testify v1.8.4 // indirect
)

replace bolt-monitor/shared/checkexecution => ../../shared/checkexecution

replace bolt-monitor/shared/dynamodbschema => ../../shared/dynamodbschema

replace bolt-monitor/shared/dynamodbrecord => ../../shared/dynamodbrecord

replace bolt-monitor/shared/aws => ../../shared/aws

replace bolt-monitor/shared/dynamodb => ../../shared/dynamodb

replace bolt-monitor/shared/escalation => ../../shared/escalation

replace bolt-monitor/shared/monitorconfig => ../../shared/monitorconfig

replace bolt-monitor/shared/resultstatus => ../../shared/resultstatus
