package builder

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCsvToMap(t *testing.T) {
	d := []string{
		"\"tolerations=key=foo,value=bar;key=foo2,value=bar2\",replicas=1",
		"namespace=default",
	}
	r, err := csvToMap(d)

	require.NoError(t, err)

	require.Contains(t, r, "tolerations")
	require.Equal(t, r["tolerations"], "key=foo,value=bar;key=foo2,value=bar2")

	require.Contains(t, r, "replicas")
	require.Equal(t, r["replicas"], "1")

	require.Contains(t, r, "namespace")
	require.Equal(t, r["namespace"], "default")
}

func TestParseBuildkitdFlags(t *testing.T) {
	testCases := []struct {
		name       string
		flags      string
		driver     string
		driverOpts map[string]string
		expected   []string
		wantErr    bool
	}{
		{
			"docker-container no flags",
			"",
			"docker-container",
			nil,
			[]string{
				"--allow-insecure-entitlement=network.host",
			},
			false,
		},
		{
			"kubernetes no flags",
			"",
			"kubernetes",
			nil,
			[]string{
				"--allow-insecure-entitlement=network.host",
			},
			false,
		},
		{
			"remote no flags",
			"",
			"remote",
			nil,
			nil,
			false,
		},
		{
			"docker-container with insecure flag",
			"--allow-insecure-entitlement=security.insecure",
			"docker-container",
			nil,
			[]string{
				"--allow-insecure-entitlement=security.insecure",
			},
			false,
		},
		{
			"docker-container with insecure and host flag",
			"--allow-insecure-entitlement=network.host --allow-insecure-entitlement=security.insecure",
			"docker-container",
			nil,
			[]string{
				"--allow-insecure-entitlement=network.host",
				"--allow-insecure-entitlement=security.insecure",
			},
			false,
		},
		{
			"docker-container with network host opt",
			"",
			"docker-container",
			map[string]string{"network": "host"},
			[]string{
				"--allow-insecure-entitlement=network.host",
			},
			false,
		},
		{
			"docker-container with host flag and network host opt",
			"--allow-insecure-entitlement=network.host",
			"docker-container",
			map[string]string{"network": "host"},
			[]string{
				"--allow-insecure-entitlement=network.host",
			},
			false,
		},
		{
			"docker-container with insecure, host flag and network host opt",
			"--allow-insecure-entitlement=network.host --allow-insecure-entitlement=security.insecure",
			"docker-container",
			map[string]string{"network": "host"},
			[]string{
				"--allow-insecure-entitlement=network.host",
				"--allow-insecure-entitlement=security.insecure",
			},
			false,
		},
		{
			"error parsing flags",
			"foo'",
			"docker-container",
			nil,
			nil,
			true,
		},
	}
	for _, tt := range testCases {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			flags, err := parseBuildkitdFlags(tt.flags, tt.driver, tt.driverOpts)
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.expected, flags)
		})
	}
}
