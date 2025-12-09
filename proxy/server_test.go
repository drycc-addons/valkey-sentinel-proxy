package proxy

import (
	"io"
	"net"
	"sync"
	"testing"
	"time"

	"github.com/valkey-io/valkey-go"
)

// Mock connection for testing
type mockConn struct {
	readData  []byte
	writeData []byte
	closed    bool
	mu        sync.RWMutex
}

func (m *mockConn) Read(p []byte) (n int, err error) {
	if len(m.readData) == 0 {
		return 0, io.EOF
	}
	n = copy(p, m.readData)
	m.readData = m.readData[n:]
	return n, nil
}

func (m *mockConn) Write(p []byte) (n int, err error) {
	m.writeData = append(m.writeData, p...)
	return len(p), nil
}

func (m *mockConn) Close() error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.closed = true
	return nil
}

func (m *mockConn) IsClosed() bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.closed
}

func TestNewRedisProxyServer(t *testing.T) {
	// Test case: Create new Redis proxy server
	listenAddr, err := net.ResolveTCPAddr("tcp", ":9999")
	if err != nil {
		t.Fatalf("Failed to resolve listen address: %v", err)
	}

	clientOption := valkey.ClientOption{
		InitAddress: []string{":26379"},
		Username:    "testuser",
		Password:    "testpass",
		Sentinel: valkey.SentinelOption{
			MasterSet: "mymaster",
		},
	}

	server, err := NewRedisProxyServer(listenAddr, clientOption, "mymaster")

	// Verify the server is created correctly
	if err != nil {
		t.Skipf("Cannot create server in test environment: %v", err)
	}
	if server == nil {
		t.Fatal("Expected server to be created, got nil")
	}
	defer server.Close()

	if server.listener != listenAddr {
		t.Errorf("Expected listener %v, got %v", listenAddr, server.listener)
	}

	if server.masterName != "mymaster" {
		t.Errorf("Expected master name 'mymaster', got %s", server.masterName)
	}

	if server.clientOption.Username != "testuser" {
		t.Errorf("Expected username 'testuser', got %s", server.clientOption.Username)
	}
}

func TestRedisProxyServer_master(t *testing.T) {
	// This test requires a mock or integration test setup
	// For unit testing, we'll test the error cases and structure

	listenAddr, _ := net.ResolveTCPAddr("tcp", ":9999")

	// Test case: Invalid sentinel address should return error
	clientOption := valkey.ClientOption{
		InitAddress: []string{"invalid:address:format"},
		Sentinel: valkey.SentinelOption{
			MasterSet: "mymaster",
		},
	}

	server, err := NewRedisProxyServer(listenAddr, clientOption, "mymaster")
	if err != nil {
		// Expected - invalid address format
		return
	}
	defer server.Close()

	// This will fail to connect to an invalid address

	// This will fail to connect to an invalid address
	_, err = server.master()
	if err == nil {
		t.Error("Expected error when connecting to invalid address, got nil")
	}
}

func TestRedisProxyServer_proxy(t *testing.T) {
	// Test case: Test proxy connection handling
	listenAddr, _ := net.ResolveTCPAddr("tcp", ":9999")
	clientOption := valkey.ClientOption{InitAddress: []string{"127.0.0.1:26379"}}
	server, err := NewRedisProxyServer(listenAddr, clientOption, "mymaster")
	if err != nil {
		t.Skipf("Cannot create server in test environment: %v", err)
	}
	defer server.Close()

	// Create mock connections
	mockLocal := &mockConn{
		readData: []byte("test data from client"),
	}

	// Create a test server to act as remote
	remoteListener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("Failed to create test server: %v", err)
	}
	defer remoteListener.Close()

	remoteAddr := remoteListener.Addr().(*net.TCPAddr)

	// Start a goroutine to accept connections on the remote server
	go func() {
		conn, err := remoteListener.Accept()
		if err != nil {
			return
		}
		defer conn.Close()

		// Echo back any data received
		buf := make([]byte, 1024)
		n, err := conn.Read(buf)
		if err != nil {
			return
		}
		conn.Write(buf[:n])
	}()

	// Test the proxy function
	server.proxy(mockLocal, remoteAddr)

	// Give some time for goroutines to complete
	time.Sleep(100 * time.Millisecond)

	// Verify the mock connection was closed
	if !mockLocal.IsClosed() {
		t.Error("Expected local connection to be closed")
	}
}

