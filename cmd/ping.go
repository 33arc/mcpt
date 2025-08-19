package cmd

import (
	"github.com/33arc/mcpt/mcp"
	"github.com/spf13/cobra"
)

var pingCmd = &cobra.Command{
	Use:   "ping",
	Short: "Ping the MCP server over HTTP",
	Long:  "Send a ping request to the MCP server over HTTP to verify connectivity.",

	Run: func(cmd *cobra.Command, args []string) {
		client := mcp.NewClient(host, false)
		client.Ping()
	},
}

func init() {
	pingCmd.Flags().StringVar(&host, "host", "http://localhost:8080/mcp", "MCP server URL")
	rootCmd.AddCommand(pingCmd)
}
