// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package kubeadm

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUpdateDefaults(t *testing.T) {
	cfg := CreateDefaultConfig()
	assert.Equal(t, defaultConfigMapName, cfg.configMapName)
	assert.Equal(t, defaultConfigMapNamespace, cfg.configMapNamespace)
}
