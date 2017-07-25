package main

import (
	"bytes"
	"fmt"
	"os"
	"time"

	"log"

	"github.com/hashicorp/terraform/communicator"
	"github.com/hashicorp/terraform/communicator/remote"
	"github.com/hashicorp/terraform/terraform"
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

	// Upload
	f, err := os.Open("./bin/debian/facter")
	if err != nil {
		log.Fatalln(err)
	}

	err = comm.Upload("./facter/facter", f)
	if err != nil {
		log.Fatalln(err)
	}

	// output buffers for SSH
	var outBuf, errBuf bytes.Buffer

	cmd := remote.Cmd{
		Command: "chmod +x ./facter/facter",
		Stdin:   os.Stdin,
		Stdout:  &outBuf,
		Stderr:  &errBuf,
	}

	err = comm.Start(&cmd)
	if err != nil {
		log.Fatalln(err)
	}
	cmd.Wait()

	cmd = remote.Cmd{
		Command: "FACTERLIB=~/facter ./facter/facter -j",
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

	cmd = remote.Cmd{
		Command: "rm -r ./facter",
		Stdin:   os.Stdin,
		Stdout:  &outBuf,
		Stderr:  &errBuf,
	}

	err = comm.Start(&cmd)
	if err != nil {
		log.Fatalln(err)
	}
	cmd.Wait()

	comm.Disconnect()

	t1 := time.Now()
	fmt.Printf("Duration %v\n", t1.Sub(t0))
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
