package main

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"io"
	"os"
	"os/exec"
	"testing"
	"time"
)

// TestMCPProtocol tests the MCP protocol integration
func TestMCPProtocol(t *testing.T) {
	// Skip if no config is available
	if os.Getenv("OED_DATA_PATH") == "" {
		t.Skip("Skipping integration test: OED_DATA_PATH not set")
	}

	// Build the server
	cmd := exec.Command("go", "build", "-o", "test_oedmcp", ".")
	if err := cmd.Run(); err != nil {
		t.Fatalf("Failed to build server: %v", err)
	}
	defer os.Remove("test_oedmcp")

	t.Run("Initialize", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		request := map[string]interface{}{
			"jsonrpc": "2.0",
			"id":      1,
			"method":  "initialize",
			"params": map[string]interface{}{
				"protocolVersion": "0.1.0",
				"capabilities":    map[string]interface{}{},
				"clientInfo": map[string]interface{}{
					"name":    "test",
					"version": "1.0",
				},
			},
		}

		response := runServerCommand(t, ctx, request)

		// Check response structure
		if response["jsonrpc"] != "2.0" {
			t.Error("Invalid jsonrpc version")
		}
		if response["id"] != float64(1) {
			t.Error("Invalid id")
		}

		result, ok := response["result"].(map[string]interface{})
		if !ok {
			t.Fatal("Missing or invalid result")
		}

		// Check server info
		serverInfo, ok := result["serverInfo"].(map[string]interface{})
		if !ok {
			t.Fatal("Missing serverInfo")
		}
		if serverInfo["name"] != "OED MCP Server" {
			t.Errorf("Wrong server name: %v", serverInfo["name"])
		}
	})

	t.Run("ListTools", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		// Initialize first
		initRequest := map[string]interface{}{
			"jsonrpc": "2.0",
			"id":      1,
			"method":  "initialize",
			"params":  map[string]interface{}{},
		}
		runServerCommand(t, ctx, initRequest)

		// List tools
		request := map[string]interface{}{
			"jsonrpc": "2.0",
			"id":      2,
			"method":  "tools/list",
			"params":  map[string]interface{}{},
		}

		response := runServerCommand(t, ctx, request)

		result, ok := response["result"].(map[string]interface{})
		if !ok {
			t.Fatal("Missing or invalid result")
		}

		tools, ok := result["tools"].([]interface{})
		if !ok {
			t.Fatal("Missing tools list")
		}

		// Check we have the expected tools
		expectedTools := []string{
			"oed_lookup",
			"oed_etymology",
			"oed_search",
			"oed_random",
			"oed_multi_lookup",
		}

		foundTools := make(map[string]bool)
		for _, tool := range tools {
			toolMap, ok := tool.(map[string]interface{})
			if !ok {
				continue
			}
			name, _ := toolMap["name"].(string)
			foundTools[name] = true
		}

		for _, expected := range expectedTools {
			if !foundTools[expected] {
				t.Errorf("Missing tool: %s", expected)
			}
		}
	})
}

// runServerCommand runs the server with a JSON-RPC request and returns the response
func runServerCommand(t *testing.T, ctx context.Context, request map[string]interface{}) map[string]interface{} {
	cmd := exec.CommandContext(ctx, "./test_oedmcp")

	// Setup pipes
	stdin, err := cmd.StdinPipe()
	if err != nil {
		t.Fatal(err)
	}
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		t.Fatal(err)
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		t.Fatal(err)
	}

	// Start the command
	if err := cmd.Start(); err != nil {
		t.Fatal(err)
	}

	// Send request
	encoder := json.NewEncoder(stdin)
	if err := encoder.Encode(request); err != nil {
		t.Fatal(err)
	}
	stdin.Close()

	// Read response
	var response map[string]interface{}
	decoder := json.NewDecoder(stdout)
	if err := decoder.Decode(&response); err != nil {
		// Read stderr for debugging
		stderrBytes, _ := io.ReadAll(stderr)
		t.Fatalf("Failed to decode response: %v\nStderr: %s", err, stderrBytes)
	}

	// Wait for process to exit
	cmd.Wait()

	return response
}