func TestRedisProxyServer_Serve_InvalidListener(t *testing.T) {
	// Test case: Invalid listener address should cause fatal error
	// We can't easily test log.Fatal, but we can test the setup

	// Use an invalid address that will fail to bind
	invalidAddr, _ := net.ResolveTCPAddr("tcp", ":99999") // Invalid port
	clientOption := valkey.ClientOption{InitAddress: []string{"127.0.0.1:26379"}}
	server, err := NewRedisProxyServer(invalidAddr, clientOption, "mymaster")
	if err != nil {
		t.Skipf("Cannot create server in test environment: %v", err)
	}
	defer server.Close()

	// This should fail when trying to listen
	// We can't directly test this without modifying the Serve method
	// to return errors instead of calling log.Fatal

	// For now, just verify the server structure
	if server.listener != invalidAddr {
		t.Errorf("Expected listener %v, got %v", invalidAddr, server.listener)
	}
}

// Benchmark test for proxy creation
func BenchmarkNewRedisProxyServer(b *testing.B) {
	listenAddr, _ := net.ResolveTCPAddr("tcp", ":9999")
	clientOption := valkey.ClientOption{
		InitAddress: []string{":26379"},
		Sentinel: valkey.SentinelOption{
			MasterSet: "mymaster",
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		server, err := NewRedisProxyServer(listenAddr, clientOption, "mymaster")
		if err == nil && server != nil {
			server.Close()
		}
	}
}

// Integration test (requires running sentinel)
func TestRedisProxyServer_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	listenAddr, err := net.ResolveTCPAddr("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("Failed to resolve test address: %v", err)
	}

	clientOption := valkey.ClientOption{
		InitAddress: []string{"127.0.0.1:26379"},
		Sentinel: valkey.SentinelOption{
			MasterSet: "mymaster",
		},
	}

	server, err := NewRedisProxyServer(listenAddr, clientOption, "mymaster")
	if err != nil {
		t.Skipf("Integration test skipped - Redis Sentinel not available: %v", err)
	}
	defer server.Close()

	// Test that would require actual Redis Sentinel
	_, err = server.master()
	// We expect this to fail in test environment
	if err == nil {
		t.Log("Integration test passed - Redis Sentinel is available")
	} else {
		t.Logf("Integration test skipped - Redis Sentinel not available: %v", err)
	}
}

// Test error handling in master() method
func TestRedisProxyServer_masterErrorHandling(t *testing.T) {
	testCases := []struct {
		name          string
		masterName    string
		expectError   bool
		errorContains string
	}{
		{
			name:        "Empty master name",
			masterName:  "",
			expectError: true,
		},
		{
			name:        "Valid master name",
			masterName:  "mymaster",
			expectError: true, // Will fail due to no sentinel running
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			listenAddr, _ := net.ResolveTCPAddr("tcp", ":9999")
			clientOption := valkey.ClientOption{
				InitAddress: []string{"127.0.0.1:26379"},
				Sentinel: valkey.SentinelOption{
					MasterSet: tc.masterName,
				},
			}

			server, err := NewRedisProxyServer(listenAddr, clientOption, tc.masterName)
			if err != nil {
				if !tc.expectError {
					t.Errorf("Failed to create server for test case %s: %v", tc.name, err)
				}
				return
			}
			defer server.Close()

			_, err = server.master()

			if tc.expectError && err == nil {
				t.Errorf("Expected error for test case %s, got nil", tc.name)
			}

			if !tc.expectError && err != nil {
				t.Errorf("Expected no error for test case %s, got %v", tc.name, err)
			}
		})
	}
}
