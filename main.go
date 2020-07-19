package main

import (
	"flag"
	"time"
	"bufio"
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
)

func main() {
	inputFile := flag.String("input", "", "Input file")
	outputFile := flag.String("output", "", "Output file")
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
	
	inputFp := filepath.Join(currDir, *inputFile)

	fmt.Println("Reading from", inputFp)

	file, err := os.Open(inputFp)
	handleErr(err)
	defer file.Close()

	scanner := bufio.NewScanner(file)
	handleErr(err)

	lineCount := 0
	start := time.Now()

	sem := make(chan struct{}, *concurrent)
	lines := map[int](chan string){}

	for scanner.Scan() {
		url := scanner.Text()

		lines[lineCount] = make(chan string)

		sem <- struct{}{}
		go fetch(lineCount, url, *printResponses, lines[lineCount], sem)

		lineCount++
	}

	output := ""
	for i := 0; i < lineCount; i++ {
		output += <- lines[i] + "\n"
	}

	fmt.Println("Elapsed:", time.Since(start))

	outputFp := filepath.Join(currDir, *outputFile)

	err = ioutil.WriteFile(outputFp, []byte(output), 0644)
	handleErr(err)

	fmt.Println("Done, wrote to: " + outputFp)
}

func fetch(count int, url string, printResponses bool, ch chan<- string, sem <-chan struct{}) {
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

	<- sem
	ch <- response
}

func handleErr(err error) {
	if err != nil {
		log.Fatal(err)
	}
}
