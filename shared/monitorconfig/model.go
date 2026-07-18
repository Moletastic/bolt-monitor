package monitorconfig

import (
	"fmt"
	"net/http"
	"sort"
	"strings"

	sharederrors "bolt-monitor/shared/errors"
	"bolt-monitor/shared/escalation"
	"bolt-monitor/shared/outboundhttp"
	"bolt-monitor/shared/rules"
)

type MonitorType string

const (
	MonitorTypeHTTP MonitorType = "http"
)

type ServiceLifecycle string

const (
	ServiceLifecycleDraft    ServiceLifecycle = "draft"
	ServiceLifecycleActive   ServiceLifecycle = "active"
	ServiceLifecycleArchived ServiceLifecycle = "archived"
)

type ServiceCategory string

const (
	ServiceCategoryServer        ServiceCategory = "server"
	ServiceCategoryDatabase      ServiceCategory = "database"
	ServiceCategoryCache         ServiceCategory = "cache"
	ServiceCategoryHTTP          ServiceCategory = "http"
	ServiceCategoryQueue         ServiceCategory = "queue"
	ServiceCategoryContainer     ServiceCategory = "container"
	ServiceCategoryFunction      ServiceCategory = "function"
	ServiceCategoryWeb           ServiceCategory = "web"
	ServiceCategoryAPI           ServiceCategory = "api"
	ServiceCategoryWorker        ServiceCategory = "worker"
	ServiceCategoryScheduler     ServiceCategory = "scheduler"
	ServiceCategoryStorage       ServiceCategory = "storage"
	ServiceCategorySearch        ServiceCategory = "search"
	ServiceCategoryAuth          ServiceCategory = "auth"
	ServiceCategoryPayments      ServiceCategory = "payments"
	ServiceCategoryAnalytics     ServiceCategory = "analytics"
	ServiceCategoryObservability ServiceCategory = "observability"
	ServiceCategoryAI            ServiceCategory = "ai"
	ServiceCategoryIntegration   ServiceCategory = "integration"
	ServiceCategoryMedia         ServiceCategory = "media"
	ServiceCategoryContent       ServiceCategory = "content"
	ServiceCategoryFinance       ServiceCategory = "finance"
	ServiceCategoryLearning      ServiceCategory = "learning"
	ServiceCategoryGaming        ServiceCategory = "gaming"
	ServiceCategoryCommerce      ServiceCategory = "commerce"
	ServiceCategoryMessaging     ServiceCategory = "messaging"
	ServiceCategorySupport       ServiceCategory = "support"
	ServiceCategoryMarketing     ServiceCategory = "marketing"
	ServiceCategoryAdmin         ServiceCategory = "admin"
	ServiceCategorySecurity      ServiceCategory = "security"
	ServiceCategoryLocation      ServiceCategory = "location"
	ServiceCategorySocial        ServiceCategory = "social"
)

var supportedServiceCategories = map[ServiceCategory]struct{}{
	ServiceCategoryServer:        {},
	ServiceCategoryDatabase:      {},
	ServiceCategoryCache:         {},
	ServiceCategoryHTTP:          {},
	ServiceCategoryQueue:         {},
	ServiceCategoryContainer:     {},
	ServiceCategoryFunction:      {},
	ServiceCategoryWeb:           {},
	ServiceCategoryAPI:           {},
	ServiceCategoryWorker:        {},
	ServiceCategoryScheduler:     {},
	ServiceCategoryStorage:       {},
	ServiceCategorySearch:        {},
	ServiceCategoryAuth:          {},
	ServiceCategoryPayments:      {},
	ServiceCategoryAnalytics:     {},
	ServiceCategoryObservability: {},
	ServiceCategoryAI:            {},
	ServiceCategoryIntegration:   {},
	ServiceCategoryMedia:         {},
	ServiceCategoryContent:       {},
	ServiceCategoryFinance:       {},
	ServiceCategoryLearning:      {},
	ServiceCategoryGaming:        {},
	ServiceCategoryCommerce:      {},
	ServiceCategoryMessaging:     {},
	ServiceCategorySupport:       {},
	ServiceCategoryMarketing:     {},
	ServiceCategoryAdmin:         {},
	ServiceCategorySecurity:      {},
	ServiceCategoryLocation:      {},
	ServiceCategorySocial:        {},
}

