module bolt-monitor/services/api-health

go 1.26.0

require (
	bolt-monitor/shared/api/response v0.0.0
	github.com/aws/aws-lambda-go v1.50.0
)

replace bolt-monitor/shared/api/response => ../../shared/api/response
