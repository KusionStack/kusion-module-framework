package workspace

import (
	"testing"

	"github.com/stretchr/testify/assert"
	v1 "kusionstack.io/kusion-api-go/api.kusion.io/v1"
)

func mockValidModuleConfigs() map[string]*v1.ModuleConfig {
	return map[string]*v1.ModuleConfig{
		"mysql": {
			Path:    "ghcr.io/kusionstack/mysql",
			Version: "0.1.0",
			Configs: v1.Configs{
				Default: v1.GenericConfig{
					"type":         "aws",
					"version":      "5.7",
					"instanceType": "db.t3.micro",
				},
				ModulePatcherConfigs: v1.ModulePatcherConfigs{
					"smallClass": {
						GenericConfig: v1.GenericConfig{
							"instanceType": "db.t3.small",
						},
						ProjectSelector: []string{"foo", "bar"},
					},
				},
			},
		},
		"network": {
			Path:    "ghcr.io/kusionstack/network",
			Version: "0.1.0",
			Configs: v1.Configs{
				Default: v1.GenericConfig{
					"type": "aws",
				},
			},
		},
	}
}

func mockGenericConfig() v1.GenericConfig {
	return v1.GenericConfig{
		"int_type_field":    2,
		"string_type_field": "kusion",
		"map_type_field": v1.GenericConfig{
			"k1": "v1",
			"k2": 2,
		},
		"string_map_type_field": v1.GenericConfig{
			"k1": "v1",
			"k2": "v2",
		},
	}
}

func Test_GetProjectModuleConfigs(t *testing.T) {
	testcases := []struct {
		name                   string
		projectName            string
		moduleConfigs          v1.ModuleConfigs
		success                bool
		expectedProjectConfigs map[string]v1.GenericConfig
	}{
		{
			name:          "successfully get project module configs",
			projectName:   "foo",
			moduleConfigs: mockValidModuleConfigs(),
			success:       true,
			expectedProjectConfigs: map[string]v1.GenericConfig{
				"mysql": {
					"type":         "aws",
					"version":      "5.7",
					"instanceType": "db.t3.small",
				},
				"network": {
					"type": "aws",
				},
			},
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			cfg, err := GetProjectModuleConfigs(tc.moduleConfigs, tc.projectName)
			assert.Equal(t, tc.success, err == nil)
			assert.Equal(t, tc.expectedProjectConfigs, cfg)
		})
	}
}

func Test_GetProjectModuleConfig(t *testing.T) {
	testcases := []struct {
		name                  string
		success               bool
		projectName           string
		moduleConfig          *v1.ModuleConfig
		expectedProjectConfig v1.GenericConfig
	}{
		{
			name:         "successfully get default project module config",
			projectName:  "baz",
			moduleConfig: mockValidModuleConfigs()["mysql"],
			success:      true,
			expectedProjectConfig: v1.GenericConfig{
				"type":         "aws",
				"version":      "5.7",
				"instanceType": "db.t3.micro",
			},
		},
		{
			name:         "successfully get override project module config",
			projectName:  "foo",
			moduleConfig: mockValidModuleConfigs()["mysql"],
			success:      true,
			expectedProjectConfig: v1.GenericConfig{
				"type":         "aws",
				"version":      "5.7",
				"instanceType": "db.t3.small",
			},
		},
		{
			name:                  "failed to get config empty project name",
			projectName:           "",
			moduleConfig:          mockValidModuleConfigs()["mysql"],
			success:               false,
			expectedProjectConfig: nil,
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			cfg, err := GetProjectModuleConfig(tc.moduleConfig, tc.projectName)
			assert.Equal(t, tc.success, err == nil)
			assert.Equal(t, tc.expectedProjectConfig, cfg)
		})
	}
}

func Test_GetIntFieldFromGenericConfig(t *testing.T) {
	r2 := int32(2)

	testcases := []struct {
		name          string
		key           string
		success       bool
		expectedValue *int32
	}{
		{
			name:          "successfully get int type field",
			key:           "int_type_field",
			success:       true,
			expectedValue: &r2,
		},
		{
			name:          "get not exist field",
			key:           "not_exist",
			success:       true,
			expectedValue: nil,
		},
		{
			name:          "get field failed not int type",
			key:           "string_type_field",
			success:       false,
			expectedValue: nil,
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			value, err := GetInt32PointerFromGenericConfig(mockGenericConfig(), tc.key)
			assert.Equal(t, tc.success, err == nil)
			assert.Equal(t, tc.expectedValue, value)
		})
	}
}

func Test_GetStringFieldFromGenericConfig(t *testing.T) {
	testcases := []struct {
		name          string
		key           string
		success       bool
		expectedValue string
	}{
		{
			name:          "successfully get string type field",
			key:           "string_type_field",
			success:       true,
			expectedValue: "kusion",
		},
		{
			name:          "get not exist field",
			key:           "not_exist",
			success:       true,
			expectedValue: "",
		},
		{
			name:          "get field failed not string type",
			key:           "int_type_field",
			success:       false,
			expectedValue: "",
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			value, err := GetStringFromGenericConfig(mockGenericConfig(), tc.key)
			assert.Equal(t, tc.success, err == nil)
			assert.Equal(t, tc.expectedValue, value)
		})
	}
}

func Test_GetMapFieldFromGenericConfig(t *testing.T) {
	testcases := []struct {
		name          string
		key           string
		success       bool
		expectedValue map[string]any
	}{
		{
			name:    "successfully get map type field",
			key:     "map_type_field",
			success: true,
			expectedValue: map[string]any{
				"k1": "v1",
				"k2": 2,
			},
		},
		{
			name:          "get not exist field",
			key:           "not_exist",
			success:       true,
			expectedValue: nil,
		},
		{
			name:          "get field failed not map type",
			key:           "int_type_field",
			success:       false,
			expectedValue: nil,
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			value, err := GetMapFromGenericConfig(mockGenericConfig(), tc.key)
			assert.Equal(t, tc.success, err == nil)
			assert.Equal(t, tc.expectedValue, value)
		})
	}
}

func Test_GetStringMapFieldFromGenericConfig(t *testing.T) {
	testcases := []struct {
		name          string
		key           string
		success       bool
		expectedValue map[string]string
	}{
		{
			name:    "successfully get string map type field",
			key:     "string_map_type_field",
			success: true,
			expectedValue: map[string]string{
				"k1": "v1",
				"k2": "v2",
			},
		},
		{
			name:          "get field failed map key not string",
			key:           "map_type_field",
			success:       false,
			expectedValue: nil,
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			value, err := GetStringMapFromGenericConfig(mockGenericConfig(), tc.key)
			assert.Equal(t, tc.success, err == nil)
			assert.Equal(t, tc.expectedValue, value)
		})
	}
}
