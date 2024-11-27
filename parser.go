package dsninjector

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"net/url"
	"os"
	"path"
	"regexp"
	"strings"
)

// Marshal returns string representation of DataSource
func Marshal(dsm DataSource) (string, error) {
	args := ""
	if names := dsm.OptionsNames(); len(names) > 0 {
		values := make([]string, 0, len(names))
		for _, n := range names {
			values = append(values, fmt.Sprintf("%s=%s", n, dsm.Option(n)))
		}
		args = "?" + strings.Join(values, "&")
	}

	return fmt.Sprintf("%s://%s@%s/%s%s", dsm.Driver(), dsm.AuthBasicBase64(), dsm.Addr(), dsm.Database(), args), nil
}

// Unmarshal reads environment variable and returns DataSource
func Unmarshal(key string, defaultValue ...string) DataSource {
	env, exists := os.LookupEnv(key)
	if !exists && len(defaultValue) > 0 {
		env = strings.Join(defaultValue, ";")
	}

	env = strings.TrimSpace(env)
	if env == "" {
		return nil
	}

	cfn, err := Parse(env)
	if err != nil {
		return nil
	}

	return cfn
}

// InitEnvFrom reads environment variables from the file and sets them
func InitEnvFrom(filePaths ...string) error {
	if len(filePaths) == 0 {
		filePaths = []string{".env"}
	}

	for _, filePath := range filePaths {
		if filePath == "" {
			continue
		}

		if s, err := os.Stat(filePath); errors.Is(err, os.ErrNotExist) || s.IsDir() {
			continue
		}

		d, err := extractEnvVarName(filePath)
		if err != nil {
			return fmt.Errorf("could not extract environment variables from %s, reason: %w", filePath, err)
		}

		for k, v := range d {
			err = os.Setenv(k, v)
			if err != nil {
				return fmt.Errorf("could not set environment variable %s, reason: %w", k, err)
			}
		}
	}

	return nil
}

// Parse reads the connection string and returns DataSource
func Parse(dns string) (*DataSourceMapper, error) {
	rx, err := regexp.Compile(`(?i)^(?:(?P<driver>[a-z+]+):)?(?://(?:(?P<credentials>[^/@]+)@)?(?P<instance>[^/@]+))?(?:/?(?P<path>[^?]+))?(?:\?(?P<params>.*))?$`)
	if err != nil {
		return nil, err
	}

	// Match the connection string
	matches := rx.FindStringSubmatch(dns)
	if matches == nil {
		return nil, fmt.Errorf("invalid connection string format: %s", dns)
	}

	m := make(map[string]string)

	// Map the named groups to their values
	names := rx.SubexpNames()
	instance := ""
	credentials := ""
	for i, name := range names {
		if name == "" || i >= len(matches) {
			continue
		}

		switch name {
		case "driver":
			m[keyDriverName] = matches[i]
		case "instance":
			instance = matches[i]
			if strings.Contains(instance, ":") {
				p := strings.SplitN(instance, ":", 2)
				m[keyHostName] = p[0]
				m[keyPortName] = p[1]
			} else {
				m[keyHostName] = instance
				m[keyPortName] = ""
			}
		case "credentials":
			credentials = matches[i]
			if strings.Contains(credentials, ":") {
				p := strings.SplitN(credentials, ":", 2)
				m[keyLoginName] = p[0]
				m[keyPasswordName] = p[1]
			} else {
				m[keyLoginName] = ""
				m[keyPasswordName] = credentials
			}
		case "path":
			if instance == "" && credentials == "" {
				m[keyDatabaseName] = path.Join("/", matches[i])
			} else {
				m[keyDatabaseName] = matches[i]
			}
		case "params":
			var q url.Values
			if q, err = url.ParseQuery(matches[i]); err != nil {
				return nil, err
			}
			for k, v := range q {
				m[k] = strings.Join(v, ",")
			}
		}
	}

	ds := DataSourceMapper(m)
	return &ds, nil
}

// extractEnvVarName reads environment variables from the file
func extractEnvVarName(filePath string) (map[string]string, error) {
	f, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer func(closer io.Closer) { err = errors.Join(err, closer.Close()) }(f)

	dataset := make(map[string]string)

	fScanner := bufio.NewScanner(f)
	fScanner.Split(bufio.ScanLines)

	for fScanner.Scan() {
		line := strings.SplitN(fScanner.Text(), "=", 2)
		if len(line) != 2 || len(line[0]) == 0 || len(line[1]) == 0 {
			continue
		}

		key := strings.ToUpper(line[0])
		key = strings.TrimSpace(key)

		val := line[1]
		val = strings.TrimSpace(val)

		dataset[key] = val
	}

	return dataset, nil
}