var allowedIntervalSeconds = map[int]struct{}{
	60:   {},
	120:  {},
	180:  {},
	300:  {},
	600:  {},
	900:  {},
	1800: {},
	3600: {},
}

type Service struct {
	TenantID           string                          `json:"tenantId"`
	ServiceID          string                          `json:"serviceId"`
	Name               string                          `json:"name"`
	Description        string                          `json:"description,omitempty"`
	LifecycleState     ServiceLifecycle                `json:"lifecycleState"`
	ServiceCategory    ServiceCategory                 `json:"serviceCategory,omitempty"`
	EscalationPolicyID string                          `json:"escalationPolicyId,omitempty"`
	BusinessHours      *escalation.BusinessHoursConfig `json:"businessHours,omitempty"`
	MonitorCount       int                             `json:"monitorCount,omitempty"`
	EnabledCount       int                             `json:"enabledMonitorCount,omitempty"`
	RollupStatus       string                          `json:"rollupStatus,omitempty"`
	CreatedAt          string                          `json:"createdAt,omitempty"`
	UpdatedAt          string                          `json:"updatedAt,omitempty"`
	MonitorSummaries   []MonitorSummary                `json:"monitors,omitempty"`
}

type Monitor struct {
	TenantID          string             `json:"tenantId"`
	ServiceID         string             `json:"serviceId"`
	MonitorID         string             `json:"monitorId"`
	Name              string             `json:"name"`
	Type              MonitorType        `json:"type"`
	IntervalSeconds   int                `json:"intervalSeconds"`
	Enabled           bool               `json:"enabled"`
	FailureThreshold  int                `json:"failureThreshold"`
	RecoveryThreshold int                `json:"recoveryThreshold"`
	HTTP              *HTTPConfiguration `json:"http,omitempty"`
}

type MonitorSummary struct {
	TenantID        string      `json:"tenantId"`
	ServiceID       string      `json:"serviceId"`
	MonitorID       string      `json:"monitorId"`
	Name            string      `json:"name"`
	Type            MonitorType `json:"type"`
	Enabled         bool        `json:"enabled"`
	IntervalSeconds int         `json:"intervalSeconds"`
	CurrentStatus   string      `json:"currentStatus,omitempty"`
	LastCheckedAt   string      `json:"lastCheckedAt,omitempty"`
	LastDurationMs  int64       `json:"lastDurationMs,omitempty"`
	LastError       string      `json:"lastError,omitempty"`
	UpdatedAt       string      `json:"updatedAt,omitempty"`
}

type HTTPConfiguration struct {
	Target               string            `json:"target"`
	Method               string            `json:"method"`
	Headers              map[string]string `json:"headers,omitempty"`
	TimeoutMs            int               `json:"timeoutMs"`
	ExpectedStatusCodes  []int             `json:"expectedStatusCodes,omitempty"`
	ExpectedBodyContains *string           `json:"expectedBodyContains,omitempty"`
}

type CreateServiceRequest struct {
	Name               string                          `json:"name"`
	Description        string                          `json:"description,omitempty"`
	ServiceCategory    ServiceCategory                 `json:"serviceCategory,omitempty"`
	EscalationPolicyID string                          `json:"escalationPolicyId,omitempty"`
	BusinessHours      *escalation.BusinessHoursConfig `json:"businessHours,omitempty"`
}

type CreateMonitorRequest struct {
	Name              string             `json:"name"`
	Type              MonitorType        `json:"type"`
	IntervalSeconds   int                `json:"intervalSeconds"`
	Enabled           bool               `json:"enabled"`
	FailureThreshold  int                `json:"failureThreshold"`
	RecoveryThreshold int                `json:"recoveryThreshold"`
	HTTP              *HTTPConfiguration `json:"http,omitempty"`
}