// TestConfigLoading tests configuration loading
func TestConfigLoading(t *testing.T) {
	t.Run("EnvironmentVariables", func(t *testing.T) {
		// This test requires actual OED files or will be skipped
		if os.Getenv("OED_DATA_PATH") == "" {
			t.Skip("OED_DATA_PATH not set")
		}

		cmd := exec.Command("go", "run", ".")
		cmd.Env = append(os.Environ(),
			"OED_DATA_PATH="+os.Getenv("OED_DATA_PATH"),
			"OED_INDEX_PATH="+os.Getenv("OED_INDEX_PATH"),
		)

		stdin, _ := cmd.StdinPipe()
		stdout, _ := cmd.StdoutPipe()

		if err := cmd.Start(); err != nil {
			t.Fatal(err)
		}

		// Send initialize request
		request := `{"jsonrpc":"2.0","id":1,"method":"initialize","params":{}}`
		stdin.Write([]byte(request + "\n"))
		stdin.Close()

		// Read response
		scanner := bufio.NewScanner(stdout)
		if scanner.Scan() {
			response := scanner.Text()
			if !bytes.Contains([]byte(response), []byte(`"result"`)) {
				t.Error("Server failed to initialize with env vars")
			}
		}

		cmd.Process.Kill()
		cmd.Wait()
	})

	t.Run("MissingConfig", func(t *testing.T) {
		// Build the server first
		buildCmd := exec.Command("go", "build", "-o", "test_missing_config", ".")
		if err := buildCmd.Run(); err != nil {
			t.Skip("Failed to build server")
		}
		defer os.Remove("test_missing_config")

		// Clear environment variables and run from temp dir
		cmd := exec.Command("./test_missing_config")
		cmd.Env = []string{"PATH=" + os.Getenv("PATH")}

		// Create temp directory with no config
		tempDir := t.TempDir()
		// Copy binary to temp dir
		exec.Command("cp", "test_missing_config", tempDir+"/test_missing_config").Run()
		cmd = exec.Command(tempDir + "/test_missing_config")
		cmd.Env = []string{"PATH=" + os.Getenv("PATH")}
		cmd.Dir = tempDir

		output, err := cmd.CombinedOutput()
		if err == nil {
			t.Error("Server should fail with missing config")
		}

		if !bytes.Contains(output, []byte("Failed to load configuration")) &&
		   !bytes.Contains(output, []byte("not configured")) {
			t.Errorf("Unexpected error message: %s", output)
		}
	})
}

// TestEndToEnd performs end-to-end testing if OED data is available
func TestEndToEnd(t *testing.T) {
	// Skip if no real OED data
	dataPath := os.Getenv("OED_DATA_PATH")
	indexPath := os.Getenv("OED_INDEX_PATH")

	if dataPath == "" || indexPath == "" {
		// Try config file
		configFile := "oed_config.json"
		if data, err := os.ReadFile(configFile); err == nil {
			var config struct {
				DataPath  string `json:"data_path"`
				IndexPath string `json:"index_path"`
			}
			if json.Unmarshal(data, &config) == nil {
				dataPath = config.DataPath
				indexPath = config.IndexPath
			}
		}
	}

	if dataPath == "" || indexPath == "" {
		t.Skip("Skipping end-to-end test: no OED data configured")
	}

	// Check if files actually exist
	if _, err := os.Stat(dataPath); os.IsNotExist(err) {
		t.Skip("OED data file not found")
	}
	if _, err := os.Stat(indexPath); os.IsNotExist(err) {
		t.Skip("OED index file not found")
	}

	// Build and run actual lookup test
	cmd := exec.Command("go", "build", "-o", "test_e2e", ".")
	if err := cmd.Run(); err != nil {
		t.Fatal(err)
	}
	defer os.Remove("test_e2e")

	// Test actual word lookup
	testCmd := exec.Command("./test_e2e")
	stdin, _ := testCmd.StdinPipe()
	stdout, _ := testCmd.StdoutPipe()

	if err := testCmd.Start(); err != nil {
		t.Fatal(err)
	}

	// Initialize
	stdin.Write([]byte(`{"jsonrpc":"2.0","id":1,"method":"initialize","params":{}}` + "\n"))

	// Lookup "dictionary"
	stdin.Write([]byte(`{"jsonrpc":"2.0","id":2,"method":"tools/call","params":{"name":"oed_lookup","arguments":{"word":"dictionary"}}}` + "\n"))
	stdin.Close()

	// Read responses
	scanner := bufio.NewScanner(stdout)
	responses := []string{}
	for scanner.Scan() {
		responses = append(responses, scanner.Text())
	}

	testCmd.Wait()

	if len(responses) < 2 {
		t.Fatal("Not enough responses")
	}

	// Check that we got a real dictionary entry
	lookupResponse := responses[1]
	if !bytes.Contains([]byte(lookupResponse), []byte("dictionary")) {
		t.Error("Lookup response doesn't contain 'dictionary'")
	}
}