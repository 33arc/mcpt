package mcp

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"
)

type Client struct {
	Host            string
	SSE             bool
	SID             string
	CTX             context.Context
	Cancel          context.CancelFunc
	HTTPClient      *http.Client
	ProtocolVersion string
}

type JSONRPCRequest struct {
	JSONRPC string      `json:"jsonrpc"`
	ID      int         `json:"id"`
	Method  string      `json:"method"`
	Params  interface{} `json:"params"`
}

func NewClient(host string, isSSE bool, protocolVersion string) *Client {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	httpClient := &http.Client{}

	if protocolVersion == "" {
		protocolVersion = "2025-06-18"
	}

	return &Client{
		Host:       host,
		SSE:        isSSE,
		SID:        "",
		CTX:        ctx,
		Cancel:     cancel,
		HTTPClient: httpClient,
		ProtocolVersion: protocolVersion
	}
}

func (c *Client) Call(tool, arguments string) {
	c.sendInitializeRequest()
	c.sendInitializedNotification()
	result := c.doOperation(tool, arguments)

	resultJSON, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		log.Fatal("Failed to marshal feature:", err)
	}

	fmt.Println(string(resultJSON))
}

func (c *Client) ListFeature(feature, output string) {
	c.sendInitializeRequest()
	c.sendInitializedNotification()
	features := c.doOperationList(feature)
	c.display(features, output)
}

func (c *Client) Ping() {
	c.sendInitializeRequest()
	c.sendInitializedNotification()
	c.ping()
}

func (c *Client) doOperation(tool, arguments string) map[string]interface{} {
	var args interface{}
	if err := json.Unmarshal([]byte(arguments), &args); err != nil {
		log.Fatal("Failed to parse arguments JSON:", err)
	}

	ping := map[string]interface{}{
		"jsonrpc": "2.0",
		"id":      "123",
		"method":  "tools/call",
		"params": map[string]interface{}{
			"name":      tool,
			"arguments": args,
		},
	}

	pingBytes, err := json.Marshal(ping)
	if err != nil {
		log.Fatal("Failed to marshal ping:", err)
	}

	pingReq, err := http.NewRequestWithContext(c.CTX, "POST", c.Host, bytes.NewBuffer(pingBytes))
	if err != nil {
		log.Fatal("Failed to create notification request:", err)
	}
	pingReq.Header.Set("Content-Type", "application/json")
	pingReq.Header.Set("Accept", "application/json")
	pingReq.Header.Set("Mcp-Session-Id", c.SID)

	pingResp, err := c.HTTPClient.Do(pingReq)
	if err != nil {
		log.Fatal("Ping request failed:", err)
	}
	defer pingResp.Body.Close()

	var respJSON map[string]interface{}
	if err := json.NewDecoder(pingResp.Body).Decode(&respJSON); err != nil {
		log.Fatal("Failed to decode JSON:", err)
	}

	return respJSON
}

func (c *Client) display(features []interface{}, output string) {
	if output == "json" {
		// marshal just the feature array back to JSON
		featureJSON, err := json.MarshalIndent(features, "", "  ")
		if err != nil {
			log.Fatal("Failed to marshal feature:", err)
		}
		fmt.Println(string(featureJSON))
	}
	if output == "call" {
		for _, f := range features { // features is a map, so _ = key, f = value
			fMap, ok := f.(map[string]interface{})
			if !ok {
				log.Fatal("feature element is not a map")
			}

			inputSchema, ok := fMap["inputSchema"].(map[string]interface{})
			if !ok {
				log.Fatal("inputSchema not found")
			}

			required, ok := inputSchema["required"].([]interface{})
			if !ok {
				log.Fatal("required not found")
			}

			properties, ok := inputSchema["properties"].(map[string]interface{})
			if !ok {
				log.Fatal("properties not found")
			}

			toolName := strings.TrimSpace(fMap["name"].(string))

			argument := ""
			requiredSet := makeSet(required)
			for _, key := range required {
				keyStr := strings.TrimSpace(key.(string))

				// type-assert the property to a map
				prop, ok := properties[keyStr].(map[string]interface{})
				if !ok {
					log.Fatalf("property %s is not a map", keyStr)
				}

				// now access the "type" field
				val, ok := prop["type"]
				if !ok {
					log.Fatalf("type not found for property %s", keyStr)
				}

				if val == "object" {
					argument += fmt.Sprintf("\"%s\":\033[31m<<%s>>\033[0m,", keyStr, keyStr+"Object")
				} else if val == "enum" {
					argument += fmt.Sprintf("\"%s\":\033[31m<<%s>>\033[0m,", keyStr, keyStr+"Enum")
				} else {
					argument += fmt.Sprintf("\"%s\":\033[31m<<%s>>\033[0m,", keyStr, val)
				}
			}

			argument = argument[:len(argument)-1] // remove the last character
			fmt.Printf("mcpt call --host '%s' --tool '%s' --arguments '{%s}'\n", c.Host, toolName, argument)
			traverseProperties(properties, "", requiredSet)
			fmt.Printf("\n")
		}
	}
}

func makeSet(arr []interface{}) map[string]struct{} {
	m := make(map[string]struct{}, len(arr))
	for _, v := range arr {
		if s, ok := v.(string); ok {
			m[s] = struct{}{}
		}
	}
	return m
}

