package main

import (
	"flag"
	"bufio"
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
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

	reader := bufio.NewReader(os.Stdin)
	currDir, err := os.Getwd()
	
	inputFp := filepath.Join(currDir, *inputFile)

	fmt.Println("Reading from", inputFp)

	file, err := os.Open(inputFp)
	handleErr(err)
	defer file.Close()

	scanner := bufio.NewScanner(file)
	handleErr(err)

	outputMap := map[int]string{}

	lineCount := 0
	for scanner.Scan() {
		res, err := http.Get(scanner.Text())
		handleErr(err)
		defer res.Body.Close()

		if res.StatusCode >= 300 {
			log.Fatal("Request", scanner.Text(), "died with", res.Status)
		}

		body, err := ioutil.ReadAll(res.Body)

		if *printResponses {
			fmt.Println(string(body))
		}

		hasher := md5.New()
		hasher.Write(body)
		hash := hex.EncodeToString(hasher.Sum(nil))

		fmt.Println(scanner.Text() + ": " + hash)
		outputMap[lineCount] = scanner.Text() + ": " + hash + "\n"

		lineCount++
	}

	output := filepath.Join(currDir, *outputFile)

	err = ioutil.WriteFile(output, []byte(output), 0644)
	handleErr(err)

	fmt.Println("Done, wrote to: " + output)
}

func handleErr(err error) {
	if err != nil {
		log.Fatal(err)
	}
}
