package dsninjector

import (
	"encoding/base64"
	"fmt"
	"strconv"
	"strings"
)

type DataSource interface {
	Driver() string
	Host() string
	Port() int
	Login() string
	Password() string
	Database() string
	OptionsNames() []string
	Option(name string, defaultValue ...string) string
	Addr(defaultPort ...int) string
	AuthBasicBase64() string
}

type DataSourceMapper map[string]string

func (dsm *DataSourceMapper) Addr(defaultPort ...int) string {
	h := dsm.Option(keyHostName)
	p := dsm.Option(keyPortName)
	if len(p) > 0 {
		p = fmt.Sprintf(":%s", p)
	} else if len(defaultPort) > 0 {
		p = fmt.Sprintf(":%d", defaultPort[0])
	}
	return h + p
}

func (dsm *DataSourceMapper) AuthBasicBase64() string {
	l := dsm.Option(keyLoginName)
	p := dsm.Option(keyPasswordName)
	basic := fmt.Sprintf("%s:%s", l, p)
	return base64.StdEncoding.EncodeToString([]byte(basic))
}

func (dsm *DataSourceMapper) Driver() string     { return dsm.Option(keyDriverName) }
func (dsm *DataSourceMapper) SetDriver(v string) { dsm.SetOption(keyDriverName, v) }

func (dsm *DataSourceMapper) Host() string     { return dsm.Option(keyHostName) }
func (dsm *DataSourceMapper) SetHost(v string) { dsm.SetOption(keyHostName, v) }

func (dsm *DataSourceMapper) Port() int {
	v, _ := strconv.ParseInt(dsm.Option(keyPortName), 10, 64)
	return int(v)
}
func (dsm *DataSourceMapper) SetPort(v int) { dsm.SetOption(keyPortName, fmt.Sprintf("%d", v)) }

func (dsm *DataSourceMapper) Login() string     { return dsm.Option(keyLoginName) }
func (dsm *DataSourceMapper) SetLogin(v string) { dsm.SetOption(keyLoginName, v) }

func (dsm *DataSourceMapper) Password() string     { return dsm.Option(keyPasswordName) }
func (dsm *DataSourceMapper) SetPassword(v string) { dsm.SetOption(keyPasswordName, v) }

func (dsm *DataSourceMapper) Database() string     { return dsm.Option(keyDatabaseName) }
func (dsm *DataSourceMapper) SetDatabase(v string) { dsm.SetOption(keyDatabaseName, v) }

func (dsm *DataSourceMapper) OptionsNames() []string {
	names := make([]string, 0, len(*dsm))
	for k := range *dsm {
		if k == keyDriverName || k == keyHostName || k == keyPortName || k == keyLoginName || k == keyPasswordName || k == keyDatabaseName {
			continue
		}
		names = append(names, k)
	}
	return names
}

func (dsm *DataSourceMapper) Option(name string, defaultValue ...string) string {
	val, exists := (*dsm)[name]
	if !exists {
		val = strings.Join(defaultValue, ";")
	}
	return val
}

func (dsm *DataSourceMapper) SetOption(name string, value string) {
	(*dsm)[name] = value
}

const (
	keyDriverName   = "driver"
	keyHostName     = "hostname"
	keyPortName     = "port"
	keyLoginName    = "login"
	keyPasswordName = "password"
	keyDatabaseName = "database"
)
