package cache

import (
	"context"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func FuzzExtractURL(f *testing.F) {
	f.Add("redis://localhost:6379")
	f.Add("redis://localhost:6379/0")
	f.Add("redis://localhost:6379/1")
	f.Add("redis://user:pass@localhost:6379/2")
	f.Add("redis://127.0.0.1:6379")
	f.Add("redis://[::1]:6379")
	f.Add("redis://localhost:6379/database")
	f.Add("")
	f.Add("invalid-url")
	f.Add("redis://")
	f.Add("redis://localhost")
	f.Add("redis://localhost:")
	f.Add("redis://localhost:abc")
	f.Add("redis://localhost:99999")
	f.Add("redis://localhost:0")
	f.Add("redis://unicode-host-🌟:6379")
	f.Add("redis://localhost:6379/unicode-db-🎯")

	f.Fuzz(func(t *testing.T, redisURL string) {
		assert.NotPanics(t, func() {
			host, database, port, err := extractURL(redisURL)

			if err == nil {
				assert.IsType(t, "", host)
				assert.IsType(t, "", database)
				assert.IsType(t, "", port)

				if (strings.Contains(redisURL, "localhost") || strings.Contains(redisURL, "127.0.0.1")) &&
					strings.Contains(redisURL, ":") && !strings.HasSuffix(redisURL, ":") && port != "" {
					assert.NotEmpty(t, host)
					assert.NotEmpty(t, port)
				}
			}
		})
	})
}

func FuzzBuildDialer(f *testing.F) {
	f.Add("redis://localhost:6379")
	f.Add("redis://127.0.0.1:6379/0")
	f.Add("redis://user:pass@localhost:6379/1")
	f.Add("")
	f.Add("invalid-url")
	f.Add("redis://")
	f.Add("redis://localhost")
	f.Add("redis://localhost:abc")
	f.Add("redis://localhost:99999")

	f.Fuzz(func(t *testing.T, redisURL string) {
		assert.NotPanics(t, func() {
			dialer := buildDialer(redisURL)
			assert.NotNil(t, dialer)

			if redisURL != "" {
				_, err := dialer()
				if err != nil {
					assert.Error(t, err)
				}
			}
		})
	})
}

func FuzzConnectParams(f *testing.F) {
	f.Add("redis://localhost:6379", 10, 5, int64(60), int64(240), true, false)
	f.Add("redis://127.0.0.1:6379/0", 0, 0, int64(0), int64(0), false, false)
	f.Add("", 10, 5, int64(60), int64(240), true, true)
	f.Add("redis://localhost:6379", -1, -1, int64(-1), int64(-1), false, false)
	f.Add("redis://localhost:6379", 1000, 500, int64(3600), int64(7200), true, true)

	f.Fuzz(func(t *testing.T, redisURL string, maxActive, maxIdle int, maxConnLifetime, idleTimeout int64, dependencyMode, newRelicEnabled bool) {
		if maxActive < 0 {
			maxActive = 0
		}
		if maxIdle < 0 {
			maxIdle = 0
		}
		if maxConnLifetime < 0 {
			maxConnLifetime = 0
		}
		if idleTimeout < 0 {
			idleTimeout = 0
		}

		if maxActive > 10000 {
			maxActive = 10000
		}
		if maxIdle > 1000 {
			maxIdle = 1000
		}
		if maxConnLifetime > 86400 {
			maxConnLifetime = 86400
		}
		if idleTimeout > 86400 {
			idleTimeout = 86400
		}

		ctx := context.Background()
		maxConnLifetimeDuration := time.Duration(maxConnLifetime) * time.Second
		idleTimeoutDuration := time.Duration(idleTimeout) * time.Second

		assert.NotPanics(t, func() {
			client, err := Connect(ctx, redisURL, maxActive, maxIdle, maxConnLifetimeDuration, idleTimeoutDuration, dependencyMode, newRelicEnabled)
			if err != nil {
				assert.Nil(t, client)
			} else {
				assert.NotNil(t, client)
				if client != nil {
					client.Close()
				}
			}
		})
	})
}

func FuzzURLParsing(f *testing.F) {
	f.Add("redis://localhost:6379")
	f.Add("redis://localhost:6379/")
	f.Add("redis://localhost:6379/0")
	f.Add("redis://localhost:6379/database")
	f.Add("redis://user@localhost:6379")
	f.Add("redis://user:@localhost:6379")
	f.Add("redis://:pass@localhost:6379")
	f.Add("redis://user:pass@localhost:6379")
	f.Add("redis://user:pass@localhost:6379/0")
	f.Add("redis://user:pass@localhost:6379/database")
	f.Add("rediss://localhost:6380")
	f.Add("redis://[::1]:6379")
	f.Add("redis://[2001:db8::1]:6379")
	f.Add("")
	f.Add("localhost:6379")
	f.Add("://localhost:6379")
	f.Add("redis://")
	f.Add("redis:///")
	f.Add("redis://localhost")
	f.Add("redis://localhost:")
	f.Add("redis://localhost:abc")
	f.Add("redis://localhost:-1")
	f.Add("redis://localhost:99999")
	f.Add("redis://localhost:6379?param=value")
	f.Add("redis://localhost:6379#fragment")
	f.Add("redis://unicode-🌟:6379")
	f.Add("redis://localhost:6379/unicode-🎯")

	f.Fuzz(func(t *testing.T, testURL string) {
		assert.NotPanics(t, func() {
			parsedURL, err := url.Parse(testURL)
			if err == nil {
				assert.NotNil(t, parsedURL)
				assert.IsType(t, "", parsedURL.Scheme)
				assert.IsType(t, "", parsedURL.Host)
				assert.IsType(t, "", parsedURL.Path)

				if parsedURL.Scheme == "redis" || parsedURL.Scheme == "rediss" {
					if parsedURL.Host != "" && strings.Contains(parsedURL.Host, ":") && !strings.HasSuffix(parsedURL.Host, ":") {
						host, database, port, extractErr := extractURL(testURL)
						if extractErr == nil {
							// Only assert host is non-empty if the URL actually has a meaningful host part
							// Skip assertion for edge cases like ":port" or "[]:port" which have empty hosts
							if !strings.HasPrefix(parsedURL.Host, ":") && parsedURL.Host != "[]:"+port {
								assert.NotEmpty(t, host)
							}
							assert.NotEmpty(t, port)
							assert.IsType(t, "", database)
						}
					}
				}
			}
		})
	})
}

func FuzzClientCreation(f *testing.F) {
	f.Add(true)

	f.Fuzz(func(t *testing.T, _ bool) {
		assert.NotPanics(t, func() {
			client := &Client{
				DependencyScriptSha: "",
				Pool:                nil,
				ScriptsLoaded:       nil,
			}

			assert.NotNil(t, client)
			assert.Empty(t, client.DependencyScriptSha)
			assert.Nil(t, client.Pool)
			assert.Nil(t, client.ScriptsLoaded)

			client.Close()

			_, err := client.GetConnectionWithContext(context.Background())
			assert.Error(t, err)
			assert.Equal(t, ErrRedisPoolNil, err)
		})
	})
}

func FuzzConnectionManagement(f *testing.F) {
	f.Add(true)

	f.Fuzz(func(t *testing.T, _ bool) {
		assert.NotPanics(t, func() {
			client := &Client{}

			conn := client.CloseConnection(nil)
			assert.Nil(t, conn)

			conn = CloseConnection(nil)
			assert.Nil(t, conn)

			client.Close()
			assert.Nil(t, client.Pool)
		})
	})
}

func FuzzScriptShaValidation(f *testing.F) {
	f.Add("a648f768f57e73e2497ccaa113d5ad9e731c5cd8")
	f.Add("")
	f.Add("invalid-sha")
	f.Add("1234567890abcdef")
	f.Add(strings.Repeat("a", 40))
	f.Add(strings.Repeat("z", 40))
	f.Add("unicode-🔑-sha")

	f.Fuzz(func(t *testing.T, sha string) {
		assert.NotPanics(t, func() {
			client := &Client{
				DependencyScriptSha: sha,
				Pool:                nil,
				ScriptsLoaded:       []string{sha},
			}

			assert.Equal(t, sha, client.DependencyScriptSha)
			if sha != "" {
				assert.Contains(t, client.ScriptsLoaded, sha)
			}

			assert.Equal(t, killByDependencySha, "a648f768f57e73e2497ccaa113d5ad9e731c5cd8")
		})
	})
}

func FuzzURLEdgeCases(f *testing.F) {
	f.Add("redis://localhost:6379/0/extra/path")
	f.Add("redis://localhost:6379?query=param")
	f.Add("redis://localhost:6379#fragment")
	f.Add("redis://localhost:6379?query=param&other=value#fragment")
	f.Add("redis://user%40domain:pass%40word@localhost:6379")
	f.Add("redis://localhost:6379/" + strings.Repeat("db", 100))
	f.Add("REDIS://LOCALHOST:6379")
	f.Add("redis://LOCALHOST:6379")

	f.Fuzz(func(t *testing.T, redisURL string) {
		assert.NotPanics(t, func() {
			host, database, port, err := extractURL(redisURL)

			if err == nil {
				assert.IsType(t, "", host)
				assert.IsType(t, "", database)
				assert.IsType(t, "", port)

				if strings.ToLower(redisURL) != redisURL && strings.Contains(strings.ToLower(redisURL), "localhost") &&
					strings.Contains(redisURL, ":") && !strings.HasSuffix(redisURL, ":") {
					assert.NotEmpty(t, host)
					assert.NotEmpty(t, port)
				}
			}

			dialer := buildDialer(redisURL)
			assert.NotNil(t, dialer)
		})
	})
}

func FuzzPoolConfiguration(f *testing.F) {
	f.Add(true, true)
	f.Add(true, false)
	f.Add(false, true)
	f.Add(false, false)

	f.Fuzz(func(t *testing.T, dependencyMode, newRelicEnabled bool) {
		assert.NotPanics(t, func() {
			ctx := context.Background()
			redisURL := "redis://localhost:6379"
			maxActive := 10
			maxIdle := 5
			maxConnLifetime := 60 * time.Second
			idleTimeout := 240 * time.Second

			client, err := Connect(ctx, redisURL, maxActive, maxIdle, maxConnLifetime, idleTimeout, dependencyMode, newRelicEnabled)

			if err != nil {
				assert.Nil(t, client)
			} else {
				assert.NotNil(t, client)
				assert.NotNil(t, client.Pool)

				if dependencyMode {
					assert.NotEmpty(t, client.DependencyScriptSha)
				}

				client.Close()
			}
		})
	})
}
