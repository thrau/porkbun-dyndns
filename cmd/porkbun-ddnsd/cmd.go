package main

import (
	"fmt"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "porkbun-dyndnsd",
	Short: "Dynamic DNS daemon for Porkbun",
}

func init() {
	// pass
	rootCmd.RunE = func(cmd *cobra.Command, args []string) error {
		return fmt.Errorf("not implemented yet")
	}
}
