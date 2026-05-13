package main

import (
	"encoding/json"
	"fmt"
	"strings"

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
	rootCmd.AddCommand(NewGetRecordCommand(app))
}

func NewMyIpCommand(app *App) *cobra.Command {
	return &cobra.Command{
		Use:   "myip",
		Short: "Get your public IP address",
		RunE: func(cmd *cobra.Command, args []string) error {
			response, err := app.client.Util.Ip(cmd.Context())

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

			records, err := app.client.Dns.ListRecords(cmd.Context(), domain)
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

func NewGetRecordCommand(app *App) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get-record --name <name> [--id <id> | --type <type>]",
		Short: "Get specific DNS records",
		RunE: func(cmd *cobra.Command, args []string) error {
			name, err := cmd.Flags().GetString("name")
			if err != nil {
				return err
			}

			parts := strings.Split(name, ".")
			domain := name
			subdomain := ""
			if len(parts) > 2 {
				domain = strings.Join(parts[len(parts)-2:], ".")
				subdomain = strings.Join(parts[:len(parts)-2], ".")
			}

			id, _ := cmd.Flags().GetString("id")
			recordType, _ := cmd.Flags().GetString("type")

			cmd.SilenceUsage = true

			var records []api.DNSRecord
			if id != "" {
				records, err = app.client.Dns.GetRecordById(cmd.Context(), domain, id)
			} else if recordType != "" {
				if subdomain != "" {
					records, err = app.client.Dns.GetRecordByNameAndType(cmd.Context(), domain, subdomain, recordType)
				} else {
					records, err = app.client.Dns.GetRecordByType(cmd.Context(), domain, recordType)
				}
			} else {
				return fmt.Errorf("either --id or --type must be provided")
			}

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

	cmd.Flags().String("name", "", "Name of the record to retrieve (e.g., www.example.com or example.com)")
	_ = cmd.MarkFlagRequired("name")

	// the records can be retrieved either by id
	cmd.Flags().String("id", "", "ID of the record to retrieve (if set, a subdomain in the name will be ignored)")

	// or by type
	cmd.Flags().String("type", "", "Type of the record to retrieve (A, MX, CNAME, etc.)")

	cmd.MarkFlagsMutuallyExclusive("id", "type")

	return cmd
}
