package main

import (
	"archive/zip"
	"context"
	"embed"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math/rand"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/bitfield/script"
	"github.com/bramvdbogaerde/go-scp"
	pcmd "github.com/elulcao/progress-bar/cmd"
	"github.com/sirupsen/logrus"
	"golang.org/x/crypto/ssh"
)

type envConfig struct {
	Host              string `json:"Host"`
	Port              int    `json:"Port"`
	Username          string `json:"Username"`
	Password          string `json:"Password"`
	Geekbench_License string `json:"Geekbench_License"`
}

// Start an iperf3 server locally.
func iperfServer() {
	cmd := exec.Command("benchmark/tools/iperf3", "-s")
	cmd.Run()
}

func pBar(hsotNum int) {
	pb := pcmd.NewPBar()
	pb.SignalHandler()
	pb.Total = uint16(150 * hsotNum)

	// pb.RenderPBar(i)
	for i := 1; uint16(i) <= pb.Total; i++ {
		pb.RenderPBar(i)
		if uint16(i) == pb.Total-1 {
			i -= 2
		}
		time.Sleep(1 * time.Second)
	}
}

func Unzip(src string, dest string) error {
	r, err := zip.OpenReader(src)
	if err != nil {
		return err
	}
	defer r.Close()

	for _, f := range r.File {
		rc, err := f.Open()
		if err != nil {
			return err
		}
		defer rc.Close()

		fpath := filepath.Join(dest, f.Name)
		if f.FileInfo().IsDir() {
			os.MkdirAll(fpath, os.ModePerm)
		} else {
			var dir string
			if lastIndex := strings.LastIndex(fpath, string(os.PathSeparator)); lastIndex > -1 {
				dir = fpath[:lastIndex]
			}
			err = os.MkdirAll(dir, os.ModePerm)
			if err != nil {
				log.Fatal(err)
				return err
			}
			f, err := os.OpenFile(
				fpath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
			if err != nil {
				return err
			}
			defer f.Close()

			_, err = io.Copy(f, rc)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func randomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyz" +
		"ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

	var seededRand *rand.Rand = rand.New(
		rand.NewSource(time.Now().UnixNano()))

	b := make([]byte, length)
	for i := range b {
		b[i] = charset[seededRand.Intn(len(charset))]
	}
	return string(b)
}

// Encapsulating Remote Execution of Shell Commands
func remoteShellExec(rClient ssh.Client, commands []string, logOut bool) error {
	outputStr := ""
	for _, command := range commands {
		session, err := rClient.NewSession()
		if err != nil {
			logrus.Errorf("Failed to create session: %s", err.Error())
		}
		defer session.Close()
		output, err := session.CombinedOutput(command)
		if err != nil {
			logrus.Errorf("Failed to run: %s", err.Error())
		}
		outputStr = string(output)
		if logOut {
			fmt.Println(" ")
			fmt.Println(outputStr)
		}
	}
	return nil
}

// Test whether the SSH target host is reachable.
func testSSH(rClient ssh.Client) bool {

	session, err := rClient.NewSession()
	if err != nil {
		logrus.Error("Failed to dial: ", err)
		return false
	}
	defer session.Close()
	return true
}

//go:embed benchmark.zip
var f embed.FS

func main() {
	srcDir := "benchmark"
	dstDir := "/tmp/" + randomString(10) + "/"

	data, _ := f.ReadFile("benchmark.zip")

	// Write byte slices to temporary zip files.
	tmpfile, err := os.CreateTemp("", "example")
	if err != nil {
		panic(err)
	}
	tmpfile.Write(data)
	tmpfile.Close()

	// Extract files from embed.
	Unzip(tmpfile.Name(), ".")

	file, err := os.Open("config.json")
	if err != nil {
		logrus.Error("Error opening file:", err)
		return
	}
	defer file.Close()
	envConfig := &envConfig{}
	decoder := json.NewDecoder(file)
	err = decoder.Decode(&envConfig)
	if err != nil {
		fmt.Println("Error decoding file:", err)
		return
	}
	// Read the local config.json file and inject it as variables.
	username := envConfig.Username
	password := envConfig.Password
	port := envConfig.Port
	geekbenchLicense := envConfig.Geekbench_License
	HostExecOut, _ := script.Exec("hostname -I").String()
	hostIPS := strings.Split(HostExecOut, " ")

	config := &ssh.ClientConfig{
		User: username,
		Auth: []ssh.AuthMethod{
			ssh.Password(password),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}

	// Print and start the motto.
	logrus.Info("Benchmarking...")

	// Get current time.
	start := time.Now()

	// 初始化 bar cmd
	go pBar(len(strings.Split(envConfig.Host, ",")))

	// 创建 iperfServer 协程
	go iperfServer()

	// Automatically close the iperfServer coroutine.
	defer func() {
		iperf3KillCmd := exec.Command("/bin/bash", "-c", "ps aux | awk '/iperf3/ {print $2}' | xargs kill -9")
		iperf3KillCmd.Run()
		cleanWorkspaceCmd := exec.Command("rm", "-rf", srcDir)
		cleanWorkspaceCmd.Run()
	}()

	// Use the testSSH function to test whether the target host is reachable.
	for _, host := range strings.Split(envConfig.Host, ",") {
		serverURL := fmt.Sprintf("%s:%d", host, port)
		client, err := ssh.Dial("tcp", serverURL, config)
		if err != nil {
			logrus.Error("Failed to dial: ", err)
		}
		if testSSH(*client) {
			logrus.Infof("Host: %s, SSH connection successful.", host)
		} else {
			logrus.Errorf("Host: %s, SSH connection failed.", host)
		}
	}

	// Split each host in config.Host by comma.
	for _, host := range strings.Split(envConfig.Host, ",") {
		serverURL := fmt.Sprintf("%s:%d", host, port)

		client, err := ssh.Dial("tcp", serverURL, config)
		if err != nil {
			logrus.Error("Failed to dial: ", err)
		}

		_ = remoteShellExec(*client, []string{"mkdir -p " + dstDir}, false)

		filepath.Walk(srcDir, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if info.IsDir() {
				remoteShellExec(*client, []string{"mkdir -p " + dstDir + path}, false)
			} else {
				scpClient, err := scp.NewClientBySSH(client)
				scpClient.Connect()
				file, err := os.Open(path)
				defer file.Close()
				if err != nil {
					logrus.Error("Failed to open file: ", err)
				}
				err = scpClient.CopyFile(context.Background(), file, dstDir+path, "0655")
				if err != nil {
					logrus.Error("Failed to copy file: ", err)
				}
				scpClient.Close()

			}
			return nil
		})

		// Execute remote script.
		logrus.Infof("Host: %s, Executing the remote script...", host)

		_ = remoteShellExec(*client, []string{
			"cd " + dstDir + srcDir + " && " + "./run.sh " + hostIPS[0] + " " + "'" + geekbenchLicense + "'",
		}, true)

		// Delete remote file.
		remoteShellExec(*client, []string{"rm -rf " + dstDir}, false)
	}

	// Print execution time.
	logrus.Infof("Total execution time for this run: %s", time.Since(start))
}
