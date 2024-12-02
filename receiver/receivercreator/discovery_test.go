// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package receivercreator

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/component"
	"go.uber.org/zap"
	"go.uber.org/zap/zaptest"

	"github.com/open-telemetry/opentelemetry-collector-contrib/extension/observer"
)

func TestK8sHintsBuilderMetrics(t *testing.T) {
	logger := zaptest.NewLogger(t, zaptest.Level(zap.InfoLevel))

	id := component.ID{}
	err := id.UnmarshalText([]byte("redis/pod-2-UID_6379"))
	assert.NoError(t, err)

	config := `
collection_interval: "20s"
timeout: "30s"
username: "username"
password: "changeme"`
	configRedis := `
collection_interval: "20s"
timeout: "130s"
username: "username"
password: "changeme"`

	tests := map[string]struct {
		inputEndpoint    observer.Endpoint
		expectedReceiver receiverTemplate
		ignoreReceivers  []string
		wantError        bool
	}{
		`metrics_pod_level_hints_only`: {
			inputEndpoint: observer.Endpoint{
				ID:     "namespace/pod-2-UID/redis(6379)",
				Target: "1.2.3.4:6379",
				Details: &observer.Port{
					Name: "redis", Pod: observer.Pod{
						Name:      "pod-2",
						Namespace: "default",
						UID:       "pod-2-UID",
						Labels:    map[string]string{"env": "prod"},
						Annotations: map[string]string{
							otelMetricsHints + "/enabled": "true",
							otelMetricsHints + "/scraper": "redis",
							otelMetricsHints + "/config":  config,
						},
					},
					Port: 6379,
				},
			},
			expectedReceiver: receiverTemplate{
				receiverConfig: receiverConfig{
					id:     id,
					config: userConfigMap{"collection_interval": "20s", "endpoint": "1.2.3.4:6379", "password": "changeme", "timeout": "30s", "username": "username"},
				}, signals: receiverSignals{metrics: true, logs: false, traces: false},
			},
			wantError:       false,
			ignoreReceivers: []string{},
		}, `metrics_pod_level_ignore`: {
			inputEndpoint: observer.Endpoint{
				ID:     "namespace/pod-2-UID/redis(6379)",
				Target: "1.2.3.4:6379",
				Details: &observer.Port{
					Name: "redis", Pod: observer.Pod{
						Name:      "pod-2",
						Namespace: "default",
						UID:       "pod-2-UID",
						Labels:    map[string]string{"env": "prod"},
						Annotations: map[string]string{
							otelMetricsHints + "/enabled": "true",
							otelMetricsHints + "/scraper": "redis",
							otelMetricsHints + "/config":  config,
						},
					},
					Port: 6379,
				},
			},
			expectedReceiver: receiverTemplate{},
			wantError:        false,
			ignoreReceivers:  []string{"redis"},
		}, `metrics_pod_level_hints_only_defaults`: {
			inputEndpoint: observer.Endpoint{
				ID:     "namespace/pod-2-UID/redis(6379)",
				Target: "1.2.3.4:6379",
				Details: &observer.Port{
					Name: "redis", Pod: observer.Pod{
						Name:      "pod-2",
						Namespace: "default",
						UID:       "pod-2-UID",
						Labels:    map[string]string{"env": "prod"},
						Annotations: map[string]string{
							otelMetricsHints + "/enabled": "true",
							otelMetricsHints + "/scraper": "redis",
						},
					},
					Port: 6379,
				},
			},
			expectedReceiver: receiverTemplate{
				receiverConfig: receiverConfig{
					id:     id,
					config: userConfigMap{"endpoint": "1.2.3.4:6379"},
				}, signals: receiverSignals{metrics: true, logs: false, traces: false},
			},
			wantError:       false,
			ignoreReceivers: []string{},
		}, `metrics_container_level_hints`: {
			inputEndpoint: observer.Endpoint{
				ID:     "namespace/pod-2-UID/redis(6379)",
				Target: "1.2.3.4:6379",
				Details: &observer.Port{
					Name: "redis", Pod: observer.Pod{
						Name:      "pod-2",
						Namespace: "default",
						UID:       "pod-2-UID",
						Labels:    map[string]string{"env": "prod"},
						Annotations: map[string]string{
							otelMetricsHints + ".6379/enabled": "true",
							otelMetricsHints + ".6379/scraper": "redis",
							otelMetricsHints + ".6379/config":  config,
						},
					},
					Port: 6379,
				},
			},
			expectedReceiver: receiverTemplate{
				receiverConfig: receiverConfig{
					id:     id,
					config: userConfigMap{"collection_interval": "20s", "endpoint": "1.2.3.4:6379", "password": "changeme", "timeout": "30s", "username": "username"},
				}, signals: receiverSignals{metrics: true, logs: false, traces: false},
			},
			wantError:       false,
			ignoreReceivers: []string{},
		}, `metrics_mix_level_hints`: {
			inputEndpoint: observer.Endpoint{
				ID:     "namespace/pod-2-UID/redis(6379)",
				Target: "1.2.3.4:6379",
				Details: &observer.Port{
					Name: "redis", Pod: observer.Pod{
						Name:      "pod-2",
						Namespace: "default",
						UID:       "pod-2-UID",
						Labels:    map[string]string{"env": "prod"},
						Annotations: map[string]string{
							otelMetricsHints + ".6379/enabled": "true",
							otelMetricsHints + ".6379/scraper": "redis",
							otelMetricsHints + "/config":       config,
							otelMetricsHints + ".6379/config":  configRedis,
						},
					},
					Port: 6379,
				},
			},
			expectedReceiver: receiverTemplate{
				receiverConfig: receiverConfig{
					id:     id,
					config: userConfigMap{"collection_interval": "20s", "endpoint": "1.2.3.4:6379", "password": "changeme", "timeout": "130s", "username": "username"},
				}, signals: receiverSignals{metrics: true, logs: false, traces: false},
			},
			wantError:       false,
			ignoreReceivers: []string{},
		}, `metrics_no_port_error`: {
			inputEndpoint: observer.Endpoint{
				ID:     "namespace/pod-2-UID/redis(6379)",
				Target: "1.2.3.4",
				Details: &observer.Port{
					Name: "redis", Pod: observer.Pod{
						Name:      "pod-2",
						Namespace: "default",
						UID:       "pod-2-UID",
						Labels:    map[string]string{"env": "prod"},
						Annotations: map[string]string{
							otelMetricsHints + "/enabled": "true",
							otelMetricsHints + "/scraper": "redis",
							otelMetricsHints + "/config":  config,
						},
					},
				},
			},
			expectedReceiver: receiverTemplate{},
			wantError:        true,
			ignoreReceivers:  []string{},
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			builder := createK8sHintsBuilder(DiscoveryConfig{Enabled: true, IgnoreReceivers: test.ignoreReceivers}, logger)
			env, err := test.inputEndpoint.Env()
			require.NoError(t, err)
			subreceiverTemplate, err := builder.createReceiverTemplateFromHints(env)
			if subreceiverTemplate == nil {
				require.Equal(t, receiverTemplate{}, test.expectedReceiver)
				return
			}
			if !test.wantError {
				require.NoError(t, err)
				require.Equal(t, subreceiverTemplate.receiverConfig.config, test.expectedReceiver.receiverConfig.config)
				require.Equal(t, subreceiverTemplate.signals, test.expectedReceiver.signals)
				require.Equal(t, subreceiverTemplate.id, test.expectedReceiver.id)
			} else {
				require.Error(t, err)
			}
		})
	}
}

