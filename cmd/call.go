/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"log"

	"github.com/33arc/mcpt/mcp"
	"github.com/spf13/cobra"
)

var tool string
var arguments string

// callCmd represents the call command
var callCmd = &cobra.Command{
	Use:   "call",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		if tool == "" {
			log.Fatal("Missing --tool JSON string")
		}

		client := mcp.NewClient(host, false, protocolVersion)
		client.Call(tool, arguments)
	},
}

func init() {
	callCmd.Flags().StringVar(&tool, "tool", "", "Tool name")
	callCmd.Flags().StringVar(&arguments, "arguments", "", "Json file containing arguments")
	rootCmd.AddCommand(callCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// callCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// callCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
