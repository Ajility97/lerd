package cli

import (
	"os"

	"github.com/spf13/cobra"
)

// NewSailCmd returns a top-level `lerd sail` command that passes everything to
// vendor/bin/sail on the host, so existing Sail workflows (lerd sail up, lerd
// sail artisan migrate, …) keep working. Registering it keeps the vendor-bin
// dispatcher from running Sail inside lerd's PHP container.
func NewSailCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "sail",
		Short: "Run vendor/bin/sail on the host",
		// DisableFlagParsing lets flags like `lerd sail up -d` pass through
		// to vendor/bin/sail unchanged.
		DisableFlagParsing: true,
		SilenceUsage:       true,
		RunE: func(cmd *cobra.Command, args []string) error {
			// DisableFlagParsing suppresses cobra's --help interception, so
			// handle it here before falling through to vendor/bin/sail.
			for _, a := range args {
				if a == "--help" || a == "-h" {
					return cmd.Help()
				}
			}
			cwd, err := os.Getwd()
			if err != nil {
				return err
			}
			// Sail drives Docker Compose, so it runs on the host. lerd's PHP
			// container has neither docker nor the bash sail's script asks for.
			return RunHostVendorBin(cwd, "vendor/bin/sail", args)
		},
	}
	return cmd
}
