package main

import (
	"bufio"
	"crypto/md5"
	"encoding/hex"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

func main() {
	inputFile := flag.String("input", "", "Input file")
	outputFile := flag.String("output", "", "Output file")
	compareToFile := flag.String("compareTo", "", "Compare to file")
	printResponses := flag.Bool("print", false, "Whether to print responses. Default: no")
	concurrent := flag.Int("concurrent", 1, "Concurrent requests. Default: 1")

	flag.Parse()

	if *inputFile == "" {
		log.Fatal("Please provide input file using -input flag")
	}

	if *outputFile == "" {
		log.Fatal("Please provide output file using -output flag")
	}

	currDir, err := os.Getwd()

	*inputFile = filepath.Join(currDir, *inputFile)

	if *compareToFile != "" {
		*compareToFile = filepath.Join(currDir, *compareToFile)
		if _, err := os.Stat(*compareToFile); err != nil {
			log.Fatal("CProblem with compare file", *compareToFile)
		}
	}

	fmt.Println("Reading from", *inputFile)

	file, err := os.Open(*inputFile)
	handleErr(err)
	defer file.Close()

	scanner := bufio.NewScanner(file)
	handleErr(err)

	lineCount := 0
	start := time.Now()

	sem := make(chan struct{}, *concurrent)
	lines := []chan string{}

	for scanner.Scan() {
		url := scanner.Text()

		lines = append(lines, make(chan string))
		sem <- struct{}{}
		go fetch(url, *printResponses, lines[lineCount], sem)

		lineCount++
	}

	var compareCh chan []string
	var compareLines []string

	if *compareToFile != "" {
		compareCh = make(chan []string)

		go readCompareFile(compareToFile, compareCh)
	}

	if compareCh != nil {
		compareLines = <- compareCh
	}

	output := ""
	for i := 0; i < lineCount; i++ {
		line := <- lines[i]
		output += line + "\n"

		if compareLines != nil {
			if len(compareLines) <= i {
				fmt.Println("Comparefile does not have line", i)
			} else if compareLines[i] != line {
				log.Fatal("Line ", i, " does not match\nOutput:\n", line, "\nComparison:\n", compareLines[i])
			}
		}
	}

	fmt.Println("Elapsed:", time.Since(start))

	*outputFile = filepath.Join(currDir, *outputFile)

	err = ioutil.WriteFile(*outputFile, []byte(output), 0644)
	handleErr(err)

	fmt.Println("Done, wrote to: " + *outputFile)
}

func fetch(url string, printResponses bool, ch chan<- string, sem <-chan struct{}) {
	res, err := http.Get(url)
	handleErr(err)
	defer res.Body.Close()

	if res.StatusCode >= 300 {
		log.Fatal("Request", url, "died with", res.Status)
	}

	body, err := ioutil.ReadAll(res.Body)

	if printResponses {
		fmt.Println(string(body))
	}

	hasher := md5.New()
	hasher.Write(body)
	hash := hex.EncodeToString(hasher.Sum(nil))

	response := url + ": " + hash
	fmt.Println(response)

	ch <- response
	<-sem
}

func handleErr(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

func readCompareFile(compareToFile *string, compareCh chan<- []string) {
	file, err := os.Open(*compareToFile)
	handleErr(err)
	defer file.Close()

	scanner := bufio.NewScanner(file)
	handleErr(err)

	compareLines := []string{}
	for scanner.Scan() {
		compareLines = append(compareLines, scanner.Text())
	}

	compareCh <- compareLines
}