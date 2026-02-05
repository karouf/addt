package otel

import (
	"os"
	"testing"
)

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()

	if cfg.Enabled != false {
		t.Errorf("Expected Enabled=false, got %v", cfg.Enabled)
	}
	if cfg.Endpoint != "http://host.docker.internal:4318" {
		t.Errorf("Expected Endpoint=http://host.docker.internal:4318, got %v", cfg.Endpoint)
	}
	if cfg.Protocol != "http/protobuf" {
		t.Errorf("Expected Protocol=http/protobuf, got %v", cfg.Protocol)
	}
	if cfg.ServiceName != "addt" {
		t.Errorf("Expected ServiceName=addt, got %v", cfg.ServiceName)
	}
}

func TestApplySettings(t *testing.T) {
	cfg := DefaultConfig()

	enabled := true
	endpoint := "http://otel.example.com:4317"
	protocol := "grpc"
	serviceName := "my-service"
	headers := "key=value"

	settings := &Settings{
		Enabled:     &enabled,
		Endpoint:    &endpoint,
		Protocol:    &protocol,
		ServiceName: &serviceName,
		Headers:     &headers,
	}

	applySettings(&cfg, settings)

	if cfg.Enabled != true {
		t.Errorf("Expected Enabled=true, got %v", cfg.Enabled)
	}
	if cfg.Endpoint != endpoint {
		t.Errorf("Expected Endpoint=%s, got %s", endpoint, cfg.Endpoint)
	}
	if cfg.Protocol != protocol {
		t.Errorf("Expected Protocol=%s, got %s", protocol, cfg.Protocol)
	}
	if cfg.ServiceName != serviceName {
		t.Errorf("Expected ServiceName=%s, got %s", serviceName, cfg.ServiceName)
	}
	if cfg.Headers != headers {
		t.Errorf("Expected Headers=%s, got %s", headers, cfg.Headers)
	}
}

func TestApplyEnvOverrides(t *testing.T) {
	// Save and restore environment
	envVars := []string{
		"ADDT_OTEL_ENABLED",
		"ADDT_OTEL_ENDPOINT",
		"ADDT_OTEL_PROTOCOL",
		"ADDT_OTEL_SERVICE_NAME",
		"ADDT_OTEL_HEADERS",
	}
	saved := make(map[string]string)
	for _, key := range envVars {
		saved[key] = os.Getenv(key)
	}
	defer func() {
		for key, val := range saved {
			if val == "" {
				os.Unsetenv(key)
			} else {
				os.Setenv(key, val)
			}
		}
	}()

	// Set test values
	os.Setenv("ADDT_OTEL_ENABLED", "true")
	os.Setenv("ADDT_OTEL_ENDPOINT", "http://env-endpoint:4317")
	os.Setenv("ADDT_OTEL_PROTOCOL", "grpc")
	os.Setenv("ADDT_OTEL_SERVICE_NAME", "env-service")
	os.Setenv("ADDT_OTEL_HEADERS", "auth=token123")

	cfg := DefaultConfig()
	applyEnvOverrides(&cfg)

	if cfg.Enabled != true {
		t.Errorf("Expected Enabled=true from env, got %v", cfg.Enabled)
	}
	if cfg.Endpoint != "http://env-endpoint:4317" {
		t.Errorf("Expected Endpoint from env, got %s", cfg.Endpoint)
	}
	if cfg.Protocol != "grpc" {
		t.Errorf("Expected Protocol=grpc from env, got %s", cfg.Protocol)
	}
	if cfg.ServiceName != "env-service" {
		t.Errorf("Expected ServiceName from env, got %s", cfg.ServiceName)
	}
	if cfg.Headers != "auth=token123" {
		t.Errorf("Expected Headers from env, got %s", cfg.Headers)
	}
}

func TestGetEnvVars(t *testing.T) {
	// Test when disabled
	cfg := Config{Enabled: false}
	env := GetEnvVars(cfg)
	if env != nil {
		t.Errorf("Expected nil env vars when disabled, got %v", env)
	}

	// Test when enabled
	cfg = Config{
		Enabled:     true,
		Endpoint:    "http://otel:4318",
		Protocol:    "http/protobuf",
		ServiceName: "test-service",
		Headers:     "key=value",
	}
	env = GetEnvVars(cfg)

	if env["OTEL_EXPORTER_OTLP_ENDPOINT"] != cfg.Endpoint {
		t.Errorf("Expected OTEL_EXPORTER_OTLP_ENDPOINT=%s, got %s", cfg.Endpoint, env["OTEL_EXPORTER_OTLP_ENDPOINT"])
	}
	if env["OTEL_EXPORTER_OTLP_PROTOCOL"] != cfg.Protocol {
		t.Errorf("Expected OTEL_EXPORTER_OTLP_PROTOCOL=%s, got %s", cfg.Protocol, env["OTEL_EXPORTER_OTLP_PROTOCOL"])
	}
	if env["OTEL_SERVICE_NAME"] != cfg.ServiceName {
		t.Errorf("Expected OTEL_SERVICE_NAME=%s, got %s", cfg.ServiceName, env["OTEL_SERVICE_NAME"])
	}
	if env["OTEL_EXPORTER_OTLP_HEADERS"] != cfg.Headers {
		t.Errorf("Expected OTEL_EXPORTER_OTLP_HEADERS=%s, got %s", cfg.Headers, env["OTEL_EXPORTER_OTLP_HEADERS"])
	}
	if env["CLAUDE_CODE_ENABLE_TELEMETRY"] != "1" {
		t.Errorf("Expected CLAUDE_CODE_ENABLE_TELEMETRY=1, got %s", env["CLAUDE_CODE_ENABLE_TELEMETRY"])
	}

	// Test without headers
	cfg.Headers = ""
	env = GetEnvVars(cfg)
	if _, ok := env["OTEL_EXPORTER_OTLP_HEADERS"]; ok {
		t.Error("Expected no OTEL_EXPORTER_OTLP_HEADERS when empty")
	}
}

func TestLoadConfig(t *testing.T) {
	// Test with nil settings
	cfg := LoadConfig(nil, nil)
	defaults := DefaultConfig()
	if cfg.Enabled != defaults.Enabled {
		t.Errorf("Expected default Enabled, got %v", cfg.Enabled)
	}

	// Test project settings override global
	globalEnabled := false
	projectEnabled := true
	globalEndpoint := "http://global:4318"
	projectEndpoint := "http://project:4318"

	global := &Settings{Enabled: &globalEnabled, Endpoint: &globalEndpoint}
	project := &Settings{Enabled: &projectEnabled, Endpoint: &projectEndpoint}

	cfg = LoadConfig(global, project)
	if cfg.Enabled != true {
		t.Errorf("Expected project setting to override global, got Enabled=%v", cfg.Enabled)
	}
	if cfg.Endpoint != projectEndpoint {
		t.Errorf("Expected project endpoint, got %s", cfg.Endpoint)
	}
}
