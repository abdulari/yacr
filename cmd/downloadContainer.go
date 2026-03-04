/*
Copyright © 2026 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/spf13/cobra"
)

var pathToContainerFolder string = "/tmp/yacr/containers/"
var force bool = false

// downloadContainerCmd represents the downloadContainer command
var downloadContainerCmd = &cobra.Command{
	Use:    "downloadContainer",
	Hidden: true,

	Run: func(cmd *cobra.Command, args []string) {
		// fmt.Println("downloadContainer called")
		// check if skopeo is installed

		result := exec.Command("skopeo", "--version").Run()
		if result != nil {
			fmt.Println("skopeo is not installed, please install it to use this command")
			os.Exit(1)
		}
		must(os.MkdirAll(pathToContainerFolder, 0755))
		imageURL := args[0]
		imageName := strings.Split(imageURL, "/")
		folderName := strings.Replace(imageName[len(imageName)-1], ":", "__", 1)
		// check if the folder already exists
		if _, err := os.Stat(pathToContainerFolder + folderName); !os.IsNotExist(err) && !force {
			fmt.Println("container already exists, skipping download")
			return
		} else {
			fmt.Println("download container image using skopeo and save it to " + pathToContainerFolder + folderName)
			result2 := exec.Command("skopeo", "copy", "docker://"+imageURL, "dir:"+pathToContainerFolder+"/"+folderName)
			result2.Stdin = cmd.InOrStdin()
			result2.Stdout = cmd.OutOrStdout()
			result2.Stderr = cmd.ErrOrStderr()
			result2.Run()
			// fmt.Println(result2)
		}

	},
}

func init() {
	rootCmd.AddCommand(downloadContainerCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// downloadContainerCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// downloadContainerCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
