package main

import "testing"

func TestWhitelistFilter(t *testing.T) {
	tests := []struct {
		name      string
		whitelist []string
		app       application
		expected  bool
	}{
		{
			name:      "empty whitelist allows all applications",
			whitelist: []string{},
			app:       application{Name: "any-app"},
			expected:  true,
		},
		{
			name:      "nil whitelist allows all applications",
			whitelist: nil,
			app:       application{Name: "any-app"},
			expected:  true,
		},
		{
			name:      "allows application in whitelist",
			whitelist: []string{"app1", "app2", "app3"},
			app:       application{Name: "app2"},
			expected:  true,
		},
		{
			name:      "blocks application not in whitelist",
			whitelist: []string{"app1", "app2", "app3"},
			app:       application{Name: "app4"},
			expected:  false,
		},
		{
			name:      "single application in whitelist - allowed",
			whitelist: []string{"only-app"},
			app:       application{Name: "only-app"},
			expected:  true,
		},
		{
			name:      "single application in whitelist - blocked",
			whitelist: []string{"only-app"},
			app:       application{Name: "other-app"},
			expected:  false,
		},
		{
			name:      "whitelist with duplicate entries",
			whitelist: []string{"app1", "app1", "app2"},
			app:       application{Name: "app1"},
			expected:  true,
		},
		{
			name:      "case does not matter",
			whitelist: []string{"App1"},
			app:       application{Name: "app1"},
			expected:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			filter := whitelistFilter(tt.whitelist)
			result := filter.allowed(tt.app)
			if result != tt.expected {
				t.Errorf("whitelistFilter.allowed() = %v, want %v for app %q with whitelist %v",
					result, tt.expected, tt.app.Name, tt.whitelist)
			}
		})
	}
}

func TestBlacklistFilter(t *testing.T) {
	tests := []struct {
		name      string
		blacklist []string
		app       application
		expected  bool
	}{
		{
			name:      "empty blacklist allows all applications",
			blacklist: []string{},
			app:       application{Name: "any-app"},
			expected:  true,
		},
		{
			name:      "nil blacklist allows all applications",
			blacklist: nil,
			app:       application{Name: "any-app"},
			expected:  true,
		},
		{
			name:      "blocks application in blacklist",
			blacklist: []string{"app1", "app2", "app3"},
			app:       application{Name: "app2"},
			expected:  false,
		},
		{
			name:      "allows application not in blacklist",
			blacklist: []string{"app1", "app2", "app3"},
			app:       application{Name: "app4"},
			expected:  true,
		},
		{
			name:      "single application in blacklist - blocked",
			blacklist: []string{"blocked-app"},
			app:       application{Name: "blocked-app"},
			expected:  false,
		},
		{
			name:      "single application in blacklist - allowed",
			blacklist: []string{"blocked-app"},
			app:       application{Name: "other-app"},
			expected:  true,
		},
		{
			name:      "blacklist with duplicate entries",
			blacklist: []string{"app1", "app1", "app2"},
			app:       application{Name: "app1"},
			expected:  false,
		},
		{
			name:      "case does not matter",
			blacklist: []string{"App1"},
			app:       application{Name: "app1"},
			expected:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			filter := blacklistFilter(tt.blacklist)
			result := filter.allowed(tt.app)
			if result != tt.expected {
				t.Errorf("blacklistFilter.allowed() = %v, want %v for app %q with blacklist %v",
					result, tt.expected, tt.app.Name, tt.blacklist)
			}
		})
	}
}

func TestFilterChain_Allowed(t *testing.T) {
	tests := []struct {
		name     string
		chain    filterChain
		app      application
		expected bool
	}{
		{
			name:     "empty chain allows application",
			chain:    filterChain{},
			app:      application{Name: "test-app"},
			expected: true,
		},
		{
			name: "single filter that allows",
			chain: filterChain{
				filterFunc(func(a application) bool { return true }),
			},
			app:      application{Name: "test-app"},
			expected: true,
		},
		{
			name: "single filter that blocks",
			chain: filterChain{
				filterFunc(func(a application) bool { return false }),
			},
			app:      application{Name: "test-app"},
			expected: false,
		},
		{
			name: "multiple filters all allow",
			chain: filterChain{
				filterFunc(func(a application) bool { return true }),
				filterFunc(func(a application) bool { return true }),
				filterFunc(func(a application) bool { return true }),
			},
			app:      application{Name: "test-app"},
			expected: true,
		},
		{
			name: "first filter blocks",
			chain: filterChain{
				filterFunc(func(a application) bool { return false }),
				filterFunc(func(a application) bool { return true }),
				filterFunc(func(a application) bool { return true }),
			},
			app:      application{Name: "test-app"},
			expected: false,
		},
		{
			name: "middle filter blocks",
			chain: filterChain{
				filterFunc(func(a application) bool { return true }),
				filterFunc(func(a application) bool { return false }),
				filterFunc(func(a application) bool { return true }),
			},
			app:      application{Name: "test-app"},
			expected: false,
		},
		{
			name: "last filter blocks",
			chain: filterChain{
				filterFunc(func(a application) bool { return true }),
				filterFunc(func(a application) bool { return true }),
				filterFunc(func(a application) bool { return false }),
			},
			app:      application{Name: "test-app"},
			expected: false,
		},
		{
			name: "whitelist and blacklist combination - allowed",
			chain: filterChain{
				whitelistFilter([]string{"app1", "app2", "app3"}),
				blacklistFilter([]string{"blocked-app"}),
			},
			app:      application{Name: "app2"},
			expected: true,
		},
		{
			name: "whitelist and blacklist combination - blocked by whitelist",
			chain: filterChain{
				whitelistFilter([]string{"app1", "app2", "app3"}),
				blacklistFilter([]string{"blocked-app"}),
			},
			app:      application{Name: "app4"},
			expected: false,
		},
		{
			name: "whitelist and blacklist combination - blocked by blacklist",
			chain: filterChain{
				whitelistFilter([]string{"app1", "app2", "app3"}),
				blacklistFilter([]string{"app2"}),
			},
			app:      application{Name: "app2"},
			expected: false,
		},
		{
			name: "no whitelist and not blocked by blacklist",
			chain: filterChain{
				whitelistFilter(nil),
				blacklistFilter([]string{"app2"}),
			},
			app:      application{Name: "app3"},
			expected: true,
		},
		{
			name: "no whitelist and blocked by blacklist",
			chain: filterChain{
				whitelistFilter(nil),
				blacklistFilter([]string{"app3"}),
			},
			app:      application{Name: "app3"},
			expected: false,
		},
		{
			name: "allowed by whitelist and no blacklist",
			chain: filterChain{
				whitelistFilter([]string{"app2"}),
				blacklistFilter(nil),
			},
			app:      application{Name: "app2"},
			expected: true,
		},
		{
			name: "denied by whitelist and no blacklist",
			chain: filterChain{
				whitelistFilter([]string{"app2"}),
				blacklistFilter(nil),
			},
			app:      application{Name: "app3"},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.chain.allowed(tt.app)
			if result != tt.expected {
				t.Errorf("filterChain.allowed() = %v, want %v for app %q",
					result, tt.expected, tt.app.Name)
			}
		})
	}
}
