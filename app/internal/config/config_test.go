package config

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoad(t *testing.T) {
	type expected struct {
		wantErr     bool
		errContains string

		transport string
		name      string
		version   string
		path      string

		port         string
		readTimeout  time.Duration
		writeTimeout time.Duration

		logMode string

		maxIdleConnsPerHost int
		proxyHost           string
		proxyLogin          string

		maxQueries           int
		defaultResults       int
		maxResults           int
		defaultTimeoutMs     int64
		maxTimeoutMs         int64
		minTimeoutMs         int64
		defaultImages        int
		maxImages            int
		defaultDocumentChars int
		maxDocumentChars     int
	}

	cases := []struct {
		name     string
		env      map[string]string
		envFile  string // content of the env file; empty means no env file passed
		expected expected
	}{
		{
			name: "defaults when no env file and no env vars",
			expected: expected{
				transport:            "stdio",
				name:                 "mcp-retrieval",
				version:              "0.1.0",
				path:                 "/mcp",
				port:                 "8080",
				readTimeout:          60 * time.Second,
				writeTimeout:         60 * time.Second,
				logMode:              "local",
				maxIdleConnsPerHost:  100,
				maxQueries:           10,
				defaultResults:       5,
				maxResults:           20,
				defaultTimeoutMs:     5000,
				maxTimeoutMs:         10000,
				minTimeoutMs:         1000,
				defaultImages:        5,
				maxImages:            10,
				defaultDocumentChars: 20000,
				maxDocumentChars:     20000,
			},
		},
		{
			name: "env overrides defaults",
			env: map[string]string{
				"MCP_TRANSPORT": "http",
				"MCP_NAME":      "env-name",
				"SERVER_PORT":   "7070",
				"LOG_MODE":      "prod",
				"MAX_QUERIES":   "7",
			},
			expected: expected{
				transport:            "http",
				name:                 "env-name",
				version:              "0.1.0",
				path:                 "/mcp",
				port:                 "7070",
				readTimeout:          60 * time.Second,
				writeTimeout:         60 * time.Second,
				logMode:              "prod",
				maxIdleConnsPerHost:  100,
				maxQueries:           7,
				defaultResults:       5,
				maxResults:           20,
				defaultTimeoutMs:     5000,
				maxTimeoutMs:         10000,
				minTimeoutMs:         1000,
				defaultImages:        5,
				maxImages:            10,
				defaultDocumentChars: 20000,
				maxDocumentChars:     20000,
			},
		},
		{
			name:    "env file is loaded",
			envFile: "MCP_NAME=from-env-file\nPROXY_HOST=proxy.local\nPROXY_LOGIN=user\nPROXY_PASSWORD=pass\nPROXY_PORT=3128\nPROXY_SCHEME=http\n",
			expected: expected{
				transport:            "stdio",
				name:                 "from-env-file",
				version:              "0.1.0",
				path:                 "/mcp",
				port:                 "8080",
				readTimeout:          60 * time.Second,
				writeTimeout:         60 * time.Second,
				logMode:              "local",
				maxIdleConnsPerHost:  100,
				proxyHost:            "proxy.local",
				proxyLogin:           "user",
				maxQueries:           10,
				defaultResults:       5,
				maxResults:           20,
				defaultTimeoutMs:     5000,
				maxTimeoutMs:         10000,
				minTimeoutMs:         1000,
				defaultImages:        5,
				maxImages:            10,
				defaultDocumentChars: 20000,
				maxDocumentChars:     20000,
			},
		},
		{
			name:     "env with non-numeric int",
			env:      map[string]string{"MAX_QUERIES": "abc"},
			expected: expected{wantErr: true, errContains: "parse env"},
		},
		{
			name:     "env with unparsable duration",
			env:      map[string]string{"SERVER_READ_TIMEOUT": "5 seconds"},
			expected: expected{wantErr: true, errContains: "parse env"},
		},
		{
			// godotenv never overwrites an already-set variable, so the real
			// environment outranks the file. This is the only precedence rule left.
			name:    "real environment wins over env file",
			env:     map[string]string{"MCP_NAME": "from-env"},
			envFile: "MCP_NAME=from-file\nLOG_MODE=prod\n",
			expected: expected{
				transport:            "stdio",
				name:                 "from-env",
				version:              "0.1.0",
				path:                 "/mcp",
				port:                 "8080",
				readTimeout:          60 * time.Second,
				writeTimeout:         60 * time.Second,
				logMode:              "prod",
				maxIdleConnsPerHost:  100,
				maxQueries:           10,
				defaultResults:       5,
				maxResults:           20,
				defaultTimeoutMs:     5000,
				maxTimeoutMs:         10000,
				minTimeoutMs:         1000,
				defaultImages:        5,
				maxImages:            10,
				defaultDocumentChars: 20000,
				maxDocumentChars:     20000,
			},
		},
		{
			name:     "validation fails on unknown transport from env",
			env:      map[string]string{"MCP_TRANSPORT": "grpc"},
			expected: expected{wantErr: true, errContains: "validate config"},
		},
		{
			name:     "validation fails when proxy host set without credentials",
			env:      map[string]string{"PROXY_HOST": "proxy.local"},
			expected: expected{wantErr: true, errContains: "validate config"},
		},
		{
			name:     "validation fails on non-positive max results",
			env:      map[string]string{"MAX_RESULTS": "0"},
			expected: expected{wantErr: true, errContains: "validate config"},
		},
		{
			name:     "validation fails on non-positive max idle conns",
			env:      map[string]string{"MAX_IDLE_CONNS_PER_HOST": "0"},
			expected: expected{wantErr: true, errContains: "validate config"},
		},
		{
			name:     "validation fails when default results exceed max results",
			env:      map[string]string{"DEFAULT_RESULTS": "50", "MAX_RESULTS": "10"},
			expected: expected{wantErr: true, errContains: "validate config"},
		},
		{
			name:     "validation fails when min timeout exceeds max timeout",
			env:      map[string]string{"MIN_TIMEOUT_MS": "20000", "MAX_TIMEOUT_MS": "10000"},
			expected: expected{wantErr: true, errContains: "validate config"},
		},
		{
			name: "validation fails when default timeout is below min timeout",
			env: map[string]string{
				"MIN_TIMEOUT_MS":     "3000",
				"DEFAULT_TIMEOUT_MS": "1000",
				"MAX_TIMEOUT_MS":     "10000",
			},
			expected: expected{wantErr: true, errContains: "validate config"},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {

			// isolate from the ambient environment (and from vars other cases loaded from .env files)
			for _, k := range []string{
				"MCP_TRANSPORT", "MCP_NAME", "MCP_VERSION", "MCP_PATH",
				"SERVER_PORT", "SERVER_READ_TIMEOUT", "SERVER_WRITE_TIMEOUT",
				"LOG_MODE", "MAX_IDLE_CONNS_PER_HOST",
				"PROXY_LOGIN", "PROXY_PASSWORD", "PROXY_HOST", "PROXY_PORT", "PROXY_SCHEME",
				"MAX_QUERIES", "DEFAULT_RESULTS", "MAX_RESULTS",
				"DEFAULT_TIMEOUT_MS", "MAX_TIMEOUT_MS", "MIN_TIMEOUT_MS",
				"DEFAULT_IMAGES", "MAX_IMAGES",
				"DEFAULT_DOCUMENT_CHARS", "MAX_DOCUMENT_CHARS",
			} {
				t.Setenv(k, "") // registers cleanup
				require.NoError(t, os.Unsetenv(k))
			}

			for k, v := range tc.env {
				t.Setenv(k, v)
			}

			envPath := ""
			if tc.envFile != "" {
				envPath = filepath.Join(t.TempDir(), ".env")
				require.NoError(t, os.WriteFile(envPath, []byte(tc.envFile), 0o600))
			}

			// act
			cfg, err := Load(envPath)

			// assert
			if tc.expected.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tc.expected.errContains)
				assert.Nil(t, cfg)

				return
			}

			require.NoError(t, err)
			require.NotNil(t, cfg)

			assert.Equal(t, tc.expected.transport, cfg.MCP.Transport)
			assert.Equal(t, tc.expected.name, cfg.MCP.Name)
			assert.Equal(t, tc.expected.version, cfg.MCP.Version)
			assert.Equal(t, tc.expected.path, cfg.MCP.Path)

			assert.Equal(t, tc.expected.port, cfg.Server.Port)
			assert.Equal(t, tc.expected.readTimeout, cfg.Server.ReadTimeout)
			assert.Equal(t, tc.expected.writeTimeout, cfg.Server.WriteTimeout)

			assert.Equal(t, tc.expected.logMode, cfg.Logger.Mode)

			assert.Equal(t, tc.expected.maxIdleConnsPerHost, cfg.RetrieverAdapter.MaxIdleConnsPerHost)
			assert.Equal(t, tc.expected.proxyHost, cfg.RetrieverAdapter.ProxyHost)
			assert.Equal(t, tc.expected.proxyLogin, cfg.RetrieverAdapter.ProxyLogin)

			assert.Equal(t, tc.expected.maxQueries, cfg.SearchUseCase.MaxQueries)
			assert.Equal(t, tc.expected.defaultResults, cfg.SearchUseCase.DefaultResults)
			assert.Equal(t, tc.expected.maxResults, cfg.SearchUseCase.MaxResults)
			assert.Equal(t, tc.expected.defaultTimeoutMs, cfg.SearchUseCase.DefaultTimeoutMs)
			assert.Equal(t, tc.expected.maxTimeoutMs, cfg.SearchUseCase.MaxTimeoutMs)
			assert.Equal(t, tc.expected.minTimeoutMs, cfg.SearchUseCase.MinTimeoutMs)
			assert.Equal(t, tc.expected.defaultImages, cfg.SearchUseCase.DefaultImages)
			assert.Equal(t, tc.expected.maxImages, cfg.SearchUseCase.MaxImages)
			assert.Equal(t, tc.expected.defaultDocumentChars, cfg.SearchUseCase.DefaultDocumentChars)
			assert.Equal(t, tc.expected.maxDocumentChars, cfg.SearchUseCase.MaxDocumentChars)
		})
	}
}

