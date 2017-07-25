package main

import (
	"bytes"
	"fmt"
	"os"

	"log"

	"github.com/hashicorp/terraform/communicator"
	"github.com/hashicorp/terraform/communicator/remote"
	"github.com/hashicorp/terraform/terraform"
)

func main() {
	fmt.Println("tfcomm")

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

	var fspew = DebugSpewFunc(spew)

	err = comm.Connect(fspew)
	if err != nil {
		log.Fatalln(err)
	}

	// UploadDir(dst, src)
	err = comm.UploadDir(".", "./facter")
	if err != nil {
		log.Fatalln(err)
	}

	var outBuf, errBuf bytes.Buffer

	cmd := remote.Cmd{
		Command: "FACTERLIB=~/facter facter -j",
		Stdin:   os.Stdin,
		Stdout:  &outBuf,
		Stderr:  &errBuf,
	}

	err = comm.Start(&cmd)
	if err != nil {
		log.Fatalln(err)
	}
	cmd.Wait()
	fmt.Println(outBuf.String())

	comm.Disconnect()

}

func spew(msg string) {
	log.Println(msg)
}

// DebugSpewFunc --
type DebugSpewFunc func(string)

// Output -- spew implementation function
func (f DebugSpewFunc) Output(msg string) {
	f(msg)
}