func TestGetConfFromAnnotations(t *testing.T) {
	config := `
endpoint: "0.0.0.0:8080"
collection_interval: "20s"
initial_delay: "20s"
read_buffer_size: "10"
nested_example:
  foo: bar`
	configNoEndpoint := `
collection_interval: "20s"
initial_delay: "20s"
read_buffer_size: "10"
nested_example:
  foo: bar`
	tests := map[string]struct {
		hintsAnn        map[string]string
		expectedConf    userConfigMap
		defaultEndpoint string
		scopeSuffix     string
		expectError     bool
	}{
		"simple_annotation_case": {
			hintsAnn: map[string]string{
				"io.opentelemetry.discovery.metrics/enabled": "true",
				"io.opentelemetry.discovery.metrics/config":  config,
			}, expectedConf: userConfigMap{
				"collection_interval": "20s",
				"endpoint":            "0.0.0.0:8080",
				"initial_delay":       "20s",
				"read_buffer_size":    "10",
				"nested_example":      userConfigMap{"foo": "bar"},
			}, defaultEndpoint: "0.0.0.0:8080",
			scopeSuffix: "",
		}, "simple_annotation_case_default_endpoint": {
			hintsAnn: map[string]string{
				"io.opentelemetry.discovery.metrics/enabled": "true",
				"io.opentelemetry.discovery.metrics/config":  configNoEndpoint,
			}, expectedConf: userConfigMap{
				"collection_interval": "20s",
				"endpoint":            "1.1.1.1:8080",
				"initial_delay":       "20s",
				"read_buffer_size":    "10",
				"nested_example":      userConfigMap{"foo": "bar"},
			}, defaultEndpoint: "1.1.1.1:8080",
			scopeSuffix: "",
		}, "simple_annotation_case_scoped": {
			hintsAnn: map[string]string{
				"io.opentelemetry.discovery.metrics.8080/enabled": "true",
				"io.opentelemetry.discovery.metrics.8080/config":  config,
			}, expectedConf: userConfigMap{
				"collection_interval": "20s",
				"endpoint":            "0.0.0.0:8080",
				"initial_delay":       "20s",
				"read_buffer_size":    "10",
				"nested_example":      userConfigMap{"foo": "bar"},
			}, defaultEndpoint: "0.0.0.0:8080",
			scopeSuffix: "8080",
		}, "simple_annotation_case_with_invalid_endpoint": {
			hintsAnn: map[string]string{
				"io.opentelemetry.discovery.metrics/enabled": "true",
				"io.opentelemetry.discovery.metrics/config":  config,
			}, expectedConf: userConfigMap{},
			defaultEndpoint: "1.2.3.4:8080",
			scopeSuffix:     "",
			expectError:     true,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			conf, err := getScraperConfFromAnnotations(test.hintsAnn, test.defaultEndpoint, test.scopeSuffix, zaptest.NewLogger(t, zaptest.Level(zap.InfoLevel)))
			if test.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(
					t,
					test.expectedConf,
					conf)
			}
		})
	}
}

