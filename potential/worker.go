package potential

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"strconv"
	"strings"

	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"
)

const workerLocation = "/opt/pworker"
const trainingSettingsLocation = "/opt/pworker_settings.gob"
const networkLocation = "/opt/network.json"

/*
Worker is an instance of a remote training worker. We send it training
data and a neural network and it sends back the neural network after
training it.
*/
type Worker struct {
	conn *ssh.Client
	host string
}

// NewWorker is a worker constructor
func NewWorker(hostAndPort string) (w *Worker, err error) {
	u := os.Getenv("USER")
	w = &Worker{host: hostAndPort}
	keyname := os.Getenv("KEYNAME")
	if keyname == "" {
		keyname = "id_rsa"
	}

	// Dial your ssh server.
	w.conn, err = ssh.Dial("tcp", w.host, &ssh.ClientConfig{
		User: u,
		Auth: []ssh.AuthMethod{
			publicKeyFile(os.Getenv("HOME") + "/.ssh/" + keyname),
		},
	})
	if err != nil {
		return w, err
	}

	return w, nil
}

// TranserExecutable checks the OS and architecture then sends that executable
// up to the worker.
func (w *Worker) TranserExecutable() (err error) {
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
	fmt.Println("Transferring training program to", w.host, remoteOS, remoteArch)

	err = w.SCPSendFile("../worker/worker_"+remoteOS+"_"+remoteArch, workerLocation)
	if err != nil {
		return err
	}
	// file is on the remote server now. need to make it executable though.
	session2, err := w.conn.NewSession()
	if err != nil {
		return err
	}
	defer session2.Close()
	cmd2 := "chmod +x " + workerLocation
	fmt.Println("  ", w.host, cmd2)
	bytes, err = session2.CombinedOutput(cmd2)
	if err != nil {
		return err
	}
	if len(bytes) > 0 {
		fmt.Println("  ", w.host, string(bytes))
	}
	return nil
}

// SCPSendFile sends the local file to the remote location
func (w *Worker) SCPSendFile(localFile string, remoteFileLocation string) (err error) {
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
	remoteFile, err := scp.Create(remoteFileLocation)
	if err != nil {
		return err
	}
	_, err = remoteFile.Write(bytes)
	if err != nil {
		return err
	}
	err = remoteFile.Close()
	if err != nil {
		return err
	}
	return nil
}

// SCPGetFile gets the remote file and returns the bytes
func (w *Worker) SCPGetFile(remoteFileLocation string, toLocalLocation string) (err error) {
	scp, err := sftp.NewClient(w.conn)
	if err != nil {
		return err
	}
	remoteFile, err := scp.Open(remoteFileLocation)
	if err != nil {
		return err
	}
	localFile, err := os.OpenFile(toLocalLocation, os.O_WRONLY, os.ModePerm)
	if err != nil {
		return err
	}
	_, err = remoteFile.WriteTo(localFile)
	if err != nil {
		return err
	}
	return nil
}

// Train runs the training session on a remote server. The remote party will
// be doing `RunWorker()`.
func (w *Worker) Train(localTrainingSettingsJSONLocation string, localTrainingNetworkJSONLocation string) (finalNetwork *Network, err error) {
	fmt.Println("Transferring settings to", w.host)
	err = w.SCPSendFile(localTrainingSettingsJSONLocation, trainingSettingsLocation)
	if err != nil {
		fmt.Println(err)
		return finalNetwork, err
	}
	fmt.Println("Transferring network to", w.host)
	err = w.SCPSendFile(localTrainingNetworkJSONLocation, networkLocation)
	if err != nil {
		fmt.Println(w.host, err)
		return finalNetwork, err
	}
	fmt.Println("Training", w.host, "\n", workerLocation)

	session, err := w.conn.NewSession()
	if err != nil {
		fmt.Println(w.host, err)
		fmt.Println(err)
		return finalNetwork, err
	}
	defer session.Close()

	// do stderr on
	stdout, err := session.StdoutPipe()
	if err != nil {
		return finalNetwork, err
	}
	stderr, err := session.StderrPipe()
	if err != nil {
		return finalNetwork, err
	}
	go io.Copy(os.Stdout, stdout)
	go io.Copy(os.Stderr, stderr)

	err = session.Run(workerLocation)
	if err != nil {
		fmt.Print(w.host, " ")
		fmt.Println(err)
		return finalNetwork, err
	}
	// it is finished; write the remote file back over the local file.
	// deleting the file first prevents some weird error where it was
	// spitting out invalid json; like it was overwriting itself.
	err = os.Remove(localTrainingNetworkJSONLocation)
	if err != nil {
		return finalNetwork, err
	}
	f, err := os.Create(localTrainingNetworkJSONLocation)
	if err != nil {
		return finalNetwork, err
	}
	err = f.Close()
	if err != nil {
		return finalNetwork, err
	}
	err = w.SCPGetFile(networkLocation, localTrainingNetworkJSONLocation)
	if err != nil {
		return finalNetwork, err
	}
	finalNetwork, err = LoadNetworkFromFile(localTrainingNetworkJSONLocation)
	return finalNetwork, err
}

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

func readWorkerfile(filename string) (remoteWorkers []string, remoteWorkerWeights []int, weightTotal int, err error) {
	b, err := ioutil.ReadFile(filename)
	if err != nil {
		return
	}
	rw := strings.Split(string(b), "\n")
	for _, w := range rw {
		if w != "" && !(string(w[0]) == "#") {
			fmt.Println("w=", w)
			parts := strings.Split(w, " ")
			if len(parts) != 2 {
				err = fmt.Errorf("Workfile should have thread weight followed by hostname:port - %s", w)
				return remoteWorkers, remoteWorkerWeights, weightTotal, err
			}
			weight, _ := strconv.Atoi(parts[0])
			hostPort := parts[1]
			remoteWorkers = append(remoteWorkers, hostPort)
			remoteWorkerWeights = append(remoteWorkerWeights, weight)
			weightTotal += weight
		} else {
			fmt.Println("skipping", w)
		}
	}
	return remoteWorkers, remoteWorkerWeights, weightTotal, err
}

// RunWorker is what gets run when this is a remote worker
func RunWorker() (err error) {
	settings, err := LoadTrainingSettingsFromFile(trainingSettingsLocation)
	if err != nil {
		return err
	}
	originalNetwork, err := LoadNetworkFromFile(networkLocation)
	if err != nil {
		return err
	}
	hn, _ := os.Hostname()
	prefix := "<" + hn + ">"
	Train(settings, originalNetwork, prefix)
	err = originalNetwork.SaveToFile(networkLocation)
	if err != nil {
		fmt.Println(prefix, err)
		return err
	}
	return nil
}
