package main

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"time"

	"log"

	"github.com/hashicorp/terraform/communicator/remote"
	"github.com/hashicorp/terraform/terraform"
	"github.com/hpcloud/terraform/communicator"
)

func main() {
	fmt.Println("tfcomm")
	t0 := time.Now()

	state := terraform.InstanceState{
		Ephemeral: terraform.EphemeralState{
			ConnInfo: map[string]string{
				"type": `ssh`,
				"user": `gertd`,
				"host": `test-ubuntu-kvm`,
			},
		},
	}

	comm, err := communicator.New(&state)
	if err != nil {
		log.Fatalln(err)
	}
	ecom := extCommunicator{Communicator: comm}

	var fspew = debugSpewFunc(spew)

	err = ecom.Connect(fspew)
	if err != nil {
		log.Fatalln(err)
	}

	// UploadDir(dst, src)
	err = ecom.UploadDir(".", "./facter")
	if err != nil {
		log.Fatalln(err)
	}

	// Upload
	f, err := os.Open("./bin/debian/facter")
	if err != nil {
		log.Fatalln(err)
	}

	err = ecom.Upload("./facter/facter", f)
	if err != nil {
		log.Fatalln(err)
	}

	if err := ecom.execCmd(`chmod +x ./facter/facter`); err != nil {
		log.Fatalln(err)
	}

	if err := ecom.execCmd(`FACTERLIB=~/facter ./facter/facter -j`); err != nil {
		log.Fatalln(err)
	}

	if err := ecom.execCmd(`rm -r ./facter`); err != nil {
		log.Fatalln(err)
	}

	ecom.Disconnect()

	t1 := time.Now()
	fmt.Printf("Duration %v\n", t1.Sub(t0))

}

type extCommunicator struct {
	communicator.Communicator
}

func (c extCommunicator) execCmd(command string) error {
	// output buffers for SSH
	var outBuf, errBuf bytes.Buffer

	cmd := remote.Cmd{
		Command: command,
		Stdin:   os.Stdin,
		Stdout:  &outBuf,
		Stderr:  &errBuf,
	}

	err := c.Start(&cmd)
	if err != nil {
		log.Fatalln(err)
	}
	cmd.Wait()

	go io.Copy(os.Stdout, &outBuf)
	go io.Copy(os.Stderr, &errBuf)

	return nil
}

func spew(msg string) {
	log.Println(msg)
}

type debugSpewFunc func(string)

func (f debugSpewFunc) Output(msg string) {
	f(msg)
}
