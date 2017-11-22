package main

import (
	"bufio"
	"encoding/gob"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"reflect"
	"strings"
)

var (
	historyFile = "history.gob"
)

func check(e error) {
	if e != nil {
		log.Fatal(e)
	}
}

type Dependencies struct {
	Data map[string]map[string]string
}

// internal structure: jsonData used for parsing package.json file
type jsonData struct {
	Deps    map[string]string `json:"dependencies"`
	DevDeps map[string]string `json:"devDependencies"`
}

func readTxtFile(path string) *Dependencies {

	textFile, err := os.Open(path)
	check(err)
	defer textFile.Close()

	req := make(map[string]string)
	scanner := bufio.NewScanner(textFile)
	for scanner.Scan() {
		line := scanner.Text()

		if strings.HasPrefix(line, "#") || line == "" || line == "\n" {
			continue
		}

		res := strings.Split(line, "==")

		if res[0] == res[len(res)-1] {
			req[res[0]] = "0.0"
		} else {
			req[res[0]] = res[len(res)-1]
		}
	}

	// check for errors
	if err = scanner.Err(); err != nil {
		log.Fatal(err)
	}

	finalMap := make(map[string]map[string]string)
	finalMap[path] = req

	return &Dependencies{Data: finalMap}
}

func readJsonFile(path string) *Dependencies {
	row_json, err := ioutil.ReadFile(path)
	check(err)

	res := jsonData{}
	json.Unmarshal([]byte(row_json), &res)

	mergedDict := make(map[string]string)

	// Iterate dependencies
	for key, val := range res.Deps {
		mergedDict[key] = val
	}

	// Iterage devDependencies
	for key, val := range res.DevDeps {
		mergedDict[key] = val
	}

	finalMap := make(map[string]map[string]string)
	finalMap[path] = mergedDict

	return &Dependencies{Data: finalMap}

}

// Save to gob file
func Save(path string, object interface{}) error {
	file, err := os.Create(path)
	check(err)
	defer file.Close()

	encoder := gob.NewEncoder(file)
	encoder.Encode(object)

	return err
}

// Load from glob
func Load(path string, object interface{}) error {
	file, err := os.Open(path)

	defer file.Close()

	if err == nil {
		decoder := gob.NewDecoder(file)
		err = decoder.Decode(object)
	}

	return err
}

func isMapDiff(mapOne, mapTwo map[string]string) bool {
	equal := reflect.DeepEqual(mapOne, mapTwo)

	if equal {
		log.Println("Equal")
		return false
	}

	log.Println("Different")
	return true

}

func main() {

	dir, err := filepath.Abs(filepath.Dir(os.Args[0]))
	check(err)

	var npmTxt, reqTxt string
	flag.StringVar(
		&npmTxt,
		"npm",
		"",
		"Full path to package.json file. Ex: ./isChangedLinux -npm=/home/package.json")

	flag.StringVar(
		&reqTxt,
		"pip",
		"",
		"Full path to requirements.txt. Ex: ./isChangedLinux -pip=/home/requirements.txt")

	flag.Parse()

	fileToCheck := ""
	inputData := new(Dependencies)
	if reqTxt != "" {
		fileToCheck = reqTxt
		requirements, err := filepath.Abs(reqTxt)
		check(err)
		//log.Println("[pip] requirements.txt:", requirements)

		inputData = readTxtFile(requirements)
		//log.Println("[pip] New Requirements:", inputData.Data)
	} else if npmTxt != "" {
		fileToCheck = npmTxt
		packageJson, err := filepath.Abs(npmTxt)

		check(err)
		//log.Println("[npm] package.json:", packageJson)

		inputData = readJsonFile(packageJson)
		//log.Println("[npm] New npm packages:", inputData.Data)
	}

	// Load old data from history.gob
	var oldData = new(Dependencies)
	oldFile, _ := filepath.Abs(dir + "/" + historyFile)
	err = Load(oldFile, oldData)
	if err != nil {
		log.Printf("Can't open history.gob. Error: %s\n", err)
		value := make(map[string]map[string]string)
		value[reqTxt] = make(map[string]string)
		oldData.Data = value
	}

	// Check if this PATH checked before
	if _, ok := oldData.Data[fileToCheck]; !ok {
		value := make(map[string]string)
		oldData.Data[fileToCheck] = value
	}

	//log.Println("Old Requirements: ", oldData.Data)

	// Compare old and new data
	changed := isMapDiff(oldData.Data[fileToCheck], inputData.Data[fileToCheck])
	log.Println("Data changed: ", changed)
	log.Println("===========================================")

	if changed {
		oldData.Data[fileToCheck] = inputData.Data[fileToCheck]
		err = Save(dir+"/"+historyFile, oldData)
		if err != nil {
			log.Printf("Can't save new data to history.gob. Error: %s", err)
		}
		os.Exit(10)
	}
	os.Exit(11)

	fmt.Println("Put arguments: npm or pip. Example: ./isChangedLinux -npm=/package.json")
}
