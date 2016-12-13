package cmd

import (
	"fmt"
	"github.com/cloudposse/github-authorized-keys/config"
	"github.com/cloudposse/github-authorized-keys/jobs"
	"github.com/cloudposse/github-authorized-keys/server"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"os"
	"time"
)

var cfgFile string

// ETCDTTLDefault - default ttl - 1day in seconds = 24 hours * 60 minutes * 60 seconds
const ETCDTTLDefault = int64(24 * 60 * 60)

// SyncUsersIntervalDefault - default interval between synchronize users - 5 minutes in seconds = 5 minutes * 60 seconds
const SyncUsersIntervalDefault = int64(5 * 60)

var flags = []flag{
	{"a", "string", "github_api_token", "", "Github API token    ( environment variable GITHUB_API_TOKEN could be used instead ) (read more https://github.com/blog/1509-personal-api-tokens)"},
	{"o", "string", "github_organization", "", "Github organization ( environment variable GITHUB_ORGANIZATION could be used instead )"},
	{"n", "string", "github_team", "", "Github team name    ( environment variable GITHUB_TEAM could be used instead )"},
	{"i", "int", "github_team_id", 0, "Github team id 	    ( environment variable GITHUB_TEAM_ID could be used instead )"},

	{"g", "string", "sync_users_gid", "", "Primary group id    ( environment variable SYNC_USERS_GID could be used instead )"},
	{"G", "strings", "sync_users_groups", []string{}, "CSV groups name     ( environment variable SYNC_USERS_GROUPS could be used instead )"},
	{"s", "string", "sync_users_shell", "/bin/bash", "User shell 	    ( environment variable SYNC_USERS_SHELL could be used instead )"},
	{"r", "string", "sync_users_root", "/", "Root directory 	    ( environment variable SYNC_USERS_ROOT could be used instead )"},
	{"c", "int64", "sync_users_interval", SyncUsersIntervalDefault, "Sync each x sec     ( environment variable SYNC_USERS_INTERVAL could be used instead )"},

	{"e", "strings", "etcd_endpoint", []string{}, "CSV etcd endpoints  ( environment variable ETCD_ENDPOINT could be used instead )"},
	{"p", "string", "etcd_prefix", "/github-authorized-keys", "Path for etcd data  ( environment variable ETCD_PREFIX could be used instead )"},
	{"t", "int64", "etcd_ttl", ETCDTTLDefault, "ETCD value's ttl    ( environment variable ETCD_TTL could be used instead )"},

	{"d", "bool", "integrate_ssh", false, "Integrate with ssh  ( environment variable INTEGRATE_SSH could be used instead )"},
	{"l", "string", "listen", ":301", "Listen              ( environment variable LISTEN could be used instead )"},
}

// RootCmd represents the base command when called without any subcommands
var RootCmd = &cobra.Command{
	Use:   "github-authorized-keys",
	Short: "Use GitHub teams to manage system user accounts and authorized_keys",
	Long: `
Use GitHub teams to manage system user accounts and authorized_keys.

Config:
  REQUIRED: Github API token        | flag --github-api-token    OR environment variable GITHUB_API_TOKEN
  REQUIRED: Github organization     | flag --github-organization OR environment variable GITHUB_ORGANIZATION
  REQUIRED: One of
  		   Github team name | flag --github-team    OR environment variable GITHUB_TEAM
  			OR
  		   Github team id   | flag --github-team-id OR Environment variable GITHUB_TEAM_ID
`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// @TODO Support viper duration type
		etcdTTL, err := time.ParseDuration(viper.GetString("etcd_ttl") + "s")

		if err != nil {
			return err
		}

		cfg := config.Config{
			GithubAPIToken:     viper.GetString("github_api_token"),
			GithubOrganization: viper.GetString("github_organization"),
			GithubTeamName:     viper.GetString("github_team"),
			GithubTeamID:       viper.GetInt("github_team_id"),

			EtcdEndpoints: fixStringSlice(viper.GetString("etcd_endpoint")),
			EtcdTTL:       etcdTTL,

			UserGID:    viper.GetString("sync_users_gid"),
			UserGroups: fixStringSlice(viper.GetString("sync_users_groups")),
			UserShell:  viper.GetString("sync_users_shell"),
			Root:       viper.GetString("sync_users_root"),
			Interval:   uint64(viper.GetInt64("sync_users_interval")),

			IntegrateWithSSH: viper.GetBool("integrate_ssh"),

			Listen: viper.GetString("listen"),
		}

		err = cfg.Validate()

		if err == nil {
			jobs.Run(cfg)
			server.Run(cfg)
		}

		return err
	},
}

// Execute adds all child commands to the root command sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := RootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(-1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	// Config file
	RootCmd.PersistentFlags().StringVar(&cfgFile, "config", "",
		"Config file         (default is $HOME/.github-authorized-keys.yaml)")

	for _, f := range flags {
		createCmdFlags(RootCmd, f)
	}
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" { // enable ability to specify config file via flag
		viper.SetConfigFile(cfgFile)
	}

	viper.SetConfigName(".github-authorized-keys") // name of config file (without extension)
	viper.AddConfigPath("$HOME")                   // adding home directory as first search path
	viper.AutomaticEnv()                           // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		fmt.Println("Using config file:", viper.ConfigFileUsed())
	}
}