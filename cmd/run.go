/*
Copyright © 2026 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"strings"
	"syscall"

	"os/user"

	"github.com/google/uuid"
	"github.com/spf13/cobra"
)

type RuntimeInfo struct {
	Id    string
	Image string
}

// runCmd represents the run command
var runCmd = &cobra.Command{
	Use:   "run",
	Short: "run the container",
	Run: func(cmd *cobra.Command, args []string) {
		logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
		info := RuntimeInfo{
			Id:    "",
			Image: "",
		}

		currentUser, err := user.Current()
		if err != nil {
			panic(err)
		}
		if currentUser.Uid != "0" {
			logger.Info("you need sudo to run this command")
			os.Exit(1)
		}

		target := "linux"

		if len(args) == 0 {
			logger.Info("Error - no container specified")
			logger.Info("yacr run <image> [command]")
			os.Exit(1)
		}

		// extract image name from args
		if !strings.Contains(args[0], "/") {
			args[0] = "docker.io/library/" + args[0]
		} else if strings.Count(args[0], "/") == 1 {
			args[0] = "docker.io/" + args[0]
		}
		image := args[0]
		info.Image = image
		imageName := strings.Split(info.Image, "/")
		folderName := strings.Replace(imageName[len(imageName)-1], ":", "__", 1)

		// check if image exists in local folder

		imagePull := exec.Command("/proc/self/exe", "pull", image)
		imagePull.Stdin = cmd.InOrStdin()
		imagePull.Stdout = cmd.OutOrStdout()
		imagePull.Stderr = cmd.ErrOrStderr()
		err = imagePull.Run()
		if err != nil {
			panic(err)
		}

		logger.Debug(pathToContainerFolder + folderName)
		containerLayers, _ := exec.Command("sh", "-c", "cd "+pathToContainerFolder+folderName+" && cat manifest.json | jq -r -j .layers[].digest").Output()
		layers := strings.Replace(string(containerLayers), "sha256:", "", 1)
		containerConfig, _ := exec.Command("sh", "-c", "cd "+pathToContainerFolder+folderName+" && cat manifest.json | jq -r -j .config.digest").Output()
		configFile := strings.Replace(string(containerConfig), "sha256:", "", 1)

		logger.Debug("layers: " + layers)
		logger.Debug("Config: " + configFile)

		configCmd, _ := exec.Command("sh", "-c", "cd "+pathToContainerFolder+folderName+" && cat "+configFile+" | jq -r -j .config.Cmd[]").Output()
		logger.Debug(string(configCmd))
		configEnv, _ := exec.Command("sh", "-c", "cd "+pathToContainerFolder+folderName+" && cat "+configFile+" | jq -r -j .config.Env[]").Output()
		logger.Debug(string(configEnv))

		// extract container info from config

		// set id for container
		id := uuid.New().String()
		id = "c4bfa5a4-4b13-4724-ab6c-26a9f0c81dc9"
		info.Id = id
		// logger.Debug(info)

		// create folder for container
		must(os.MkdirAll("/tmp/yacr/runs/"+id+"/merged", 0755))
		must(os.MkdirAll("/tmp/yacr/runs/"+id+"/upper", 0755))
		must(os.MkdirAll("/tmp/yacr/runs/"+id+"/blob", 0755))
		must(os.MkdirAll("/tmp/yacr/runs/"+id+"/work", 0755))

		// TODO: download container
		// containerFolder := "/tmp/yacr/containers/alpine__latest"

		// digest := strings.Replace("sha256:589002ba0eaed121a1dbf42f6648f29e5be55d5c8a6ee0f8eaa0285cc21ac153", "sha256:", "", 1)

		folderToExtract := "/tmp/yacr/runs/" + id + "/blob" + "/" + layers

		// extract container to folder
		must(os.MkdirAll(folderToExtract, 0755))
		extractContainer := exec.Command("tar", "-zxvf", pathToContainerFolder+folderName+"/"+layers, "-C", folderToExtract)
		must(extractContainer.Run())

		// mount container
		// cmdMountContainer := "mount -t overlay overlay -o lowerdir=" + folderToExtract + ",upperdir=/tmp/yacr/runs/" + id + "/upper,workdir=/tmp/yacr/runs/" + id + "/work /tmp/yacr/runs/" + id + "/merged"
		// fmt.Println(cmdMountContainer)
		must(syscall.Mount("overlay", "/tmp/yacr/runs/"+id+"/merged", "overlay", 0, "lowerdir="+folderToExtract+",upperdir=/tmp/yacr/runs/"+id+"/upper,workdir=/tmp/yacr/runs/"+id+"/work"))
		// must(exec.Command("sh", "-c", cmdMountContainer).Run())

		// write container info to file
		infoFile := "/tmp/yacr/runs/" + id + "/info.json"
		infoData := fmt.Sprintf(`{"id":"%s","image":"%s"}`, info.Id, info.Image)
		must(os.WriteFile(infoFile, []byte(infoData), 0644))

		cmdlineToRun := append([]string{"helperToRunTheCommandInContainer"}, id)
		if len(args) > 1 {
			cmdlineToRun = append(cmdlineToRun, args[1:]...)
		} else {
			cmdlineToRun = append(cmdlineToRun, string(configCmd))
		}

		terminal := exec.Command("/proc/self/exe", cmdlineToRun...)
		terminal.Env = strings.Split(string(configEnv), "\n")
		terminal.Stdin = cmd.InOrStdin()
		terminal.Stdout = cmd.OutOrStdout()
		terminal.Stderr = cmd.ErrOrStderr()

		if target == "linux" {
			terminal.SysProcAttr = &syscall.SysProcAttr{
				Cloneflags: syscall.CLONE_NEWUTS | syscall.CLONE_NEWPID | syscall.CLONE_NEWNS,
			}
		}

		terminal.Run()

		// cleaning up
		cmdUnmountContainer := "/tmp/yacr/runs/" + id + "/merged"
		must(exec.Command("umount", cmdUnmountContainer).Run())

		// need to use "rm" command to remove the container after run.
		//must(os.RemoveAll("/tmp/yacr/runs/" + id))
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
