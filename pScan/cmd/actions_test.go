package cmd

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"strings"
	"net"
	"strconv"
	"testing"
	"pragprog.com/rggo/cobra/pScan/scan"
)

func setup(t *testing.T, hosts []string, initList bool) (string, func()) {
	// create temp file
	tf, err := os.CreateTemp("", "pScan")
	if err != nil {
		t.Fatal(err)
	}
	tf.Close()

	// initialise list if needed
	if initList {
		hl := &scan.HostsList{}

		for _, h := range hosts {
			hl.Add(h)
		}

		if err := hl.Save(tf.Name()); err != nil {
			t.Fatal(err)
		}
	}

	// return temp file name and cleanup function
	return tf.Name(), func() {
		os.Remove(tf.Name())
	}
}

func TestHostActions(t *testing.T) {
	// define hosts for actions test
	hosts := []string{
		"host1",
		"host2",
		"host3",
	}

	// test cases for Action test
	testCases := []struct {
		name string
		args []string
		expectedOut string
		initList bool
		actionFunction func(io.Writer, string, []string) error
	}{
		{
			name: "AddAction",
			args: hosts,
			expectedOut: "Added host: host1\nAdded host: host2\nAdded host: host3\n",
			initList: false,
			actionFunction: addAction,
		},
		{
			name: "ListAction",
			expectedOut: "host1\nhost2\nhost3\n",
			initList: true,
			actionFunction: listAction,
		},
		{
			name: "DeleteAction",
			args: []string{"host1", "host2"},
			expectedOut: "Deleted host: host1\nDeleted host: host2\n",
			initList: true,
			actionFunction: deleteAction,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// setup action test
			tf, cleanup := setup(t, hosts, tc.initList)
			defer cleanup()

			// define var to capure action output
			var out bytes.Buffer

			// execute action and capture output
			if err := tc.actionFunction(&out, tf, tc.args); err != nil {
				t.Fatalf("Expected no error, got %q\n", err)
			}

			// Test Actions output
			if out.String() != tc.expectedOut {
				t.Errorf("Expected output %q, got %q\n", tc.expectedOut, out.String())
			}
		})
	}
}

func TestIntegration(t *testing.T) {
	// define hosts for integration test
	hosts := []string{
		"host1",
		"host2",
		"host3",
	}

	// setup integration test
	tf, cleanup := setup(t, hosts, false)
	defer cleanup()

	delHost := "host2"
	hostsEnd := []string{
		"host1",
		"host3",
	}

	// define var to capture output
	var out bytes.Buffer

	// define expected output for all actions
	expectedOut := ""
	for _, v := range hosts {
		expectedOut += fmt.Sprintf("Added host: %s\n", v)
	}
	expectedOut += strings.Join(hosts, "\n")
	expectedOut += fmt.Sprintln()
	expectedOut += fmt.Sprintf("Deleted host: %s\n", delHost)
	expectedOut += strings.Join(hostsEnd, "\n")
	expectedOut += fmt.Sprintln()
	for _, v := range hostsEnd {
		expectedOut += fmt.Sprintf("%s: Host not found\n", v)
		expectedOut += fmt.Sprintln()
	}

	// actions for integration test

	// add hosts to the list
	if err := addAction(&out, tf, hosts); err != nil {
		t.Fatalf("Expected no error, got %q\n", err)
	}

	// list hosts
	if err := listAction(&out, tf, nil); err != nil {
		t.Fatalf("Expected no error, got %q\n", err)
	}

	// delete host2
	if err := deleteAction(&out, tf, []string{delHost}); err != nil {
		t.Fatalf("Expected no error, got %q\n", err)
	}

	// list hosts after delete
	if err := listAction(&out, tf, nil); err != nil {
		t.Fatalf("Expected no error, got %q\n", err)
	}

	// scan hosts
	if err := scanAction(&out, tf, nil, 1); err != nil {
		t.Fatalf("Expected no error, got %q\n", err)
	}

	// test integration output
	if out.String() != expectedOut {
		t.Errorf("Expected output %q, got %q\n", expectedOut, out.String())
	}
}

func TestScanAction(t *testing.T) {
	// define hosts for scan test
	hosts := []string{
		"localhost",
		"unknownhostonmachine",
	}

	// setup scan test
	tf, cleanup := setup(t, hosts, true)
	defer cleanup()

	ports := []int{}

	//Init ports, 1 open, 1 closed
	for i := 0; i < 2; i++ {
		ln, err := net.Listen("tcp", net.JoinHostPort("localhost", "0"))
		if err != nil {
			t.Fatal(err)
		}
		defer ln.Close()

		_, portStr, err := net.SplitHostPort(ln.Addr().String())
		if err != nil {
			t.Fatal(err)
		}

		port, err := strconv.Atoi(portStr)
		if err != nil {
			t.Fatal(err)
		}

		ports = append(ports, port)
		if i == 1 {
			ln.Close()
		}
	}

	// define expected output for scan action
	expectedOut := fmt.Sprintln("localhost: ")
	expectedOut += fmt.Sprintf("\t%d: open\n", ports[0])
	expectedOut += fmt.Sprintf("\t%d: closed\n", ports[1])
	expectedOut += fmt.Sprintln()
	expectedOut += fmt.Sprintln("unknownhostonmachine: Host not found")
	expectedOut += fmt.Sprintln()

	// to capture the scan output
	var out bytes.Buffer

	// execute scan and capture output
	if err := scanAction(&out, tf, ports, 1); err != nil {
		t.Fatalf("Expected no error, got %q\n", err)
	}

	// test scan output
	if out.String() != expectedOut {
		t.Errorf("Expected output %q, got %q\n", expectedOut, out.String())
	}
}

func TestDocsAction(t *testing.T) {
	// creating temp directory
	dir, err := os.MkdirTemp("", "pScan")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(dir)

	// buffer to catch output
	var out bytes.Buffer
	// expected output
	var expectedOut = fmt.Sprintf("Documentation successfully created in %s\n", dir)

	if err = docsAction(&out, dir); err != nil {
		t.Fatalf("Expected no error, got %q\n", err)
	}

	if out.String() != expectedOut {
		t.Errorf("Expected output %q, got %q\n", expectedOut, out.String())
	}

	// SHOULD WE CHECK THE NUMBER OF FILE IT GENERATES??
	// files, err := os.ReadDir(dir)
	// if err != nil {
	// 	t.Fatal(err)
	// }

	// if len(files) >= 8 {
	// 	t.Errorf("Expected atleast 8 markdown")
	// }
	
	// fmt.Printf("NUMBER OF COMMANDS: %d\n\n", len(rootCmd.Commands()))
	// for _, c := range rootCmd.Commands() {
	// 	fmt.Println(c)
	// }
}