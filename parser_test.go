package dsninjector

import (
	"fmt"
	"github.com/stretchr/testify/require"
	"os"
	"path"
	"testing"
)

func TestMarshal(t *testing.T) {
	t.Skip("not implemented")
}

func TestUnmarshal(t *testing.T) {
	t.Skip("not implemented")
}

func TestInitEnvFrom(t *testing.T) {
	key01 := "DSN_INJECTOR_TEST_KEY_01"
	key02 := "DSN_INJECTOR_TEST_KEY_02"
	key03 := "DSN_INJECTOR_TEST_KEY_03"

	require.Equal(t, "", os.Getenv(key01))
	require.Equal(t, "", os.Getenv(key02))
	require.Equal(t, "", os.Getenv(key03))

	p1 := path.Join(t.TempDir(), "dsn_injector_test.1.env")
	p2 := path.Join(t.TempDir(), "dsn_injector_test.2.env")
	p3 := path.Join(t.TempDir(), "dsn_injector_test.3.env")

	require.NoError(t, os.WriteFile(p1, []byte(fmt.Sprintf("%s=TEST_VALUE_01", key01)), 0666))
	require.NoError(t, os.WriteFile(p2, []byte(fmt.Sprintf("%s=TEST_VALUE_02", key02)), 0666))
	require.NoError(t, os.WriteFile(p3, []byte(fmt.Sprintf("%s=TEST_VALUE_03", key03)), 0666))

	require.NoError(t, InitEnvFrom(p1, p2, p3))

	require.Equal(t, "TEST_VALUE_01", os.Getenv(key01))
	require.Equal(t, "TEST_VALUE_02", os.Getenv(key02))
	require.Equal(t, "TEST_VALUE_03", os.Getenv(key03))
}

func TestParse(t *testing.T) {
	successfulTestCases := []struct {
		src string
		dsm *DataSourceMapper
	}{
		{
			src: "postgres://user:password@localhost:5432/dbname",
			dsm: &DataSourceMapper{keyDriverName: "postgres", keyHostName: "localhost", keyPortName: "5432", keyLoginName: "user", keyPasswordName: "password", keyDatabaseName: "dbname"},
		},
		{
			src: "mysql://:@localhost/dbname",
			dsm: &DataSourceMapper{keyDriverName: "mysql", keyHostName: "localhost", keyPortName: "", keyLoginName: "", keyPasswordName: "", keyDatabaseName: "dbname"},
		},
		{
			src: "postgres://user:password@localhost:5432/dbname?sslmode=disable&timeout=30",
			dsm: &DataSourceMapper{keyDriverName: "postgres", keyHostName: "localhost", keyPortName: "5432", keyLoginName: "user", keyPasswordName: "password", keyDatabaseName: "dbname", "sslmode": "disable", "timeout": "30"},
		},
		{
			src: "postgres://user:pass@localhost/dbname",
			dsm: &DataSourceMapper{keyDriverName: "postgres", keyHostName: "localhost", keyPortName: "", keyLoginName: "user", keyPasswordName: "pass", keyDatabaseName: "dbname"},
		},
		{
			src: "pg://user:pass@localhost/dbname?sslmode=disable",
			dsm: &DataSourceMapper{keyDriverName: "pg", keyHostName: "localhost", keyPortName: "", keyLoginName: "user", keyPasswordName: "pass", keyDatabaseName: "dbname", "sslmode": "disable"},
		},
		{
			src: "mysql://user:pass@localhost/dbname",
			dsm: &DataSourceMapper{keyDriverName: "mysql", keyHostName: "localhost", keyPortName: "", keyLoginName: "user", keyPasswordName: "pass", keyDatabaseName: "dbname"},
		},
		{
			src: "mysql:/var/run/mysqld/mysqld.sock",
			dsm: &DataSourceMapper{keyDriverName: "mysql", keyHostName: "", keyPortName: "", keyLoginName: "", keyPasswordName: "", keyDatabaseName: "/var/run/mysqld/mysqld.sock"},
		},
		{
			src: "sqlserver://user:pass@remote-host.com/dbname",
			dsm: &DataSourceMapper{keyDriverName: "sqlserver", keyHostName: "remote-host.com", keyPortName: "", keyLoginName: "user", keyPasswordName: "pass", keyDatabaseName: "dbname"},
		},
		{
			src: "mssql://user:pass@remote-host.com/instance/dbname",
			dsm: &DataSourceMapper{keyDriverName: "mssql", keyHostName: "remote-host.com", keyPortName: "", keyLoginName: "user", keyPasswordName: "pass", keyDatabaseName: "instance/dbname"},
		},
		{
			src: "ms://user:pass@remote-host.com:port/instance/dbname?keepAlive=10",
			dsm: &DataSourceMapper{keyDriverName: "ms", keyHostName: "remote-host.com", keyPortName: "port", keyLoginName: "user", keyPasswordName: "pass", keyDatabaseName: "instance/dbname", "keepAlive": "10"},
		},
		{
			src: "oracle://user:pass@somehost.com/sid",
			dsm: &DataSourceMapper{keyDriverName: "oracle", keyHostName: "somehost.com", keyPortName: "", keyLoginName: "user", keyPasswordName: "pass", keyDatabaseName: "sid"},
		},
		{
			src: "sap://user:pass@localhost/dbname",
			dsm: &DataSourceMapper{keyDriverName: "sap", keyHostName: "localhost", keyPortName: "", keyLoginName: "user", keyPasswordName: "pass", keyDatabaseName: "dbname"},
		},
		{
			src: "sqlite:/path/to/file.db",
			dsm: &DataSourceMapper{keyDriverName: "sqlite", keyHostName: "", keyPortName: "", keyLoginName: "", keyPasswordName: "", keyDatabaseName: "/path/to/file.db"},
		},
		{
			src: "file:myfile.sqlite3?loc=auto",
			dsm: &DataSourceMapper{keyDriverName: "file", keyHostName: "", keyPortName: "", keyLoginName: "", keyPasswordName: "", keyDatabaseName: "/myfile.sqlite3", "loc": "auto"},
		},
		{
			src: "odbc+postgres://user:pass@localhost:port/dbname?option1=",
			dsm: &DataSourceMapper{keyDriverName: "odbc+postgres", keyHostName: "localhost", keyPortName: "port", keyLoginName: "user", keyPasswordName: "pass", keyDatabaseName: "dbname", "option1": ""},
		},
		{
			src: "https://localhost:8080/dbname?option1=1",
			dsm: &DataSourceMapper{keyDriverName: "https", keyHostName: "localhost", keyPortName: "8080", keyLoginName: "", keyPasswordName: "", keyDatabaseName: "dbname", "option1": "1"},
		},
	}

	t.Run("successful", func(t *testing.T) {
		for _, testcase := range successfulTestCases {
			t.Run(testcase.src, func(t *testing.T) {
				got, err := Parse(testcase.src)
				require.NoError(t, err, testcase.src)
				require.Equal(t, testcase.dsm, got, testcase.src)
			})
		}
	})
}
