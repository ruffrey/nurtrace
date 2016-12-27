package main

import (
	"bleh/potential"
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/pkg/sftp"

	"golang.org/x/crypto/ssh"
)

func publicKeyFile(file string) ssh.AuthMethod {
	buffer, err := ioutil.ReadFile(file)
	if err != nil {
		return nil
	}

	key, err := ssh.ParsePrivateKey(buffer)
	if err != nil {
		return nil
	}
	return ssh.PublicKeys(key)
}

const workerLocation = "/opt/pworker"

// Worker is
type Worker struct {
	conn *ssh.Client
	host string
}

// NewWorker is a worker constructor
func NewWorker(hostAndPort string) (w *Worker, err error) {
	// u, err := user.Current()
	// if err != nil {
	// 	return w, err
	// }
	w = &Worker{host: hostAndPort}

	// Dial your ssh server.
	w.conn, err = ssh.Dial("tcp", w.host, &ssh.ClientConfig{
		User: "root",
		Auth: []ssh.AuthMethod{
			publicKeyFile(os.Getenv("HOME") + "/.ssh/id_rsa"),
		},
	})
	if err != nil {
		return w, err
	}

	return w, nil
}

func (w *Worker) transerExecutable() (err error) {
	var remoteOS string
	var remoteArch string

	session, err := w.conn.NewSession()
	if err != nil {
		return err
	}
	defer session.Close()

	bytes, err := session.CombinedOutput("uname -s;uname -m;")
	if err != nil {
		return err
	}
	s := strings.Split(strings.ToLower(string(bytes)), "\n")

	remoteOS = s[0]
	remoteArch = s[1]
	fmt.Println(remoteOS, remoteArch)

	err = w.scpSendFile("worker_" + remoteOS + "_" + remoteArch)
	return err
}

func (w *Worker) scpSendFile(localFile string) (err error) {
	session, err := w.conn.NewSession()
	if err != nil {
		return err
	}
	defer session.Close()

	bytes, err := ioutil.ReadFile(localFile)
	if err != nil {
		return err
	}
	scp, err := sftp.NewClient(w.conn)
	if err != nil {
		return err
	}
	remoteFile, err := scp.Create(workerLocation)
	if err != nil {
		return err
	}
	_, err = remoteFile.Write(bytes)
	return err
}

// Train is
func (w *Worker) Train(network *potential.Network, samples []potential.TrainingSample) (finalNetwork *potential.Network, err error) {
	session, err := w.conn.NewSession()
	if err != nil {
		return network, err
	}
	defer session.Close()

	return network, err
}

func main() {
	hostPort := os.Args[1]
	setup := len(os.Args) > 2 && os.Args[2] == "setup"
	fmt.Println("New worker at", hostPort)
	w, err := NewWorker(hostPort)
	if err != nil {
		panic(err)
	}
	defer w.conn.Close()
	fmt.Println("Connected to", hostPort)

	if setup {
		fmt.Println("Setting up")
		err = w.transerExecutable()
		if err != nil {
			panic(err)
		}
	}

	fmt.Println("Training")

	session, err := w.conn.NewSession()
	if err != nil {
		panic(err)
	}
	defer session.Close()
	bytes, err := session.CombinedOutput(workerLocation)
	fmt.Println(string(bytes))
	if err != nil {
		fmt.Print(hostPort, " ")
		panic(err)
	}

	if err != nil {
		panic(err)
	}
}
