module bolt-monitor/shared/monitorconfig

go 1.26.0

require (
	bolt-monitor/shared/domainvalues v0.0.0
	bolt-monitor/shared/errors v0.0.0
	bolt-monitor/shared/escalation v0.0.0
	bolt-monitor/shared/outboundhttp v0.0.0
	bolt-monitor/shared/rules v0.0.0
)

replace bolt-monitor/shared/domainvalues => ../domainvalues

replace bolt-monitor/shared/errors => ../errors

replace bolt-monitor/shared/escalation => ../escalation

replace bolt-monitor/shared/outboundhttp => ../outboundhttp

replace bolt-monitor/shared/rules => ../rules
