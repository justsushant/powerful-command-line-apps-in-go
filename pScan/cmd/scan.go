/*
Copyright © 2023 justsushant
*/
package cmd

import (
	"fmt"
	"io"
	"os"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"pragprog.com/rggo/cobra/pScan/scan"
)

// scanCmd represents the scan command
var scanCmd = &cobra.Command{
	Use:   "scan",
	Short: "Run a port scan on the hosts",
	RunE: func(cmd *cobra.Command, args []string) error {
		hostsFile, err := cmd.Flags().GetString("hosts-file")
		if err != nil {
			return err
		}

		timeout, err := cmd.Flags().GetInt("timeout")
		if err != nil {
			return err
		}

		portsSlice, err := cmd.Flags().GetIntSlice("ports")
		if err != nil {
			return err
		}

		portsRange, err := cmd.Flags().GetString("portRange")
		if err != nil {
			return err
		}

		return scanAction(os.Stdout, hostsFile, aggregatePorts(portsSlice, portsRange), time.Duration(timeout))
	},
}

func init() {
	rootCmd.AddCommand(scanCmd)

	scanCmd.Flags().IntSliceP("ports", "p", []int{22, 80, 443}, "ports to scan")
	scanCmd.Flags().StringP("portRange", "r", "", "port range to scan, like lowerBound-upperBound")
	scanCmd.Flags().IntP("timeout", "t", 1, "timeout for port scan in seconds")
	// scanCmd.Flags().BoolP("closed", "c", false, "show only closed ports")
	// scanCmd.Flags().BoolP("open", "o", false, "show only open ports")

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// scanCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// scanCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

func aggregatePorts(portsSlice []int, portRange string) []int {
	if portRange == "" {
		return portsSlice
	}

	// to store the ports aggregated
	var ports []int
	ports = append(ports, portsSlice...)

	r := strings.Split(portRange, "-")
	if len(r) != 2 {
		fmt.Errorf("Invalid port range")
	}

	lowerBound, err := strconv.Atoi(r[0])
	if err != nil && validatePort(lowerBound) {
		fmt.Errorf("Invalid lower bound of the range: %q", err)
	}

	upperBound, err := strconv.Atoi(r[1])
	if err != nil && validatePort(upperBound) {
		fmt.Errorf("Invalid upper bound of the range: %q", err)
	}

	if lowerBound > upperBound {
		fmt.Errorf("Upper bound of the range should be greater than lower bound")
	}

	for i := lowerBound; i <= upperBound; i++ {
		if !slices.Contains(portsSlice, i) {
			ports = append(ports, i)
		}
	}

	return ports
}

func validatePort(port int) bool {
	var (
		LOWER_BOUND = 1
		UPPER_BOUND = 65535
	)

	if port >= LOWER_BOUND && port <= UPPER_BOUND {
		return true
	}
	return false
}

func scanAction(out io.Writer, hostsFile string, ports []int, timeout time.Duration) error {
	hl := &scan.HostsList{}

	if err := hl.Load(hostsFile); err != nil {
		return err
	}

	results := scan.Run(hl, ports, timeout)
	return printResults(out, results)
}

func printResults(out io.Writer, results []scan.Results) error {
	message := ""

	for _, r := range results {
		message += fmt.Sprintf("%s: ", r.Host)

		if r.NotFound {
			message += fmt.Sprint("Host not found\n\n")
			continue
		}

		message += fmt.Sprintln()

		for _, p := range r.PortStates {
			message += fmt.Sprintf("\t%d: %s\n", p.Port, p.Open)
		}

		message += fmt.Sprintln()
	}

	_, err := fmt.Fprint(out, message)
	return err
}
