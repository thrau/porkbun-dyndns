package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/knadh/koanf"
	"github.com/knadh/koanf/parsers/toml"
	"github.com/knadh/koanf/providers/file"
	"github.com/spf13/cobra"
	"github.com/thrau/porkbun-dyndns/internal/api"
	"github.com/thrau/porkbun-dyndns/internal/daemon"
	"github.com/thrau/porkbun-dyndns/internal/util"
)

const defaultConfigPath = "/etc/porkbun-ddnsd/config.toml"

func loadConfig(settings *koanf.Koanf, cfg *daemon.Config) (err error) {
	name := settings.String("name")
	if name == "" {
		return fmt.Errorf("name is required in configuration")
	}
	domain, subdomain := util.SplitDomain(name)

	cfg.Domain = domain
	cfg.Subdomain = subdomain

	intervalStr := settings.String("interval")
	if intervalStr != "" {
		interval, err := time.ParseDuration(intervalStr)
		if err != nil {
			return fmt.Errorf("invalid interval format: %s", intervalStr)
		}
		cfg.Interval = interval
	}

	return nil
}

var rootCmd = &cobra.Command{
	Use:   "porkbun-ddnsd",
	Short: "Dynamic DNS daemon for Porkbun",
	RunE: func(cmd *cobra.Command, args []string) (err error) {
		// get config path
		path, err := cmd.Flags().GetString("config")
		if err != nil {
			return err
		}
		// check if config exists
		stat, err := os.Stat(path)
		if err != nil {
			return err
		}
		if stat.IsDir() {
			return fmt.Errorf("config path %s is a directory", path)
		}
		cmd.SilenceUsage = true

		fmt.Printf("[%s] loading daemon config from %s\n", time.Now().Format(time.RFC3339), path)
		// parse config file
		settings := koanf.New(".")
		err = settings.Load(file.Provider(path), toml.Parser())
		if err != nil {
			return err
		}

		// parse into daemon config and then validate
		cfg := daemon.Config{
			Interval: 15 * time.Minute,
		}
		err = loadConfig(settings, &cfg)
		if err != nil {
			return err
		}
		if cfg.Domain == "" {
			return fmt.Errorf("domain is required")
		}
		if cfg.Interval <= 0 {
			return fmt.Errorf("interval must be greater than 0")
		}

		var apiKey, secretKey string
		// TODO: get credentials alternatively from systemd encrypted credentials, or from environment
		//  (PORKBUN_API_KEY and PORKBUN_SECRET_KEY)
		apiKey = settings.String("api_key")
		secretKey = settings.String("secret_key")

		if apiKey == "" || secretKey == "" {
			return fmt.Errorf("porkbun API key and secret key are required")
		}
		client := api.NewClient(api.WithCredentials(apiKey, secretKey))

		d := daemon.NewDaemon(
			client.Dns,
			client.Util,
			&cfg,
		)

		shutdownSignals := make(chan os.Signal, 1)
		signal.Notify(shutdownSignals, syscall.SIGTERM, syscall.SIGINT)
		go func() {
			<-shutdownSignals
			d.Close()
		}()

		d.Run(context.Background())

		fmt.Println("OK, bye!")
		return nil
	},
}

func init() {
	rootCmd.Flags().String("config", defaultConfigPath, "path to config file")
}
