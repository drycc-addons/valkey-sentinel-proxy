package main

import (
	"net"
	"os"
	"testing"

	"github.com/valkey-io/valkey-go"
)

func TestMain(m *testing.M) {
	// Save original flag values
	originalListen := *listen
	originalMaster := *master
	originalMaxProcs := *maxProcs
	originalSentinelAddr := *sentinelAddr
	originalSentinelUser := *sentinelUser
	originalSentinelPass := *sentinelPass

	// Run tests
	code := m.Run()

	// Restore original flag values
	*listen = originalListen
	*master = originalMaster
	*maxProcs = originalMaxProcs
	*sentinelAddr = originalSentinelAddr
	*sentinelUser = originalSentinelUser
	*sentinelPass = originalSentinelPass

	os.Exit(code)
}

func TestFlagDefaults(t *testing.T) {
	// Test case: Verify default flag values
	tests := []struct {
		name     string
		flagPtr  interface{}
		expected interface{}
	}{
		{"listen default", *listen, ":9999"},
		{"master default", *master, ""},
		{"maxProcs default", *maxProcs, 1},
		{"sentinelAddr default", *sentinelAddr, ":26379"},
		{"sentinelUser default", *sentinelUser, ""},
		{"sentinelPass default", *sentinelPass, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.flagPtr != tt.expected {
				t.Errorf("Expected %v, got %v", tt.expected, tt.flagPtr)
			}
		})
	}
}

func TestResolveTCPAddr(t *testing.T) {
	// Test case: Valid address resolution
	testAddr := ":9999"
	addr, err := net.ResolveTCPAddr("tcp", testAddr)
	if err != nil {
		t.Fatalf("Failed to resolve valid address %s: %v", testAddr, err)
	}

	if addr.Port != 9999 {
		t.Errorf("Expected port 9999, got %d", addr.Port)
	}
}

func TestResolveTCPAddrInvalid(t *testing.T) {
	// Test case: Invalid address resolution should fail
	invalidAddresses := []string{
		"invalid:address:format",
		":99999999",            // Port too large
		"256.256.256.256:8080", // Invalid IP
	}

	for _, addr := range invalidAddresses {
		t.Run("invalid_"+addr, func(t *testing.T) {
			_, err := net.ResolveTCPAddr("tcp", addr)
			if err == nil {
				t.Errorf("Expected error when resolving invalid address %s, got nil", addr)
			}
		})
	}
}

func TestClientOptionConstruction(t *testing.T) {
	// Test case: Verify ClientOption construction with different values
	testCases := []struct {
		name         string
		sentinelAddr string
		sentinelUser string
		sentinelPass string
		master       string
	}{
		{
			name:         "Basic configuration",
			sentinelAddr: ":26379",
			sentinelUser: "",
			sentinelPass: "",
			master:       "mymaster",
		},
		{
			name:         "With authentication",
			sentinelAddr: "127.0.0.1:26379",
			sentinelUser: "admin",
			sentinelPass: "password123",
			master:       "redis-master",
		},
		{
			name:         "Custom port",
			sentinelAddr: "localhost:16379",
			sentinelUser: "user",
			sentinelPass: "pass",
			master:       "custom-master",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			clientOption := valkey.ClientOption{
				InitAddress: []string{tc.sentinelAddr},
				Username:    tc.sentinelUser,
				Password:    tc.sentinelPass,
				Sentinel: valkey.SentinelOption{
					MasterSet: tc.master,
				},
			}

			// Verify the client option is constructed correctly
			if len(clientOption.InitAddress) != 1 || clientOption.InitAddress[0] != tc.sentinelAddr {
				t.Errorf("Expected InitAddress [%s], got %v", tc.sentinelAddr, clientOption.InitAddress)
			}

			if clientOption.Username != tc.sentinelUser {
				t.Errorf("Expected Username %s, got %s", tc.sentinelUser, clientOption.Username)
			}

			if clientOption.Password != tc.sentinelPass {
				t.Errorf("Expected Password %s, got %s", tc.sentinelPass, clientOption.Password)
			}

			if clientOption.Sentinel.MasterSet != tc.master {
				t.Errorf("Expected MasterSet %s, got %s", tc.master, clientOption.Sentinel.MasterSet)
			}
		})
	}
}

// Test flag parsing behavior
func TestFlagParsing(t *testing.T) {
	// Test case: Verify flags can be set programmatically
	// Note: This test modifies global flags, so we need to be careful

	// Save original values
	origListen := *listen
	origMaster := *master
	origMaxProcs := *maxProcs

	// Set test values
	testListen := ":8888"
	testMaster := "testmaster"
	testMaxProcs := 2

	*listen = testListen
	*master = testMaster
	*maxProcs = testMaxProcs

	// Verify values are set
	if *listen != testListen {
		t.Errorf("Expected listen %s, got %s", testListen, *listen)
	}

	if *master != testMaster {
		t.Errorf("Expected master %s, got %s", testMaster, *master)
	}

	if *maxProcs != testMaxProcs {
		t.Errorf("Expected maxProcs %d, got %d", testMaxProcs, *maxProcs)
	}

	// Restore original values
	*listen = origListen
	*master = origMaster
	*maxProcs = origMaxProcs
}

// Benchmark test for address resolution
func BenchmarkResolveTCPAddr(b *testing.B) {
	addr := ":9999"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := net.ResolveTCPAddr("tcp", addr)
		if err != nil {
			b.Fatalf("Failed to resolve address: %v", err)
		}
	}
}

// Benchmark test for ClientOption creation
func BenchmarkClientOptionCreation(b *testing.B) {
	sentinelAddr := ":26379"
	sentinelUser := "user"
	sentinelPass := "pass"
	master := "mymaster"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = valkey.ClientOption{
			InitAddress: []string{sentinelAddr},
			Username:    sentinelUser,
			Password:    sentinelPass,
			Sentinel: valkey.SentinelOption{
				MasterSet: master,
			},
		}
	}
}