func TestDiscoveryMetricsEnabled(t *testing.T) {
	config := `
endpoint: "0.0.0.0:8080"`
	tests := map[string]struct {
		hintsAnn    map[string]string
		expected    bool
		scopeSuffix string
	}{
		"test_enabled": {
			hintsAnn: map[string]string{
				"io.opentelemetry.discovery.metrics/config":  config,
				"io.opentelemetry.discovery.metrics/enabled": "true",
			},
			expected:    true,
			scopeSuffix: "",
		}, "test_disabled": {
			hintsAnn: map[string]string{
				"io.opentelemetry.discovery.metrics/config":  config,
				"io.opentelemetry.discovery.metrics/enabled": "false",
			},
			expected:    false,
			scopeSuffix: "",
		}, "test_enabled_scope": {
			hintsAnn: map[string]string{
				"io.opentelemetry.discovery.metrics/config":       config,
				"io.opentelemetry.discovery.metrics.8080/enabled": "true",
			},
			expected:    true,
			scopeSuffix: "8080",
		}, "test_disabled_scoped": {
			hintsAnn: map[string]string{
				"io.opentelemetry.discovery.metrics/config":       config,
				"io.opentelemetry.discovery.metrics.8080/enabled": "false",
			},
			expected:    false,
			scopeSuffix: "8080",
		},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			assert.Equal(
				t,
				test.expected,
				discoveryMetricsEnabled(test.hintsAnn, otelMetricsHints, test.scopeSuffix),
			)
		})
	}
}

func TestValidateEndpoint(t *testing.T) {
	tests := map[string]struct {
		endpoint        string
		defaultEndpoint string
		expectError     bool
	}{
		"test_valid": {
			endpoint:        "http://1.2.3.4:8080/stats",
			defaultEndpoint: "1.2.3.4:8080",
			expectError:     false,
		},
		"test_invalid": {
			endpoint:        "http://0.0.0.0:8080/some?foo=1.2.3.4:8080",
			defaultEndpoint: "1.2.3.4:8080",
			expectError:     true,
		},
		"test_valid_no_scheme": {
			endpoint:        "1.2.3.4:8080/stats",
			defaultEndpoint: "1.2.3.4:8080",
			expectError:     false,
		},
		"test_valid_no_scheme_no_path": {
			endpoint:        "1.2.3.4:8080",
			defaultEndpoint: "1.2.3.4:8080",
			expectError:     false,
		},
		"test_valid_no_scheme_dynamic": {
			endpoint:        "`endpoint`/stats",
			defaultEndpoint: "1.2.3.4:8080",
			expectError:     false,
		},
		"test_valid_dynamic": {
			endpoint:        "http://`endpoint`/stats",
			defaultEndpoint: "1.2.3.4:8080",
			expectError:     false,
		},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			err := validateEndpoint(test.endpoint, test.defaultEndpoint)
			if test.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
