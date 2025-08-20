/*
Copyright © 2025 33arc
*/
package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

var host string
var output string
var protocolVersion string
var sseEnabled bool

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "mcpt",
	Short: "A brief description of your application",
	Long: `A longer description that spans multiple lines and likely contains
examples and usage of using your application. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	// Uncomment the following line if your bare application
	// has an action associated with it:
	// Run: func(cmd *cobra.Command, args []string) { },
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().BoolVar(&sseEnabled, "sse", false, "enable SSE")
	rootCmd.PersistentFlags().StringVar(&output, "output", "json", "Which output to use")
	rootCmd.PersistentFlags().StringVar(&host, "host", "http://localhost:8080/mcp", "MCP server URL")
	rootCmd.PersistentFlags().StringVar(&protocolVersion, "protocol-version", "2025-06-18", "MCP protocol version")
	rootCmd.MarkPersistentFlagRequired("host")
	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.

	// rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.mcpt.yaml)")

	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
