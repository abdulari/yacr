/*
Copyright © 2026 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"os"
	"os/exec"
	"syscall"

	"github.com/spf13/cobra"
)

// helperToRunTheCommandInContainerCmd represents the helperToRunTheCommandInContainer command
var helperToRunTheCommandInContainerCmd = &cobra.Command{
	Use:    "helperToRunTheCommandInContainer",
	Hidden: true,
	Run: func(cmd *cobra.Command, args []string) {
		// fmt.Printf("running %v as PID %d \n", args[1], os.Getpid())
		id := args[0]
		// fmt.Println("container id = " + id)
		terminal := exec.Command(args[1], args[2:]...)
		terminal.Stdin = cmd.InOrStdin()
		terminal.Stdout = cmd.OutOrStdout()
		terminal.Stderr = cmd.ErrOrStderr()

		// chroot to merged folder
		must(syscall.Chroot("/tmp/yacr/runs/" + id + "/merged"))
		must(os.Chdir("/"))
		must(syscall.Mount("proc", "proc", "proc", 0, ""))

		terminal.Run()

		// we changed the root. so  must be relative to the latest root.
		must(syscall.Unmount("proc", 0))
		must(syscall.Chroot("/"))
		must(os.Chdir("/"))
		// must(syscall.Unmount("/tmp/yacr/runs/"+id+"/merged", 0))
	},
}

func init() {
	rootCmd.AddCommand(helperToRunTheCommandInContainerCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// helperToRunTheCommandInContainerCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// helperToRunTheCommandInContainerCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
