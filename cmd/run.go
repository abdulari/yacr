/*
Copyright © 2026 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"syscall"

	"os/user"

	"github.com/google/uuid"
	"github.com/spf13/cobra"
)

// runCmd represents the run command
var runCmd = &cobra.Command{
	Use:   "run",
	Short: "run the container",
	Run: func(cmd *cobra.Command, args []string) {

		currentUser, err := user.Current()
		if err != nil {
			panic(err)
		}
		if currentUser.Uid != "0" {
			fmt.Println("you need sudo to run this command")
			os.Exit(1)
		}

		target := "linux"

		if len(args) == 0 {
			fmt.Println("assume command is /bin/bash")
			args = append([]string{"/bin/bash"}, args...)
		}

		// set id for container
		id := uuid.New().String()
		id = "c4bfa5a4-4b13-4724-ab6c-26a9f0c81dc9"

		// create folder for container
		must(os.MkdirAll("/tmp/yacr/runs/"+id+"/merged", 0755))
		must(os.MkdirAll("/tmp/yacr/runs/"+id+"/upper", 0755))
		must(os.MkdirAll("/tmp/yacr/runs/"+id+"/blob", 0755))
		must(os.MkdirAll("/tmp/yacr/runs/"+id+"/work", 0755))

		// TODO: download container
		containerFolder := "/tmp/yacr/containers/alpine__latest"
		digest := strings.Replace("sha256:589002ba0eaed121a1dbf42f6648f29e5be55d5c8a6ee0f8eaa0285cc21ac153", "sha256:", "", 1)

		folderToExtract := "/tmp/yacr/runs/" + id + "/blob" + "/" + digest

		// extract container to folder
		must(os.MkdirAll(folderToExtract, 0755))
		extractContainer := exec.Command("tar", "-zxvf", containerFolder+"/"+digest, "-C", folderToExtract)
		must(extractContainer.Run())

		// mount container
		cmdMountContainer := "mount -t overlay overlay -o lowerdir=" + folderToExtract + ",upperdir=/tmp/yacr/runs/" + id + "/upper,workdir=/tmp/yacr/runs/" + id + "/work /tmp/yacr/runs/" + id + "/merged"
		must(exec.Command("sh", "-c", cmdMountContainer).Run())

		// cmdlineToRun := append([]string{"helperToRunTheCommandInContainer"}, args...)
		cmdlineToRun := append([]string{"helperToRunTheCommandInContainer"}, id)
		cmdlineToRun = append(cmdlineToRun, args...)
		fmt.Printf("cmdlineToRun: %v\n", cmdlineToRun)
		terminal := exec.Command("/proc/self/exe", cmdlineToRun...)
		terminal.Stdin = cmd.InOrStdin()
		terminal.Stdout = cmd.OutOrStdout()
		terminal.Stderr = cmd.ErrOrStderr()

		if target == "linux" {
			terminal.SysProcAttr = &syscall.SysProcAttr{
				Cloneflags: syscall.CLONE_NEWUTS | syscall.CLONE_NEWPID | syscall.CLONE_NEWNS,
			}
		}

		terminal.Run()
	},
}

func must(err error) {
	if err != nil {
		panic(err)
	}
}

func init() {
	rootCmd.AddCommand(runCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// runCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// runCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
