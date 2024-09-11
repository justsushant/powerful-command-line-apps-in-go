package main_test

import (
	"fmt"
	"io"
	"os"            // uses os types
	"os/exec"       // executes external commands
	"path/filepath" // deals with directory paths
	"runtime"       // identifies the running os
	"testing"
)

var (
	binName = "todo"
	fileName = ".test.json"
)

func TestMain(m *testing.M) {
	fmt.Println("Building tool...")

	if runtime.GOOS == "windows" {
		binName += ".exe"
	}

	build := exec.Command("go", "build", "-o", binName)
	if err := build.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Can't build tool %s: %s", binName, err)
		os.Exit(1)
	}

	fmt.Println("Running tests....")
	result := m.Run()

	fmt.Println("Cleaning up...")
	os.Remove(binName)
	os.Remove(fileName)

	os.Exit(result)
}

func TestTodoCLI(t *testing.T) {
	task := "test task number 1"
	task2 := "test task number 2"

	dir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}

	cmdPath := filepath.Join(dir, binName)

	t.Run("AddNewTaskFromArguments", func(t *testing.T) {
		cmd := exec.Command(cmdPath, "-add", task)

		if err := cmd.Run(); err != nil {
			t.Fatal(err)
		}
	})

	t.Run("AddNewTaskFromSTDIN", func(t *testing.T) {
		cmd := exec.Command(cmdPath, "-add")
		cmdStdIn, err := cmd.StdinPipe()
		if err != nil {
			t.Fatal(err)
		}

		io.WriteString(cmdStdIn, task2)
		cmdStdIn.Close()

		if err := cmd.Run(); err != nil {
			t.Fatal(err)
		}
	})

	t.Run("ListTasks", func(t *testing.T) {
		cmd := exec.Command(cmdPath, "-list")
		out, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatal(err)
		}

		expected := fmt.Sprintf("  1: %s\n  2: %s\n", task, task2)

		if expected != string(out) {
			t.Errorf("Expected %q, got %q instead\n", expected, string(out))
		}
	})

	t.Run("CompleteTask", func(t *testing.T) {
		cmd := exec.Command(cmdPath, "-complete", "1")
		out, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatal(err)
		}

		expected := ""

		if expected != string(out) {
			t.Errorf("Expected %q, got %q instead\n", expected, string(out))
		}
	})

	t.Run("ListIncompleteTasks", func (t *testing.T) {
		cmd := exec.Command(cmdPath, "-incomplete")
		out, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatal(err)
		}

		expected := fmt.Sprintf("  1: %s\n", task2)

		if expected != string(out) {
			t.Errorf("Expected %q, got %q instead\n", expected, string(out))
		}
	})

	// t.Run("ListTasksVerbose", func(t *testing.T) {
	// 	cmd := exec.Command(cmdPath, "-verbose")
	// 	out, err := cmd.CombinedOutput()
	// 	if err != nil {
	// 		t.Fatal(err)
	// 	}
		
	// 	expected := fmt.Sprintf("  Task 1: %s\n  2: %s\n", task, task2)
	// 	if expected != string(out) {
	// 		t.Errorf("Expected %q, got %q instead\n", expected, string(out))
	// 	}
	// })

	t.Run("DeleteAndListTask", func(t *testing.T) {
		// deletes the task
		cmd := exec.Command(cmdPath, "-delete", "1")
		if err := cmd.Run(); err != nil {
			t.Fatal(err)
		}
		
		// list the tasks
		cmd = exec.Command(cmdPath, "-list")
		out, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatal(err)
		}

		// expected list of tasks
		expected := fmt.Sprintf("  1: %s\n", task2)

		if expected != string(out) {
			t.Errorf("Expected %q, got %q instead\n", expected, string(out))
		}
	})



	// t.Run("CompleteAndListTask", func(t *testing.T) {
	// 	// add multiple new tasks
	// 	newTask2 := "test task number 2"
	// 	newTask3 := "test task number 3"

	// 	cmd := exec.Command(cmdPath, "-task", newTask2)
	// 	if err := cmd.Run(); err != nil {
	// 		t.Fatal(err)
	// 	}
	// 	cmd = exec.Command(cmdPath, "-task", newTask3)
	// 	if err := cmd.Run(); err != nil {
	// 		t.Fatal(err)
	// 	}

	// 	// complete the new task 2
	// 	cmd = exec.Command(cmdPath, "-complete", "1")
	// 	if err := cmd.Run(); err != nil {
	// 		t.Fatal(err)
	// 	}

	// 	// check the toDo list
	// 	cmd = exec.Command(cmdPath, "-list")
	// 	out, err := cmd.CombinedOutput()
	// 	if err != nil {
	// 		t.Fatal(err)
	// 	}

	// 	expected := newTask3 + "\n"

	// 	if expected != string(out) {
	// 		t.Errorf("Expected %q, got %q instead\n", expected, string(out))
	// 	}
	// })
}