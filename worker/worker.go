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
	u := os.Getenv("USER")
	w = &Worker{host: hostAndPort}

	// Dial your ssh server.
	w.conn, err = ssh.Dial("tcp", w.host, &ssh.ClientConfig{
		User: u,
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
	var hostPort string
	var actionArg string
	var setup bool
	var train bool
	var run bool

	if len(os.Args) > 1 {
		if len(os.Args) > 2 {
			hostPort = os.Args[1]
			actionArg = os.Args[2]
		} else {
			actionArg = os.Args[1]
		}
		switch actionArg {
		case "setup": // setup remote worker
			setup = true
			break
		case "train": // train remote worker
			train = true
			break
		case "run": // while on remote worker, run the training
			run = true
		default:
			fmt.Println("Last argument should be either 'setup' or 'train'")
			return
		}
	}

	if setup {
		fmt.Println("New worker at", hostPort)
		w, err := NewWorker(hostPort)
		if err != nil {
			fmt.Println(err)
			return
		}
		defer w.conn.Close()
		fmt.Println("Connected to", hostPort)

		fmt.Println("Setting up", hostPort)
		err = w.transerExecutable()
		if err != nil {
			fmt.Println(err)
			return
		}
		fmt.Println("Added executable to", hostPort, workerLocation)
		return
	}

	if train {
		fmt.Println("New worker at", hostPort)
		w, err := NewWorker(hostPort)
		if err != nil {
			fmt.Println(err)
		}
		defer w.conn.Close()
		fmt.Println("Connected to", hostPort)

		fmt.Println("Training", hostPort)

		session, err := w.conn.NewSession()
		if err != nil {
			fmt.Println(err)
		}
		defer session.Close()
		bytes, err := session.CombinedOutput(workerLocation + " run")
		fmt.Println(string(bytes))
		if err != nil {
			fmt.Print(hostPort, " ")
			fmt.Println(err)
			return
		}
		return
	}

	if run {
		fmt.Println("Hello World.")
		return
	}

	fmt.Println("Usage: [USER=someone] worker [hostname:port] action")
	fmt.Println("  For running a training session on a remote machine, or running locally.")
	fmt.Println("  - USER - optional; is an alternative user to use when connecting via ssh")
	fmt.Println("  - hostname:port - required for setup or train on remote server. bleh.example.com:22")
	fmt.Println("  - action - [run|setup|train]")
	fmt.Println("       - run - run training locally with the supplied settings")
	fmt.Println("       - setup - setup the remote server for training")
	fmt.Println("       - train - run the training on the remote server")
}
