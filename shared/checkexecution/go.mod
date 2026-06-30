module bolt-monitor/shared/checkexecution

go 1.26.0

require (
	bolt-monitor/shared/monitorconfig v0.0.0
	bolt-monitor/shared/probelocationcatalog v0.0.0
)

replace bolt-monitor/shared/monitorconfig => ../monitorconfig

replace bolt-monitor/shared/probelocationcatalog => ../probelocationcatalog
