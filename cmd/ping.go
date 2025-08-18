package cmd

import (
	"bytes"
	"context"
	"encoding/json"
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

var pingCmd = &cobra.Command{
	Use:   "ping",
	Short: "Ping the MCP server over HTTP",
	Long:  "Send a ping request to the MCP server over HTTP to verify connectivity.",
	Run: func(cmd *cobra.Command, args []string) {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		url := "http://localhost:8080/mcp"

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
	},
}

func init() {
	rootCmd.AddCommand(pingCmd)
}
