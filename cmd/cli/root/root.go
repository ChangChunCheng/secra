package root

import (
	"fmt"
	"runtime"

	"github.com/spf13/cobra"
	"gitlab.com/jacky850509/secra/cmd/cli/backup"
	"gitlab.com/jacky850509/secra/cmd/cli/health"
	importcmd "gitlab.com/jacky850509/secra/cmd/cli/import"
	"gitlab.com/jacky850509/secra/cmd/cli/migrate"
	"gitlab.com/jacky850509/secra/cmd/cli/resource"
	"gitlab.com/jacky850509/secra/cmd/cli/user"
	"gitlab.com/jacky850509/secra/internal/config"
)

var rawVersion bool

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the detailed version and build information of Secra",
	Run: func(cmd *cobra.Command, args []string) {
		if rawVersion {
			fmt.Print(config.Version)
			return
		}
		fmt.Printf("Secra Vulnerability Platform\n")
		fmt.Printf("----------------------------\n")
		fmt.Printf("Version:    %s\n", config.Version)
		fmt.Printf("Build Date: %s\n", config.BuildDate)
		fmt.Printf("Git Commit: %s\n", config.Commit)
		fmt.Printf("Built By:   %s\n", config.BuiltBy)
		fmt.Printf("OS/Arch:    %s/%s (Target: %s/%s)\n", runtime.GOOS, runtime.GOARCH, config.OS, config.Arch)
		fmt.Printf("Go Version: %s\n", runtime.Version())
	},
}

// rootCmd is the base command for secra CLI
var rootCmd = &cobra.Command{
	Use:   "secra",
	Short: "Secra CLI tool",
}

// Execute runs the root command
func Execute() error {
	return rootCmd.Execute()
}

func init() {
	versionCmd.Flags().BoolVar(&rawVersion, "raw", false, "Print only the version string")
	rootCmd.AddCommand(versionCmd)
	rootCmd.AddCommand(user.Cmd)
	rootCmd.AddCommand(resource.Cmd)
	rootCmd.AddCommand(migrate.Cmd)
	rootCmd.AddCommand(importcmd.Cmd)
	rootCmd.AddCommand(backup.BackupCmd)
	rootCmd.AddCommand(health.Cmd)
}
