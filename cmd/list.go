/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/spf13/cobra"
)

// listCmd represents the list command
var listCmd = &cobra.Command{
	Use:   "list",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
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

		// --- Third request: ping

		ping := map[string]interface{}{
			"jsonrpc": "2.0",
			"id":      "123",
			"method":  "tools/list",
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

		var respJSON map[string]interface{}
		if err := json.NewDecoder(resp.Body).Decode(&respJSON); err != nil {
			log.Fatal("Failed to decode JSON:", err)
		}

		pretty, _ := json.MarshalIndent(respJSON, "", "  ")
		fmt.Println(string(pretty))

	},
}

func init() {
	listCmd.Flags().StringVar(&host, "host", "http://localhost:8080/mcp", "MCP server URL")
	toolsCmd.AddCommand(listCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// listCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// listCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
