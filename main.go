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
	packageFile, _      = filepath.Abs("/package.gob")
	requirementsFile, _ = filepath.Abs("/requirements.gob")
)

func check(e error) {
	if e != nil {
		log.Fatal(e)
	}
}

type Dependencies struct {
	Data map[string]string
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

	return &Dependencies{Data: req}
}

type jsonData struct {
	Deps    map[string]string `json:"dependencies"`
	DevDeps map[string]string `json:"devDependencies"`
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

	return &Dependencies{Data: mergedDict}

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

	var pkgJson, reqTxt string
	flag.StringVar(
		&pkgJson,
		"npm",
		"",
		"Relative path to package.json file. Ex: ./isChangedLinux -npm=/package.json")

	flag.StringVar(
		&reqTxt,
		"pip",
		"",
		"Relative path to requirements.txt. Ex: ./isChangedLinux -pip=/requirements.txt")

	flag.Parse()

	if reqTxt != "" {
		requirements, err := filepath.Abs(dir + reqTxt)
		log.Println("requirements.txt:", requirements)

		newReq := readTxtFile(requirements)
		//log.Println("New Requirements: ", newReq.Data)

		// Load old data
		var oldReq = new(Dependencies)
		oldFile, _ := filepath.Abs(dir + requirementsFile)
		err = Load(oldFile, oldReq)
		if err != nil {
			oldReq = new(Dependencies)
		}

		//log.Println("Old Requirements: ", oldReq.Data)

		// Compare old and new requirements.txt
		changed := isMapDiff(oldReq.Data, newReq.Data)
		log.Println("requirements.txt changed: ", changed)
		log.Println("===========================================")

		if changed {
			err = Save(dir+requirementsFile, newReq)
			if err != nil {
				log.Printf("Can't save previous requirements. Error: %s", err)
			}
			os.Exit(10)
		}
		os.Exit(11)
	}

	if pkgJson != "" {
		packageJson, err := filepath.Abs(dir + pkgJson)

		log.Println("package.json:", packageJson)

		newPack := readJsonFile(packageJson)
		//log.Println(newPack)

		// Load old data
		var oldPack = new(Dependencies)
		oldFile, _ := filepath.Abs(dir + packageFile)
		err = Load(oldFile, oldPack)
		if err != nil {
			oldPack = new(Dependencies)
		}

		changed := isMapDiff(oldPack.Data, newPack.Data)
		log.Println("package.json changed: ", changed)
		log.Println("===========================================")

		if changed {
			err = Save(dir+packageFile, newPack)
			if err != nil {
				log.Printf("Can't save previous package.json. Error: %s", err)
			}
			os.Exit(10)
		}

		os.Exit(11)
	}

	fmt.Println("Put arguments: npm or pip. Example: ./isChangedLinux -npm=/package.json")
}
