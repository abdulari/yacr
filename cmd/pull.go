/*
Copyright © 2026 abdulari <abdul.arif.b.abdul.muttalib@intel.com>
*/
package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/spf13/cobra"
)

var pathToContainerFolder string = "/tmp/yacr/containers/"
var force bool = false
var skipDownload bool = false

type ContainerInfo struct {
	Name       string `json:"name"`
	Tag        string `json:"tag"`
	Url        string `json:"url"`
	Layers     string `json:"layers"`
	ConfigFile string `json:"configFile"`
	ImageName  string `json:"imageName"`
	FolderName string `json:"folderName"`
}

// downloadContainerCmd represents the downloadContainer command
var downloadContainerCmd = &cobra.Command{
	Use:   "pull <container url>",
	Short: "download the container image from registry and save it to local folder",

	Run: func(cmd *cobra.Command, args []string) {
		// fmt.Println("downloadContainer called")
		// check if skopeo is installed
		force, _ = cmd.Flags().GetBool("force")
		skipDownload, _ = cmd.Flags().GetBool("skipDownload")

		info := ContainerInfo{
			Name:       "",
			Tag:        "",
			Url:        "",
			ImageName:  "",
			FolderName: "",
			Layers:     "",
			ConfigFile: "",
		}

		result := exec.Command("skopeo", "--version").Run()
		if result != nil {
			fmt.Println("skopeo is not installed, please install it to use this command")
			os.Exit(1)
		}
		if len(args) == 0 {
			fmt.Println("ERROR - no container url provided\n")
			fmt.Println("please provide the image url to download.")
			fmt.Println("example: yacr pull docker.io/library/alpine:latest")

			os.Exit(1)
		}
		must(os.MkdirAll(pathToContainerFolder, 0755))

		// handle image url default to docker.io if no registry is provided
		if !strings.Contains(args[0], "/") {
			args[0] = "docker.io/library/" + args[0]
		} else if strings.Count(args[0], "/") == 1 {
			args[0] = "docker.io/" + args[0]
		}
		// ^[a-zA-Z0-9]+[\:]?[a-zA-Z0-9]*$
		// regexp.MustCompile(``).MatchString(args[0])

		info.Url = args[0]

		imageName := strings.Split(info.Url, "/")
		info.ImageName = imageName[len(imageName)-1]
		info.Name = strings.Split(info.ImageName, ":")[0]
		info.Tag = strings.Split(info.ImageName, ":")[1]
		folderName := strings.Replace(info.ImageName, ":", "__", 1)

		info.FolderName = pathToContainerFolder + folderName
		// check if the folder already exists
		if _, err := os.Stat(info.FolderName + "/container.json"); !os.IsNotExist(err) && !force {
			fmt.Println("container already exists, skipping download")
			fmt.Println("add --force flag to force download the container image again")
		} else {
			fmt.Println("download container image using skopeo and save it to " + info.FolderName)
			if skipDownload == false {
				result2 := exec.Command("skopeo", "copy", "docker://"+info.Url, "dir:"+info.FolderName)
				result2.Stdin = cmd.InOrStdin()
				result2.Stdout = cmd.OutOrStdout()
				result2.Stderr = cmd.ErrOrStderr()
				result2.Run()
			}

			//unpack configuration file and layer files
			containerLayers, _ := exec.Command("sh", "-c", "cd "+pathToContainerFolder+folderName+" && cat manifest.json | jq -r -j .layers[].digest").Output()
			info.Layers = strings.Replace(string(containerLayers), "sha256:", "", 1)
			containerConfig, _ := exec.Command("sh", "-c", "cd "+pathToContainerFolder+folderName+" && cat manifest.json | jq -r -j .config.digest").Output()
			info.ConfigFile = strings.Replace(string(containerConfig), "sha256:", "", 1)

			jsonData, err := json.Marshal(info)
			if err != nil {
				panic(err)
			}
			fmt.Println(string(jsonData))
			must(os.WriteFile(info.FolderName+"/container.json", jsonData, 0644))

		}
		os.Exit(0)
	},
}

func init() {
	downloadContainerCmd.Flags().BoolP("force", "f", false, "force download the container image even if it already exists")
	downloadContainerCmd.Flags().BoolP("skipDownload", "s", false, "skip downloading the container image")
	rootCmd.AddCommand(downloadContainerCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// downloadContainerCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// downloadContainerCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
