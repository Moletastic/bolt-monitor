module bolt-monitor/shared/resultstatus

go 1.26.0

require (
	bolt-monitor/shared/checkexecution v0.0.0
	bolt-monitor/shared/dynamodbschema v0.0.0
)

replace bolt-monitor/shared/checkexecution => ../checkexecution

replace bolt-monitor/shared/dynamodbschema => ../dynamodbschema
