package config

import (
	"slices"
	"testing"
	"time"
)

func TestLoadConfig(t *testing.T) {
	tests := []struct {
		name     string
		envVars  map[string]string
		expected Config
	}{
		{
			name: "loads all basic configuration values",
			envVars: map[string]string{
				"TRUENAS_URL":      "https://truenas.local",
				"TRUENAS_USERNAME": "admin",
				"TRUENAS_PASSWORD": "password123",
				"TRUENAS_API_KEY":  "api-key-xyz",
			},
			expected: Config{
				URL:        "https://truenas.local",
				Username:   "admin",
				Password:   "password123",
				APIKey:     "api-key-xyz",
				CheckDelay: 6 * time.Hour,
			},
		},
		{
			name: "uses default check delay when not provided",
			envVars: map[string]string{
				"TRUENAS_URL": "https://truenas.local",
			},
			expected: Config{
				URL:        "https://truenas.local",
				CheckDelay: 6 * time.Hour,
			},
		},
		{
			name: "parses custom check delay",
			envVars: map[string]string{
				"TRUENAS_URL": "https://truenas.local",
				"CHECK_DELAY": "1h30m",
			},
			expected: Config{
				URL:        "https://truenas.local",
				CheckDelay: 90 * time.Minute,
			},
		},
		{
			name: "falls back to default when check delay is invalid",
			envVars: map[string]string{
				"TRUENAS_URL": "https://truenas.local",
				"CHECK_DELAY": "invalid-duration",
			},
			expected: Config{
				URL:        "https://truenas.local",
				CheckDelay: 6 * time.Hour,
			},
		},
		{
			name: "parses whitelist with single application",
			envVars: map[string]string{
				"TRUENAS_URL":   "https://truenas.local",
				"APP_WHITELIST": "app1",
			},
			expected: Config{
				URL:         "https://truenas.local",
				CheckDelay:  6 * time.Hour,
				Whitelisted: []string{"app1"},
			},
		},
		{
			name: "parses whitelist with multiple applications",
			envVars: map[string]string{
				"TRUENAS_URL":   "https://truenas.local",
				"APP_WHITELIST": "app1,app2,app3",
			},
			expected: Config{
				URL:         "https://truenas.local",
				CheckDelay:  6 * time.Hour,
				Whitelisted: []string{"app1", "app2", "app3"},
			},
		},
		{
			name: "parses blacklist with single application",
			envVars: map[string]string{
				"TRUENAS_URL":   "https://truenas.local",
				"APP_BLACKLIST": "blocked-app",
			},
			expected: Config{
				URL:         "https://truenas.local",
				CheckDelay:  6 * time.Hour,
				Blacklisted: []string{"blocked-app"},
			},
		},
		{
			name: "parses blacklist with multiple applications",
			envVars: map[string]string{
				"TRUENAS_URL":   "https://truenas.local",
				"APP_BLACKLIST": "app1,app2,app3",
			},
			expected: Config{
				URL:         "https://truenas.local",
				CheckDelay:  6 * time.Hour,
				Blacklisted: []string{"app1", "app2", "app3"},
			},
		},
		{
			name: "parses both whitelist and blacklist",
			envVars: map[string]string{
				"TRUENAS_URL":   "https://truenas.local",
				"APP_WHITELIST": "app1,app2,app3",
				"APP_BLACKLIST": "blocked1,blocked2",
			},
			expected: Config{
				URL:         "https://truenas.local",
				CheckDelay:  6 * time.Hour,
				Whitelisted: []string{"app1", "app2", "app3"},
				Blacklisted: []string{"blocked1", "blocked2"},
			},
		},
		{
			name: "trims whitespace from whitelist",
			envVars: map[string]string{
				"TRUENAS_URL":   "https://truenas.local",
				"APP_WHITELIST": "  app1,app2,app3  ",
			},
			expected: Config{
				URL:         "https://truenas.local",
				CheckDelay:  6 * time.Hour,
				Whitelisted: []string{"app1", "app2", "app3"},
			},
		},
		{
			name: "trims whitespace from blacklist",
			envVars: map[string]string{
				"TRUENAS_URL":   "https://truenas.local",
				"APP_BLACKLIST": "  blocked1,blocked2  ",
			},
			expected: Config{
				URL:         "https://truenas.local",
				CheckDelay:  6 * time.Hour,
				Blacklisted: []string{"blocked1", "blocked2"},
			},
		},
		{
			name: "ignores empty whitelist",
			envVars: map[string]string{
				"TRUENAS_URL":   "https://truenas.local",
				"APP_WHITELIST": "",
			},
			expected: Config{
				URL:        "https://truenas.local",
				CheckDelay: 6 * time.Hour,
			},
		},
		{
			name: "ignores whitespace-only whitelist",
			envVars: map[string]string{
				"TRUENAS_URL":   "https://truenas.local",
				"APP_WHITELIST": "   ",
			},
			expected: Config{
				URL:        "https://truenas.local",
				CheckDelay: 6 * time.Hour,
			},
		},
		{
			name: "ignores empty blacklist",
			envVars: map[string]string{
				"TRUENAS_URL":   "https://truenas.local",
				"APP_BLACKLIST": "",
			},
			expected: Config{
				URL:        "https://truenas.local",
				CheckDelay: 6 * time.Hour,
			},
		},
		{
			name: "ignores whitespace-only blacklist",
			envVars: map[string]string{
				"TRUENAS_URL":   "https://truenas.local",
				"APP_BLACKLIST": "   ",
			},
			expected: Config{
				URL:        "https://truenas.local",
				CheckDelay: 6 * time.Hour,
			},
		},
		{
			name: "loads complete configuration with all fields",
			envVars: map[string]string{
				"TRUENAS_URL":      "https://truenas.example.com",
				"TRUENAS_USERNAME": "superadmin",
				"TRUENAS_PASSWORD": "secret",
				"TRUENAS_API_KEY":  "key123",
				"CHECK_DELAY":      "12h",
				"APP_WHITELIST":    "webapp,backend,frontend",
				"APP_BLACKLIST":    "test-app,dev-app",
			},
			expected: Config{
				URL:         "https://truenas.example.com",
				Username:    "superadmin",
				Password:    "secret",
				APIKey:      "key123",
				CheckDelay:  12 * time.Hour,
				Whitelisted: []string{"webapp", "backend", "frontend"},
				Blacklisted: []string{"test-app", "dev-app"},
			},
		},
		{
			name: "handles various time duration formats",
			envVars: map[string]string{
				"TRUENAS_URL": "https://truenas.local",
				"CHECK_DELAY": "2h30m45s",
			},
			expected: Config{
				URL:        "https://truenas.local",
				CheckDelay: 2*time.Hour + 30*time.Minute + 45*time.Second,
			},
		},
		{
			name: "handles minutes as check delay",
			envVars: map[string]string{
				"TRUENAS_URL": "https://truenas.local",
				"CHECK_DELAY": "45m",
			},
			expected: Config{
				URL:        "https://truenas.local",
				CheckDelay: 45 * time.Minute,
			},
		},
		{
			name: "handles seconds as check delay",
			envVars: map[string]string{
				"TRUENAS_URL": "https://truenas.local",
				"CHECK_DELAY": "300s",
			},
			expected: Config{
				URL:        "https://truenas.local",
				CheckDelay: 300 * time.Second,
			},
		},
		{
			name: "handles mixed case and whitespace in whitelist",
			envVars: map[string]string{
				"TRUENAS_URL":   "https://truenas.local",
				"APP_WHITELIST": "  MyApp  , ANOTHER_APP ,  third-app  ",
			},
			expected: Config{
				URL:         "https://truenas.local",
				CheckDelay:  6 * time.Hour,
				Whitelisted: []string{"myapp", "another_app", "third-app"},
			},
		},
		{
			name: "handles mixed case and whitespace in blacklist",
			envVars: map[string]string{
				"TRUENAS_URL":   "https://truenas.local",
				"APP_BLACKLIST": "  BlockedApp  , ANOTHER_BLOCKED ,  third-blocked  ",
			},
			expected: Config{
				URL:         "https://truenas.local",
				CheckDelay:  6 * time.Hour,
				Blacklisted: []string{"blockedapp", "another_blocked", "third-blocked"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set up test environment variables
			for key, value := range tt.envVars {
				t.Setenv(key, value)
			}

			// Load configuration
			result := LoadConfig()

			// Verify all fields
			if result.URL != tt.expected.URL {
				t.Errorf("LoadConfig().URL = %q, want %q", result.URL, tt.expected.URL)
			}
			if result.Username != tt.expected.Username {
				t.Errorf("LoadConfig().Username = %q, want %q", result.Username, tt.expected.Username)
			}
			if result.Password != tt.expected.Password {
				t.Errorf("LoadConfig().Password = %q, want %q", result.Password, tt.expected.Password)
			}
			if result.APIKey != tt.expected.APIKey {
				t.Errorf("LoadConfig().APIKey = %q, want %q", result.APIKey, tt.expected.APIKey)
			}
			if result.CheckDelay != tt.expected.CheckDelay {
				t.Errorf("LoadConfig().CheckDelay = %v, want %v", result.CheckDelay, tt.expected.CheckDelay)
			}

			// Compare slices
			if !slices.Equal(result.Whitelisted, tt.expected.Whitelisted) {
				t.Errorf("LoadConfig().Whitelisted = %v, want %v", result.Whitelisted, tt.expected.Whitelisted)
			}
			if !slices.Equal(result.Blacklisted, tt.expected.Blacklisted) {
				t.Errorf("LoadConfig().Blacklisted = %v, want %v", result.Blacklisted, tt.expected.Blacklisted)
			}
		})
	}
}