type Record struct {
	PK                string             `json:"pk"`
	SK                string             `json:"sk"`
	EntityType        string             `json:"entityType"`
	TenantID          string             `json:"tenantId"`
	ServiceID         string             `json:"serviceId"`
	MonitorID         string             `json:"monitorId"`
	Name              string             `json:"name"`
	Type              MonitorType        `json:"type"`
	IntervalSeconds   int                `json:"intervalSeconds"`
	Enabled           bool               `json:"enabled"`
	FailureThreshold  int                `json:"failureThreshold"`
	RecoveryThreshold int                `json:"recoveryThreshold"`
	HTTP              *HTTPConfiguration `json:"http,omitempty"`
}

func SupportedServiceCategories() []ServiceCategory {
	categories := make([]ServiceCategory, 0, len(supportedServiceCategories))
	for cat := range supportedServiceCategories {
		categories = append(categories, cat)
	}
	sort.Slice(categories, func(i, j int) bool { return categories[i] < categories[j] })
	return categories
}

func (r CreateServiceRequest) ToService(tenantID string) (Service, error) {
	service := Service{
		TenantID:           normalizeTenantID(tenantID),
		Name:               strings.TrimSpace(r.Name),
		Description:        strings.TrimSpace(r.Description),
		LifecycleState:     ServiceLifecycleDraft,
		ServiceCategory:    normalizeServiceCategory(r.ServiceCategory),
		EscalationPolicyID: strings.TrimSpace(r.EscalationPolicyID),
		BusinessHours:      cloneBusinessHoursConfig(r.BusinessHours),
	}
	if err := service.Validate(); err != nil {
		return Service{}, err
	}
	return service, nil
}

func (r CreateMonitorRequest) ToMonitor(serviceID, tenantID, monitorID string) (Monitor, error) {
	failureThreshold := r.FailureThreshold
	if failureThreshold < 1 {
		failureThreshold = 1
	}
	recoveryThreshold := r.RecoveryThreshold
	if recoveryThreshold < 1 {
		recoveryThreshold = 1
	}
	monitor := Monitor{
		TenantID:          normalizeTenantID(tenantID),
		ServiceID:         normalizeSlug(serviceID),
		MonitorID:         monitorID,
		Name:              strings.TrimSpace(r.Name),
		Type:              r.Type,
		IntervalSeconds:   r.IntervalSeconds,
		Enabled:           r.Enabled,
		FailureThreshold:  failureThreshold,
		RecoveryThreshold: recoveryThreshold,
		HTTP:              cloneHTTPConfiguration(r.HTTP),
	}
	if err := monitor.Validate(); err != nil {
		return Monitor{}, err
	}
	return monitor, nil
}

func (s Service) Validate() error {
	var builder rules.Builder[Service]
	builder.Add(rules.Field("tenantId", func(service Service) error {
		if strings.TrimSpace(service.TenantID) == "" {
			return validationError("required")
		}
		return nil
	}))
	builder.Add(rules.Field("name", func(service Service) error {
		if strings.TrimSpace(service.Name) == "" {
			return validationError("required")
		}
		return nil
	}))
	builder.Add(rules.Field("serviceCategory", func(service Service) error {
		if service.ServiceCategory == "" {
			return nil
		}
		if _, ok := supportedServiceCategories[service.ServiceCategory]; !ok {
			return validationError(fmt.Sprintf("%q is not supported", service.ServiceCategory))
		}
		return nil
	}))
	return firstJoinedError(builder.Build()(s))
}