func TestLoadEnvFile(t *testing.T) {
	cases := []struct {
		name      string
		content   string
		writeIt   bool
		isDir     bool
		emptyPath bool
		expected  struct {
			wantErr bool
			key     string
			value   string
		}
	}{
		{
			name:    "existing file is applied",
			content: "SOME_TEST_KEY=some-value\n",
			writeIt: true,
			expected: struct {
				wantErr bool
				key     string
				value   string
			}{key: "SOME_TEST_KEY", value: "some-value"},
		},
		{
			name:    "missing file is not an error",
			writeIt: false,
			expected: struct {
				wantErr bool
				key     string
				value   string
			}{},
		},
		{
			name:      "empty path falls back to .env in the working directory",
			content:   "SOME_CWD_TEST_KEY=cwd-value\n",
			writeIt:   true,
			emptyPath: true,
			expected: struct {
				wantErr bool
				key     string
				value   string
			}{key: "SOME_CWD_TEST_KEY", value: "cwd-value"},
		},
		{
			name:      "empty path is not an error without .env in the working directory",
			emptyPath: true,
			expected: struct {
				wantErr bool
				key     string
				value   string
			}{},
		},
		{
			name:  "unreadable path returns error",
			isDir: true,
			expected: struct {
				wantErr bool
				key     string
				value   string
			}{wantErr: true},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			// arrange
			dir := t.TempDir()
			path := filepath.Join(dir, ".env")
			if tc.emptyPath {
				// the fallback resolves ".env" against the working directory
				t.Chdir(dir)
			}
			if tc.isDir {
				require.NoError(t, os.Mkdir(path, 0o750))
			}
			if tc.writeIt {
				require.NoError(t, os.WriteFile(path, []byte(tc.content), 0o600))
			}
			if tc.expected.key != "" {
				t.Setenv(tc.expected.key, "")
				require.NoError(t, os.Unsetenv(tc.expected.key))
			}

			// act
			arg := path
			if tc.emptyPath {
				arg = ""
			}
			err := loadEnvFile(arg)

			// assert
			if tc.expected.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), "load env file")

				return
			}

			require.NoError(t, err)
			if tc.expected.key != "" {
				assert.Equal(t, tc.expected.value, os.Getenv(tc.expected.key))
			}
		})
	}
}
