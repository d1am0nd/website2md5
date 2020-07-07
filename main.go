package main

import (
	"os"
	"log"
	"io/ioutil"
	"strings"
	"encoding/hex"
	"crypto/md5"
	"net/http"
	"path/filepath"
	"fmt"
	"bufio"
)

func main() {
	reader := bufio.NewReader(os.Stdin)
	currDir, err := os.Getwd()

	fmt.Println("Provide path to file with urls")
	fp, err := reader.ReadString('\n')
	handleErr(err)

	fmt.Println("Print responses? (default no)")
	printChoice, err := reader.ReadString('\n')
	handleErr(err)

	printResponse := strings.HasPrefix(printChoice, "y") || strings.HasPrefix(printChoice, "Y")

	fp = filepath.Join(currDir, fp)
	fp = strings.TrimSuffix(fp, "\n")

	fmt.Println(fp)

	file, err := os.Open(fp)
	handleErr(err)
	defer file.Close()

	scanner := bufio.NewScanner(file)
	handleErr(err)

	fileOutput := ""

	for scanner.Scan() {
		res, err := http.Get(scanner.Text())
		handleErr(err)
		defer res.Body.Close()

		body, err := ioutil.ReadAll(res.Body)

		if printResponse {
			fmt.Println(string(body))
		}

		hasher := md5.New()
		hasher.Write(body)
		hash := hex.EncodeToString(hasher.Sum(nil))

		fmt.Println(scanner.Text() + ": " + hash)
		fileOutput += scanner.Text() + ": " + hash + "\n"
	}

	fmt.Println("Provide file to output to (defaults output.txt)")
	output, err := reader.ReadString('\n')
	handleErr(err)
	output = strings.TrimSuffix(output, "\n")
	if len(output) == 0 {
		output = "output.txt"
	}

	output = filepath.Join(currDir, output)

	fmt.Println(fileOutput)
	fmt.Println(output)

	err = ioutil.WriteFile(output, []byte(fileOutput), 0644)
	handleErr(err)
}

func handleErr(err error) {
	if err != nil {
		log.Fatal(err)
	}
}