func (m Monitor) Validate() error {
	var builder rules.Builder[Monitor]
	builder.Add(rules.Field("tenantId", func(monitor Monitor) error {
		if strings.TrimSpace(monitor.TenantID) == "" {
			return validationError("required")
		}
		return nil
	}))
	builder.Add(rules.Field("serviceId", func(monitor Monitor) error {
		if strings.TrimSpace(monitor.ServiceID) == "" {
			return validationError("required")
		}
		return nil
	}))
	builder.Add(rules.Field("name", func(monitor Monitor) error {
		if strings.TrimSpace(monitor.Name) == "" {
			return validationError("required")
		}
		return nil
	}))
	builder.Add(rules.Field("intervalSeconds", func(monitor Monitor) error {
		if !IsAllowedIntervalSeconds(monitor.IntervalSeconds) {
			return validationError(fmt.Sprintf("must be one of: %s", strings.Trim(strings.Join(strings.Fields(fmt.Sprint(AllowedIntervalSeconds())), ", "), "[]")))
		}
		return nil
	}))
	builder.Add(rules.Field("failureThreshold", func(monitor Monitor) error {
		if monitor.FailureThreshold < 1 {
			return validationError("must be at least 1")
		}
		return nil
	}))
	builder.Add(rules.Field("recoveryThreshold", func(monitor Monitor) error {
		if monitor.RecoveryThreshold < 1 {
			return validationError("must be at least 1")
		}
		return nil
	}))
	builder.Add(func(monitor Monitor) error {
		switch monitor.Type {
		case MonitorTypeHTTP:
			if monitor.HTTP == nil {
				return sharederrors.Wrap(sharederrors.CodeValidationFailed, nil, map[string]any{"field": "http", "reason": fmt.Sprintf("configuration is required for type %q", monitor.Type)})
			}
			return monitor.HTTP.Validate()
		default:
			return sharederrors.Wrap(sharederrors.CodeValidationFailed, nil, map[string]any{"field": "type", "reason": fmt.Sprintf("unsupported monitor type %q", monitor.Type)})
		}
	})
	return firstJoinedError(builder.Build()(m))
}

func validationError(reason string) error {
	return sharederrors.New(sharederrors.CodeValidationFailed, map[string]any{"reason": reason})
}

func firstJoinedError(err error) error {
	if err == nil {
		return nil
	}
	type joined interface{ Unwrap() []error }
	if unwrapped, ok := err.(joined); ok {
		for _, child := range unwrapped.Unwrap() {
			if child != nil {
				return child
			}
		}
	}
	return err
}

func IsAllowedIntervalSeconds(intervalSeconds int) bool {
	_, ok := allowedIntervalSeconds[intervalSeconds]
	return ok
}

func AllowedIntervalSeconds() []int {
	values := make([]int, 0, len(allowedIntervalSeconds))
	for value := range allowedIntervalSeconds {
		values = append(values, value)
	}
	sort.Ints(values)
	return values
}

func (h HTTPConfiguration) Validate() error {
	if strings.TrimSpace(h.Target) == "" {
		return sharederrors.Wrap(sharederrors.CodeValidationFailed, nil, map[string]any{"field": "http.target", "reason": "required"})
	}
	if _, err := outboundhttp.ValidateURL(h.Target); err != nil {
		return sharederrors.Wrap(sharederrors.CodeValidationFailed, nil, map[string]any{"field": "http.target", "reason": outboundhttp.SafeMessage(err)})
	}

	if strings.TrimSpace(h.Method) == "" {
		return sharederrors.Wrap(sharederrors.CodeValidationFailed, nil, map[string]any{"field": "http.method", "reason": "required"})
	}
	if !validHTTPMethod(h.Method) {
		return sharederrors.Wrap(sharederrors.CodeValidationFailed, nil, map[string]any{"field": "http.method", "reason": fmt.Sprintf("%q is not supported", h.Method)})
	}
	if h.TimeoutMs <= 0 {
		return sharederrors.Wrap(sharederrors.CodeValidationFailed, nil, map[string]any{"field": "http.timeoutMs", "reason": "must be greater than 0"})
	}
	if h.TimeoutMs > int(outboundhttp.MaxRequestTimeout.Milliseconds()) {
		return sharederrors.Wrap(sharederrors.CodeValidationFailed, nil, map[string]any{"field": "http.timeoutMs", "reason": "must be no greater than 30000"})
	}

	for _, code := range h.ExpectedStatusCodes {
		if code < 100 || code > 599 {
			return sharederrors.Wrap(sharederrors.CodeValidationFailed, nil, map[string]any{"field": "http.expectedStatusCodes", "reason": fmt.Sprintf("contains invalid status code %d", code)})
		}
	}

	return nil
}

