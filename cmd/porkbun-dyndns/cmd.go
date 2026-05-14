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
	rootCmd.AddCommand(NewUpdateRecordCommand(app))
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

			fmt.Println(response.YourIp)

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

			fmt.Println(string(jsonData))

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

			domain, subdomain := SplitDomain(name)

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

			fmt.Println(string(jsonData))

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

func NewUpdateRecordCommand(app *App) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update-record --name <name> --type <type> --content <content>",
		Short: "Update a DNS record by name and type",
		RunE: func(cmd *cobra.Command, args []string) error {

			content, err := cmd.Flags().GetString("content")
			if err != nil {
				return err
			}
			name, err := cmd.Flags().GetString("name")
			if err != nil {
				return err
			}
			recordType, err := cmd.Flags().GetString("type")
			if err != nil {
				return err
			}
			notes, err := cmd.Flags().GetString("notes")
			if err != nil {
				return err
			}
			prio, err := cmd.Flags().GetInt("priority")
			if err != nil {
				return err
			}
			ttl, err := cmd.Flags().GetInt("ttl")
			if err != nil {
				return err
			}

			// parse domain and subdomain (www.example.com = example.com, www)
			domain, subdomain := SplitDomain(name)

			cmd.SilenceUsage = true

			// build the update payload
			update := &api.DNSRecordUpdate{
				Content: content,
			}
			if cmd.Flags().Changed("notes") {
				update.Notes = api.String(notes)
			}
			if cmd.Flags().Changed("prio") {
				update.Priority = api.Int(prio)
			}
			if cmd.Flags().Changed("ttl") {
				update.Ttl = api.Int(ttl)
			}

			// perform the update
			err = app.client.Dns.UpdateRecordByNameAndType(cmd.Context(), domain, subdomain, recordType, update)
			if err != nil {
				return err
			}
			return nil
		},
	}

	cmd.Flags().String("name", "", "Name of the record to retrieve (e.g., www.example.com or example.com)")
	cmd.Flags().String("type", "", "Type of the record to retrieve (A, MX, CNAME, etc.)")
	cmd.Flags().String("content", "", "The content of the record to set")
	cmd.Flags().String("notes", "", "Notes to store with the record (not served in DNS). Pass an empty string to clear existing notes; omit this field to leave notes unchanged.")
	cmd.Flags().Int("priority", 0, "Priority (optional)")
	cmd.Flags().Int("ttl", 0, "Time to live in seconds (optional)")

	_ = cmd.MarkFlagRequired("name")
	_ = cmd.MarkFlagRequired("type")
	_ = cmd.MarkFlagRequired("content")

	return cmd
}

// SplitDomain splits a fully qualified domain name into the domain and subdomain portions.
// For example, "www.example.com" would return "example.com", "www".
// If the name does not contain a subdomain, the subdomain will be an empty string.
func SplitDomain(name string) (domain string, subdomain string) {
	parts := strings.Split(name, ".")
	domain = name
	subdomain = ""
	if len(parts) > 2 {
		domain = strings.Join(parts[len(parts)-2:], ".")
		subdomain = strings.Join(parts[:len(parts)-2], ".")
	}
	return domain, subdomain
}
