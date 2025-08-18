package cmd

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/spf13/cobra"
)

type JSONRPCRequest struct {
	JSONRPC string      `json:"jsonrpc"`
	ID      int         `json:"id"`
	Method  string      `json:"method"`
	Params  interface{} `json:"params"`
}

var host string

var pingCmd = &cobra.Command{
	Use:   "ping",
	Short: "Ping the MCP server over HTTP",
	Long:  "Send a ping request to the MCP server over HTTP to verify connectivity.",
	Run: func(cmd *cobra.Command, args []string) {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		url := host

		// --- First request: initialize ---
		reqBody := JSONRPCRequest{
			JSONRPC: "2.0",
			ID:      1,
			Method:  "initialize",
			Params: map[string]interface{}{
				"capabilities": map[string]interface{}{
					"textDocument": map[string]interface{}{
						"synchronization": map[string]bool{"didSave": true},
					},
				},
				"clientInfo": map[string]interface{}{
					"name":    "hurl-client",
					"version": "1.0.0",
				},
				"protocolVersion": "2.0",
			},
		}

		bodyBytes, err := json.Marshal(reqBody)
		if err != nil {
			log.Fatal("Failed to marshal request:", err)
		}

		req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(bodyBytes))
		if err != nil {
			log.Fatal("Failed to create request:", err)
		}
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Accept", "application/json")

		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			log.Fatal("Request failed:", err)
		}
		defer resp.Body.Close()

		sessionID := resp.Header.Get("Mcp-Session-Id")
		if sessionID == "" {
			log.Fatal("MCP-Session-ID not found in response headers")
		}
		log.Println("MCP-Session-ID:", sessionID)

		// --- Second request: notifications/initialized ---
		notification := map[string]interface{}{
			"jsonrpc": "2.0",
			"method":  "notifications/initialized",
		}

		notificationBytes, err := json.Marshal(notification)
		if err != nil {
			log.Fatal("Failed to marshal notification:", err)
		}

		notifReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(notificationBytes))
		if err != nil {
			log.Fatal("Failed to create notification request:", err)
		}
		notifReq.Header.Set("Content-Type", "application/json")
		notifReq.Header.Set("Accept", "application/json")
		notifReq.Header.Set("Mcp-Session-Id", sessionID)

		notifResp, err := client.Do(notifReq)
		if err != nil {
			log.Fatal("Notification request failed:", err)
		}
		defer notifResp.Body.Close()

		log.Println("Sent notifications/initialized, response status:", notifResp.Status)

		log.Println("Ready for the operation phase...")
		// --- Third request: ping

		ping := map[string]interface{}{
			"jsonrpc": "2.0",
			"id":      "123",
			"method":  "ping",
		}

		pingBytes, err := json.Marshal(ping)
		if err != nil {
			log.Fatal("Failed to marshal ping:", err)
		}

		pingReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(pingBytes))
		if err != nil {
			log.Fatal("Failed to create notification request:", err)
		}
		pingReq.Header.Set("Content-Type", "application/json")
		pingReq.Header.Set("Accept", "application/json")
		pingReq.Header.Set("Mcp-Session-Id", sessionID)

		pingResp, err := client.Do(pingReq)
		if err != nil {
			log.Fatal("Ping request failed:", err)
		}
		defer pingResp.Body.Close()

		respBody, err := io.ReadAll(pingResp.Body)
		if err != nil {
			log.Fatal("Failed to read response body:", err)
		}

		var respJSON map[string]interface{}
		if err := json.Unmarshal(respBody, &respJSON); err != nil {
			log.Fatal("Failed to unmarshal response:", err)
		}
		expectedID := "123"

		if respJSON["jsonrpc"] != "2.0" {
			log.Fatal("Unexpected jsonrpc value")
		}

		if respJSON["id"] != expectedID {
			log.Fatal("Unexpected id value")
		}

		if result, ok := respJSON["result"].(map[string]interface{}); !ok || len(result) != 0 {
			log.Fatal("Unexpected result value")
		}

		log.Println("Ping OK ✅")
	},
}

func init() {
	pingCmd.Flags().StringVar(&host, "host", "http://localhost:8080/mcp", "MCP server URL")
	rootCmd.AddCommand(pingCmd)
}