func traverseProperties(properties map[string]interface{}, prefix string, required map[string]struct{}) {
	for key, val := range properties {
		fullKey := key
		if prefix != "" {
			fullKey = prefix + "." + key
		}

		propMap, ok := val.(map[string]interface{})
		if !ok {
			continue
		}

		// print type if exists
		if typeVal, ok := propMap["type"]; ok {
			var t string
			t = fmt.Sprintf("%v", typeVal) // convert interface{} → string
			if enumVal, ok := propMap["enum"]; ok {
				if enumArr, ok := enumVal.([]interface{}); ok {
					t = "enum("
					for i, e := range enumArr {
						if i > 0 {
							t += ","
						}
						t += "\"" + (e.(string)) + "\""
					}
					t += ")"
				}
			}
			if _, exists := required[fullKey]; exists {
				fmt.Printf("\033[31m[REQUIRED]\033[0m\033[34m %s -> type: %s\033[0m", fullKey, t)
			} else {
				fmt.Printf("           \033[34m%s -> type: %s\033[0m", fullKey, t)
			}
			if arrVal, ok := propMap["items"].(map[string]interface{}); ok {
				for key, val := range arrVal {
					// fmt.Printf("ArrItem %s", val)
					if key == "type" {
						fmt.Printf("\033[34m[ArrayItems -> type: %s]\033[0m", val)
					}
				}
				_, ok := propMap["uniqueItems"]
				if ok {
					fmt.Printf("[UNIQUE]")
				}
			}
			fmt.Printf("\n")

		}

		// if nested properties, recurse
		if nested, ok := propMap["properties"]; ok {
			if nestedProps, ok := nested.(map[string]interface{}); ok {
				traverseProperties(nestedProps, fullKey, required)
			}
		}
	}
}

func (c *Client) sendInitializeRequest() {
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
				"name":    "go-client",
				"version": "1.0.0",
			},
			"protocolVersion":c.ProtocolVersion,
		},
	}

	bodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		log.Fatal("Failed to marshal request:", err)
	}

	req, err := http.NewRequestWithContext(c.CTX, "POST", c.Host, bytes.NewBuffer(bodyBytes))
	if err != nil {
		log.Fatal("Failed to create request:", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		log.Fatal("Request failed:", err)
	}
	defer resp.Body.Close()

	sessionID := resp.Header.Get("Mcp-Session-Id")
	if sessionID == "" {
		log.Fatal("MCP-Session-ID not found in response headers")
	}
	c.SID = sessionID
}

func (c *Client) sendInitializedNotification() {
	// --- Second request: notifications/initialized ---
	notification := map[string]interface{}{
		"jsonrpc": "2.0",
		"method":  "notifications/initialized",
	}

	notificationBytes, err := json.Marshal(notification)
	if err != nil {
		log.Fatal("Failed to marshal notification:", err)
	}

	notifReq, err := http.NewRequestWithContext(c.CTX, "POST", c.Host, bytes.NewBuffer(notificationBytes))
	if err != nil {
		log.Fatal("Failed to create notification request:", err)
	}
	notifReq.Header.Set("Content-Type", "application/json")
	notifReq.Header.Set("Accept", "application/json")
	notifReq.Header.Set("Mcp-Session-Id", c.SID)

	notifResp, err := c.HTTPClient.Do(notifReq)
	if err != nil {
		log.Fatal("Notification request failed:", err)
	}
	defer notifResp.Body.Close()
}

func (c *Client) ping() {
	ping := map[string]interface{}{
		"jsonrpc": "2.0",
		"id":      "123",
		"method":  "ping",
	}

	pingBytes, err := json.Marshal(ping)
	if err != nil {
		log.Fatal("Failed to marshal ping:", err)
	}

	pingReq, err := http.NewRequestWithContext(c.CTX, "POST", c.Host, bytes.NewBuffer(pingBytes))
	if err != nil {
		log.Fatal("Failed to create notification request:", err)
	}
	pingReq.Header.Set("Content-Type", "application/json")
	pingReq.Header.Set("Accept", "application/json")
	pingReq.Header.Set("Mcp-Session-Id", c.SID)

	pingResp, err := c.HTTPClient.Do(pingReq)
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
}

func (c *Client) doOperationList(feature string) []interface{} {
	method := feature + "/list"
	ping := map[string]interface{}{
		"jsonrpc": "2.0",
		"id":      "123",
		"method":  method,
	}

	pingBytes, err := json.Marshal(ping)
	if err != nil {
		log.Fatal("Failed to marshal ping:", err)
	}

	pingReq, err := http.NewRequestWithContext(c.CTX, "POST", c.Host, bytes.NewBuffer(pingBytes))
	if err != nil {
		log.Fatal("Failed to create notification request:", err)
	}
	pingReq.Header.Set("Content-Type", "application/json")
	pingReq.Header.Set("Accept", "application/json")
	pingReq.Header.Set("Mcp-Session-Id", c.SID)

	pingResp, err := c.HTTPClient.Do(pingReq)
	if err != nil {
		log.Fatal("Ping request failed:", err)
	}
	defer pingResp.Body.Close()

	var respJSON map[string]interface{}
	if err := json.NewDecoder(pingResp.Body).Decode(&respJSON); err != nil {
		log.Fatal("Failed to decode JSON:", err)
	}

	// drill into result
	result, ok := respJSON["result"].(map[string]interface{})
	if !ok {
		log.Fatal("result not found or wrong type")
	}

	// drill into feature
	features, ok := result[feature].([]interface{})
	if !ok {
		log.Fatal("feature not found or wrong type")
	}

	return features

}