func (m Monitor) ToRecord() (Record, error) {
	if err := m.Validate(); err != nil {
		return Record{}, err
	}
	return Record{
		PK:                "MONITOR#" + strings.ToUpper(m.TenantID) + "#" + strings.ToUpper(m.ServiceID) + "#" + strings.ToUpper(m.MonitorID),
		SK:                "META",
		EntityType:        "Monitor",
		TenantID:          m.TenantID,
		ServiceID:         m.ServiceID,
		MonitorID:         m.MonitorID,
		Name:              m.Name,
		Type:              m.Type,
		IntervalSeconds:   m.IntervalSeconds,
		Enabled:           m.Enabled,
		FailureThreshold:  m.FailureThreshold,
		RecoveryThreshold: m.RecoveryThreshold,
		HTTP:              cloneHTTPConfiguration(m.HTTP),
	}, nil
}

func MonitorFromRecord(record Record) (Monitor, error) {
	monitor := Monitor{
		TenantID:          normalizeTenantID(record.TenantID),
		ServiceID:         normalizeSlug(record.ServiceID),
		MonitorID:         normalizeSlug(record.MonitorID),
		Name:              strings.TrimSpace(record.Name),
		Type:              record.Type,
		IntervalSeconds:   record.IntervalSeconds,
		Enabled:           record.Enabled,
		FailureThreshold:  record.FailureThreshold,
		RecoveryThreshold: record.RecoveryThreshold,
		HTTP:              cloneHTTPConfiguration(record.HTTP),
	}
	if err := monitor.Validate(); err != nil {
		return Monitor{}, err
	}
	return monitor, nil
}

func validHTTPMethod(method string) bool {
	upper := strings.ToUpper(method)
	return upper == http.MethodGet ||
		upper == http.MethodHead ||
		upper == http.MethodPost ||
		upper == http.MethodPut ||
		upper == http.MethodPatch ||
		upper == http.MethodDelete ||
		upper == http.MethodOptions
}

func cloneHTTPConfiguration(input *HTTPConfiguration) *HTTPConfiguration {
	if input == nil {
		return nil
	}
	headers := make(map[string]string, len(input.Headers))
	for key, value := range input.Headers {
		headers[key] = value
	}
	statusCodes := append([]int(nil), input.ExpectedStatusCodes...)
	var expectedBodyContains *string
	if input.ExpectedBodyContains != nil {
		value := *input.ExpectedBodyContains
		expectedBodyContains = &value
	}
	return &HTTPConfiguration{
		Target:               strings.TrimSpace(input.Target),
		Method:               strings.ToUpper(strings.TrimSpace(input.Method)),
		Headers:              headers,
		TimeoutMs:            input.TimeoutMs,
		ExpectedStatusCodes:  statusCodes,
		ExpectedBodyContains: expectedBodyContains,
	}
}

func cloneBusinessHoursConfig(input *escalation.BusinessHoursConfig) *escalation.BusinessHoursConfig {
	if input == nil {
		return nil
	}
	daysOfWeek := append([]int(nil), input.DaysOfWeek...)
	sort.Ints(daysOfWeek)
	return &escalation.BusinessHoursConfig{
		Timezone:   strings.TrimSpace(input.Timezone),
		StartHour:  input.StartHour,
		EndHour:    input.EndHour,
		DaysOfWeek: daysOfWeek,
	}
}

func normalizeTenantID(value string) string {
	return strings.ToUpper(strings.TrimSpace(value))
}

func normalizeSlug(value string) string {
	return strings.ToLower(strings.TrimSpace(value))
}

func normalizeServiceCategory(value ServiceCategory) ServiceCategory {
	return ServiceCategory(strings.ToLower(strings.TrimSpace(string(value))))
}
