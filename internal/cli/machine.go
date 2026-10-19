package cli

import (
	"fmt"
	"strconv"

	"github.com/geodro/lerd/internal/feedback"
	"github.com/spf13/cobra"
)

// NewMachineCmd returns the machine parent command for Podman Machine
// maintenance operations.
func NewMachineCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "machine",
		Short: "Manage the Podman Machine VM (macOS)",
	}
	cmd.AddCommand(newMachineResetCmd())
	cmd.AddCommand(newMachineReclaimCmd())
	cmd.AddCommand(newMachineMemoryCmd())
	return cmd
}

func newMachineReclaimCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "reclaim",
		Short: "Return disk the Podman Machine VM no longer uses to the host (macOS)",
		Long: `Return disk the Podman Machine VM is holding but no longer using.

The VM's disk image is a sparse file that only ever grows: blocks freed inside
the VM stay charged to the host until the guest filesystem is trimmed. This
caps the VM's systemd journal, vacuums it, then trims the filesystem, and
reports how much host disk came back. Containers, images and site data are
untouched.`,
		RunE: func(_ *cobra.Command, _ []string) error {
			return runMachineReclaim()
		},
	}
}

func newMachineResetCmd() *cobra.Command {
	var assumeYes bool
	cmd := &cobra.Command{
		Use:   "reset",
		Short: "Recreate the Podman Machine VM (fixes corrupted container storage; data is preserved)",
		Long: `Recreate the Podman Machine VM.

Use this when lerd start fails with a container-storage error such as
"getting graph driver info ... overlay: invalid argument" after an unclean
host shutdown. lerd's databases and site data are stored on the host and are
preserved; container images are rebuilt automatically on the next start.`,
		RunE: func(_ *cobra.Command, _ []string) error {
			return runMachineReset(assumeYes)
		},
	}
	cmd.Flags().BoolVarP(&assumeYes, "yes", "y", false, "Skip the confirmation prompt")
	return cmd
}

func newMachineMemoryCmd() *cobra.Command {
	var assumeYes bool
	cmd := &cobra.Command{
		Use:   "memory [GB]",
		Short: "Show or set the Podman Machine VM's memory (macOS)",
		Long: `Show whether the Podman Machine VM has run out of memory since it booted,
or restart it with GB of memory. The restart stops every container and brings
them back with lerd start; site data is untouched.`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(_ *cobra.Command, args []string) error {
			if len(args) == 0 {
				printMachineMemory(MachineMemoryStatus())
				return nil
			}
			gb, err := strconv.Atoi(args[0])
			if err != nil {
				return fmt.Errorf("%q is not a whole number of GB", args[0])
			}
			if !assumeYes && !confirmMachineRestart(gb) {
				fmt.Println("Cancelled.")
				return nil
			}
			return ResizeMachineMemory(int64(gb)*1024, nil)
		},
	}
	cmd.Flags().BoolVarP(&assumeYes, "yes", "y", false, "Skip the confirmation prompt")
	return cmd
}

func printMachineMemory(st MachineMemory) {
	if len(st.Kills) == 0 {
		fmt.Println("The Podman Machine has not run out of memory since it booted.")
		return
	}
	fmt.Println("The Podman Machine " + st.Summary() + ".")
	fmt.Println("Fix: " + st.Hint())
}

func confirmMachineRestart(gb int) bool {
	return feedback.Confirm(fmt.Sprintf("Restart the Podman Machine with %d GB? Every container stops until lerd start brings them back.", gb), false)
}
