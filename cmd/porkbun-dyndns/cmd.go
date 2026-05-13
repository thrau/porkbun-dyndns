package main

import (
	"context"
	"encoding/json"

	"github.com/spf13/cobra"
	"github.com/thrau/porkbun-dyndns/internal/api"
)

var rootCmd = &cobra.Command{
	Use:   "porkbun-dyndns",
	Short: "Dynamic DNS utilities for Porkbun",
}

type App struct {
	client *api.Client
}

func init() {
	app := &App{client: api.NewClient(api.WithEnvironmentCredentials())}

	rootCmd.AddCommand(NewMyIpCommand(app))
	rootCmd.AddCommand(NewListRecordsCommand(app))
}

func NewMyIpCommand(app *App) *cobra.Command {
	return &cobra.Command{
		Use:   "myip",
		Short: "Get your public IP address",
		RunE: func(cmd *cobra.Command, args []string) error {
			response, err := app.client.Util.Ip(context.TODO())

			if err != nil {
				return err
			}

			cmd.Println(response.YourIp)

			return nil
		},
	}
}

func NewListRecordsCommand(app *App) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list-records",
		Short: "List DNS records for a domain",

		RunE: func(cmd *cobra.Command, args []string) error {
			domain, err := cmd.Flags().GetString("domain")
			if err != nil {
				return err
			}
			cmd.SilenceUsage = true

			records, err := app.client.Dns.ListRecords(context.TODO(), domain)
			if err != nil {
				return err
			}

			jsonData, err := json.MarshalIndent(records, "", "  ")
			if err != nil {
				return err
			}

			cmd.Println(string(jsonData))

			return nil
		},
	}

	cmd.Flags().String("domain", "", "Domain name to list records for")
	_ = cmd.MarkFlagRequired("domain")

	return cmd
}
