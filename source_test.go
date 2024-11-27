package dsninjector

import (
	"encoding/base64"
	"github.com/stretchr/testify/require"
	"testing"
)

var _ DataSource = &DataSourceMapper{}

func TestDataSourceMapper_Addr(t *testing.T) {
	testcases := []struct {
		name        string
		mapper      DataSourceMapper
		defaultPort []int
		expected    string
	}{
		{
			name:     "With hostname and port",
			mapper:   DataSourceMapper{"hostname": "localhost", "port": "5432"},
			expected: "localhost:5432",
		},
		{
			name:        "With hostname and default port",
			mapper:      DataSourceMapper{"hostname": "localhost"},
			defaultPort: []int{3306},
			expected:    "localhost:3306",
		},
		{
			name:     "With empty port and without default port",
			mapper:   DataSourceMapper{"hostname": "localhost"},
			expected: "localhost",
		},
		{
			name:     "Without any information",
			mapper:   DataSourceMapper{},
			expected: "",
		},
	}

	for _, test := range testcases {
		t.Run(test.name, func(t *testing.T) {
			require.Equal(t, test.expected, test.mapper.Addr(test.defaultPort...))
		})
	}
}

func TestDataSourceMapper_AuthBasicBase64(t *testing.T) {
	testcases := []struct {
		name     string
		mapper   DataSourceMapper
		expected string
	}{
		{
			name:     "Basic Authentication",
			mapper:   DataSourceMapper{"login": "user", "password": "pass"},
			expected: base64.StdEncoding.EncodeToString([]byte("user:pass")),
		},
		{
			name:     "Empty credentials",
			mapper:   DataSourceMapper{"login": "", "password": ""},
			expected: base64.StdEncoding.EncodeToString([]byte(":")),
		},
	}

	for _, test := range testcases {
		t.Run(test.name, func(t *testing.T) {
			require.Equal(t, test.expected, test.mapper.AuthBasicBase64())
		})
	}
}

func TestDataSourceMapper_Driver(t *testing.T) {
	testcases := []struct {
		name     string
		mapper   DataSourceMapper
		expected string
	}{
		{
			name:     "With driver",
			mapper:   DataSourceMapper{"driver": "postgres"},
			expected: "postgres",
		},
		{
			name:     "Without driver",
			mapper:   DataSourceMapper{"driver": ""},
			expected: "",
		},
	}

	for _, test := range testcases {
		t.Run(test.name, func(t *testing.T) {
			require.Equal(t, test.expected, test.mapper.Driver())
		})
	}
}

func TestDataSourceMapper_SetDriver(t *testing.T) {
	t.Skip("not implemented")
}

func TestDataSourceMapper_Host(t *testing.T) {
	tests := []struct {
		name     string
		mapper   DataSourceMapper
		expected string
	}{
		{
			name:     "With hostname",
			mapper:   DataSourceMapper{"hostname": "localhost"},
			expected: "localhost",
		},
		{
			name:     "Without hostname",
			mapper:   DataSourceMapper{"hostname": ""},
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.Equal(t, tt.expected, tt.mapper.Host())
		})
	}
}

func TestDataSourceMapper_SetHost(t *testing.T) {
	t.Skip("not implemented")
}

func TestDataSourceMapper_Port(t *testing.T) {
	tests := []struct {
		name     string
		mapper   DataSourceMapper
		expected int
	}{
		{
			name:     "With port",
			mapper:   DataSourceMapper{"port": "3306"},
			expected: 3306,
		},
		{
			name:     "Without port",
			mapper:   DataSourceMapper{"port": ""},
			expected: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.Equal(t, tt.expected, tt.mapper.Port())
		})
	}
}

func TestDataSourceMapper_SetPort(t *testing.T) {
	t.Skip("not implemented")
}

func TestDataSourceMapper_Login(t *testing.T) {
	tests := []struct {
		name     string
		mapper   DataSourceMapper
		expected string
	}{
		{
			name:     "With login",
			mapper:   DataSourceMapper{"login": "login"},
			expected: "login",
		},
		{
			name:     "Without login",
			mapper:   DataSourceMapper{"login": ""},
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.Equal(t, tt.expected, tt.mapper.Login())
		})
	}
}

func TestDataSourceMapper_SetLogin(t *testing.T) {
	t.Skip("not implemented")
}

func TestDataSourceMapper_Password(t *testing.T) {
	tests := []struct {
		name     string
		mapper   DataSourceMapper
		expected string
	}{
		{
			name:     "With password",
			mapper:   DataSourceMapper{"password": "password"},
			expected: "password",
		},
		{
			name:     "Without password",
			mapper:   DataSourceMapper{"password": ""},
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.Equal(t, tt.expected, tt.mapper.Password())
		})
	}
}

func TestDataSourceMapper_SetPassword(t *testing.T) {
	t.Skip("not implemented")
}

func TestDataSourceMapper_Database(t *testing.T) {
	tests := []struct {
		name     string
		mapper   DataSourceMapper
		expected string
	}{
		{
			name:     "With database",
			mapper:   DataSourceMapper{"database": "database"},
			expected: "database",
		},
		{
			name:     "Without database",
			mapper:   DataSourceMapper{"database": ""},
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.Equal(t, tt.expected, tt.mapper.Database())
		})
	}
}

func TestDataSourceMapper_SetDatabase(t *testing.T) {
	t.Skip("not implemented")
}

func TestDataSourceMapper_OptionsNames(t *testing.T) {
	t.Skip("not implemented")
}

func TestDataSourceMapper_Option(t *testing.T) {
	t.Skip("not implemented")
}

func TestDataSourceMapper_SetOption(t *testing.T) {
	t.Skip("not implemented")
}
