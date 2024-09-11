package main

import (
	"encoding/csv"
	"fmt"
	"io"
	"strconv"
)

// statsFunc defines a generic statistical function
type statsFunc func(data []float64) float64

func sum(data []float64) float64 {
	sum := 0.0

	for _, v := range data {
		sum += v
	}

	return sum
}

func avg (data []float64) float64 {
	return sum(data)/float64(len(data))
}

func min (data []float64) float64 {
	var min float64 = data[0]

	for _, num := range data[1:] {
		if !(num >= min) {
			min = num
		}
	}

	return min
}

func max (data []float64) float64 {
	var max float64 = data[0]

	for _, num := range data[1:] {
		if num > max {
			max = num
		}
	}

	return max
}

func csv2float(r io.Reader, column int) ([]float64, error) {
	// create the CSV Reader used to read in data from CSV files
	cr := csv.NewReader(r)
	//  reuse the same slice for each read operation to reduce the memory allocation
	cr.ReuseRecord = true

	// adjusting for zero based index
	column--

	var data []float64

	// looping through all records
	for i := 0; ; i++ {
		row, err := cr.Read()

		// reached the end of the file
		if err == io.EOF {
			break
		}

		if err != nil {
			return nil, fmt.Errorf("Cannot read data from file: %w", err)
		}

		// skip the header of csv file
		if i == 0 {
			continue
		}

		// checking number of columns in csv file
		if len(row) <= column {
			// file does not have that many columns
			return nil, fmt.Errorf("%w: File has only %d columns", ErrInvalidColumn, len(row))
		}

		// try to convert data read into a float number
		v, err := strconv.ParseFloat(row[column], 64)
		if err != nil {
			return nil, fmt.Errorf("%w: %s", ErrNotNumber, err)
		}

		data = append(data, v)
	}

	// return the slice of float64 and nil error
	return data, nil
}