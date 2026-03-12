package client

import "testing"

func TestWhitelistFilter(t *testing.T) {
	tests := []struct {
		name      string
		whitelist []string
		app       Application
		expected  bool
	}{
		{
			name:      "empty whitelist allows all applications",
			whitelist: []string{},
			app:       Application{Name: "any-app"},
			expected:  true,
		},
		{
			name:      "nil whitelist allows all applications",
			whitelist: nil,
			app:       Application{Name: "any-app"},
			expected:  true,
		},
		{
			name:      "allows Application in whitelist",
			whitelist: []string{"app1", "app2", "app3"},
			app:       Application{Name: "app2"},
			expected:  true,
		},
		{
			name:      "blocks Application not in whitelist",
			whitelist: []string{"app1", "app2", "app3"},
			app:       Application{Name: "app4"},
			expected:  false,
		},
		{
			name:      "single Application in whitelist - Allowed",
			whitelist: []string{"only-app"},
			app:       Application{Name: "only-app"},
			expected:  true,
		},
		{
			name:      "single Application in whitelist - blocked",
			whitelist: []string{"only-app"},
			app:       Application{Name: "other-app"},
			expected:  false,
		},
		{
			name:      "whitelist with duplicate entries",
			whitelist: []string{"app1", "app1", "app2"},
			app:       Application{Name: "app1"},
			expected:  true,
		},
		{
			name:      "case does not matter",
			whitelist: []string{"App1"},
			app:       Application{Name: "app1"},
			expected:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			filter := WhitelistFilter(tt.whitelist)
			result := filter.Allowed(tt.app)
			if result != tt.expected {
				t.Errorf("WhitelistFilter.Allowed() = %v, want %v for app %q with whitelist %v",
					result, tt.expected, tt.app.Name, tt.whitelist)
			}
		})
	}
}

func TestBlacklistFilter(t *testing.T) {
	tests := []struct {
		name      string
		blacklist []string
		app       Application
		expected  bool
	}{
		{
			name:      "empty blacklist allows all applications",
			blacklist: []string{},
			app:       Application{Name: "any-app"},
			expected:  true,
		},
		{
			name:      "nil blacklist allows all applications",
			blacklist: nil,
			app:       Application{Name: "any-app"},
			expected:  true,
		},
		{
			name:      "blocks Application in blacklist",
			blacklist: []string{"app1", "app2", "app3"},
			app:       Application{Name: "app2"},
			expected:  false,
		},
		{
			name:      "allows Application not in blacklist",
			blacklist: []string{"app1", "app2", "app3"},
			app:       Application{Name: "app4"},
			expected:  true,
		},
		{
			name:      "single Application in blacklist - blocked",
			blacklist: []string{"blocked-app"},
			app:       Application{Name: "blocked-app"},
			expected:  false,
		},
		{
			name:      "single Application in blacklist - Allowed",
			blacklist: []string{"blocked-app"},
			app:       Application{Name: "other-app"},
			expected:  true,
		},
		{
			name:      "blacklist with duplicate entries",
			blacklist: []string{"app1", "app1", "app2"},
			app:       Application{Name: "app1"},
			expected:  false,
		},
		{
			name:      "case does not matter",
			blacklist: []string{"App1"},
			app:       Application{Name: "app1"},
			expected:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			filter := BlacklistFilter(tt.blacklist)
			result := filter.Allowed(tt.app)
			if result != tt.expected {
				t.Errorf("BlacklistFilter.Allowed() = %v, want %v for app %q with blacklist %v",
					result, tt.expected, tt.app.Name, tt.blacklist)
			}
		})
	}
}

func TestFilterChain_Allowed(t *testing.T) {
	tests := []struct {
		name     string
		chain    FilterChain
		app      Application
		expected bool
	}{
		{
			name:     "empty chain allows Application",
			chain:    FilterChain{},
			app:      Application{Name: "test-app"},
			expected: true,
		},
		{
			name: "single Filter that allows",
			chain: FilterChain{
				FilterFunc(func(a Application) bool { return true }),
			},
			app:      Application{Name: "test-app"},
			expected: true,
		},
		{
			name: "single Filter that blocks",
			chain: FilterChain{
				FilterFunc(func(a Application) bool { return false }),
			},
			app:      Application{Name: "test-app"},
			expected: false,
		},
		{
			name: "multiple filters all allow",
			chain: FilterChain{
				FilterFunc(func(a Application) bool { return true }),
				FilterFunc(func(a Application) bool { return true }),
				FilterFunc(func(a Application) bool { return true }),
			},
			app:      Application{Name: "test-app"},
			expected: true,
		},
		{
			name: "first Filter blocks",
			chain: FilterChain{
				FilterFunc(func(a Application) bool { return false }),
				FilterFunc(func(a Application) bool { return true }),
				FilterFunc(func(a Application) bool { return true }),
			},
			app:      Application{Name: "test-app"},
			expected: false,
		},
		{
			name: "middle Filter blocks",
			chain: FilterChain{
				FilterFunc(func(a Application) bool { return true }),
				FilterFunc(func(a Application) bool { return false }),
				FilterFunc(func(a Application) bool { return true }),
			},
			app:      Application{Name: "test-app"},
			expected: false,
		},
		{
			name: "last Filter blocks",
			chain: FilterChain{
				FilterFunc(func(a Application) bool { return true }),
				FilterFunc(func(a Application) bool { return true }),
				FilterFunc(func(a Application) bool { return false }),
			},
			app:      Application{Name: "test-app"},
			expected: false,
		},
		{
			name: "whitelist and blacklist combination - Allowed",
			chain: FilterChain{
				WhitelistFilter([]string{"app1", "app2", "app3"}),
				BlacklistFilter([]string{"blocked-app"}),
			},
			app:      Application{Name: "app2"},
			expected: true,
		},
		{
			name: "whitelist and blacklist combination - blocked by whitelist",
			chain: FilterChain{
				WhitelistFilter([]string{"app1", "app2", "app3"}),
				BlacklistFilter([]string{"blocked-app"}),
			},
			app:      Application{Name: "app4"},
			expected: false,
		},
		{
			name: "whitelist and blacklist combination - blocked by blacklist",
			chain: FilterChain{
				WhitelistFilter([]string{"app1", "app2", "app3"}),
				BlacklistFilter([]string{"app2"}),
			},
			app:      Application{Name: "app2"},
			expected: false,
		},
		{
			name: "no whitelist and not blocked by blacklist",
			chain: FilterChain{
				WhitelistFilter(nil),
				BlacklistFilter([]string{"app2"}),
			},
			app:      Application{Name: "app3"},
			expected: true,
		},
		{
			name: "no whitelist and blocked by blacklist",
			chain: FilterChain{
				WhitelistFilter(nil),
				BlacklistFilter([]string{"app3"}),
			},
			app:      Application{Name: "app3"},
			expected: false,
		},
		{
			name: "Allowed by whitelist and no blacklist",
			chain: FilterChain{
				WhitelistFilter([]string{"app2"}),
				BlacklistFilter(nil),
			},
			app:      Application{Name: "app2"},
			expected: true,
		},
		{
			name: "denied by whitelist and no blacklist",
			chain: FilterChain{
				WhitelistFilter([]string{"app2"}),
				BlacklistFilter(nil),
			},
			app:      Application{Name: "app3"},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.chain.Allowed(tt.app)
			if result != tt.expected {
				t.Errorf("FilterChain.Allowed() = %v, want %v for app %q",
					result, tt.expected, tt.app.Name)
			}
		})
	}
}
