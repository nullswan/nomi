package main

import (
	"log"

	"github.com/nullswan/nomi/internal/updater"
	"github.com/spf13/cobra"
)

const binaryName = "nomi"

var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "Automatically update the application",
	Run: func(_ *cobra.Command, _ []string) {
		config := updater.Config{
			Repository:     "nullswan/nomi",
			CurrentVersion: buildVersion,
			BinaryName:     binaryName,
		}

		up := updater.New(config)
		if err := up.Update(); err != nil {
			log.Fatalf("Update failed: %v", err)
		}
	},
}
