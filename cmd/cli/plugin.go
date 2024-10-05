package main

// import (
// 	"fmt"

// 	"github.com/spf13/cobra"
// )

// var pluginCmd = &cobra.Command{
// 	Use:   "plugin",
// 	Short: "Manage plugins",
// 	Run: func(cmd *cobra.Command, args []string) {
// 		cmd.Help()
// 	},
// }

// var pluginListCmd = &cobra.Command{
// 	Use:   "list",
// 	Short: "List all plugins",
// 	Run: func(cmd *cobra.Command, args []string) {
// 		fmt.Println("Listing plugins...")
// 		// TODO: Implement listing plugins
// 	},
// }

// var pluginEnableCmd = &cobra.Command{
// 	Use:   "enable [plugin name]",
// 	Short: "Enable a plugin",
// 	Args:  cobra.ExactArgs(1),
// 	Run: func(cmd *cobra.Command, args []string) {
// 		pluginName := args[0]
// 		fmt.Printf("Enabling plugin '%s'...\n", pluginName)
// 		// TODO: Implement enabling plugin
// 	},
// }

// var pluginDisableCmd = &cobra.Command{
// 	Use:   "disable [plugin name]",
// 	Short: "Disable a plugin",
// 	Args:  cobra.ExactArgs(1),
// 	Run: func(cmd *cobra.Command, args []string) {
// 		pluginName := args[0]
// 		fmt.Printf("Disabling plugin '%s'...\n", pluginName)
// 		// TODO: Implement disabling plugin
// 	},
// }
