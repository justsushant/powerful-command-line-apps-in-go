package main

import (
	"bufio" // read data from STDIN
	"flag"
	"fmt"
	"io" // for io.Reader interface
	"os"
	"strings"
	"time"

	"pragprog.com/rggo/interacting/todo"
)

// env variable for fileName
const fileNameEnvVar = "TODO_FILENAME"

// default file name
var todoFileName = ".todo.json"

func main() {
	flag.Usage = func () {
		// fmt.Fprintf(flag.CommandLine.Output(), "%s tool. Developed for The Pragmatic Bookshelf\n", os.Args[0])
		fmt.Fprintf(flag.CommandLine.Output(), "Developed for The Pragmatic Bookshelf\n")
		fmt.Fprintf(flag.CommandLine.Output(), "Copyright 2020\n")
		fmt.Fprintf(flag.CommandLine.Output(), "Usage information:\n")
		flag.PrintDefaults()

	}

	// setting the command line flags
	add := flag.Bool("add", false, "Add to be included in the ToDo list")
	list := flag.Bool("list", false, "List all tasks")
	complete := flag.Int("complete", 0, "Item to be completed")
	delete := flag.Int("delete", 0, "Deletes an item from the list")
	verbose := flag.Bool("verbose", false, "Shows verbose output")
	incomplete := flag.Bool("incomplete", false, "Shows only incomplete tasks")

	// parsing the command line flags
	flag.Parse()

	// check if user has defined env var for custom filename
	if os.Getenv(fileNameEnvVar) != "" {
		todoFileName = os.Getenv(fileNameEnvVar)
	}

	// defining a toDo items list
	l := &todo.List{}

	// reading toDo items from file
	if err := l.Get(todoFileName); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	// decide what to do based on the number of arguments provided
	switch {
	case *list:
		// list current toDo items
		fmt.Print(l)
	case *complete > 0:
		// complete the given item
		if err := l.Complete(*complete); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}

		// save the new list
		if err := l.Save(todoFileName); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
	case *add:
		// when any arguments (excluding flags) are provided
		// they will be used as the new task
		t, err := getTask(os.Stdin, flag.Args()...)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}

		// add the task
		l.Add(t)

		// save the new list
		if err := l.Save(todoFileName); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
	case *delete > 0:
		// delete the given item
		if err := l.Delete(*delete); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}

		// save the new list
		if err := l.Save(todoFileName); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
	case *verbose:
		verboseOutput := ""
		for i, item := range *l {
			var status string
			if item.Done {
				status = fmt.Sprintf("Completed at %v", item.CompletedAt.Format(time.RFC1123))
			} else {
				status = "Incomplete"
			}
			verboseOutput += fmt.Sprintf("  Task %d: %s | Created at: %v | Status: %s\n", i+1, item.Task, item.CreatedAt.Format(time.RFC1123), status)
		}
		fmt.Println(verboseOutput)
	case *incomplete:
		incompleteTasks := &todo.List{}

		for _, item := range *l {
			if !item.Done {
				*incompleteTasks = append(*incompleteTasks, item)
			}
		}

		fmt.Print(incompleteTasks)
	default:
		// invalid flag
		fmt.Fprintln(os.Stderr, "Invalid Option")
		os.Exit(1)
	}

}

// io.Reader interface can be used whenever you expect to read data
// files, buffers, archives, HTTP requests, and others satisfy this interface
// By using it, you decouple your implementation from specific types, 
// allowing your code to work with any types that implement the io.Reader interface

// getTask function decides where to get the description for a new​ task from: arguments or STDIN
func getTask(r io.Reader, args ...string) (string, error) {
	if len(args) > 0 {
		return strings.Join(args, " "), nil
	}

	s := bufio.NewScanner(r)
	s.Scan()
	if err := s.Err(); err != nil {
		return "", err
	}

	if len(s.Text()) == 0 {
		return "", fmt.Errorf("Task cannot be blank")
	}

	return s.Text(), nil
}