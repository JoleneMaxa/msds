// Copyright 2019 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     https://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Command gce starts a web server.
package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"encoding/csv"
	"regexp"
	"strconv"
	"time"
)

type MSDSCourse struct {
	CID string `json:"courseI_D`
	CNAME string `json:"course_name"`
	CPREREQ string `json:"prerequisite"` 
	} 

var CSVFILE = "./data.csv"

type MSDSCourseCatalog []MSDSCourse 
var data = MSDSCourseCatalog{}
var index map[string]int

func readCSVFile(filepath string) error {
	_, err := os.Stat(filepath)
	if err != nil {
		return err
	}

	f, err := os.Open(filepath)
	if err != nil {
		return err
	}
	defer f.Close()

	// CSV file read all at once
	lines, err := csv.NewReader(f).ReadAll()
	if err != nil {
		return err
	}

	for _, line := range lines {
		// temp := Entry{
		// 	Name:       line[0],
		// 	Surname:    line[1],
		// 	Tel:        line[2],
		// 	LastAccess: line[3],
		// }
		temp := MSDSCourse{
			CID:       line[0],
			CNAME:    line[1],
			CPREREQ:        line[2],
		}
		// Storing to global variable
		data = append(data, temp)
	}

	return nil
}

func saveCSVFile(filepath string) error {
	csvfile, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer csvfile.Close()

	csvwriter := csv.NewWriter(csvfile)
	for _, row := range data {
		// temp := []string{row.Name, row.Surname, row.Tel}
		temp := []string{row.CID, row.CNAME, row.CPREREQ}
		_ = csvwriter.Write(temp)
	}
	csvwriter.Flush()
	return nil
}

func createIndex() error {
	index = make(map[string]int)
	for i, k := range data {
		//key := k.Tel
		key := k.CID
		index[key] = i
	}
	return nil
}

// Initialized by the user – returns a pointer
// If it returns nil, there was an error
func initS(I, N, P string) *MSDSCourse {
//func initS(N, S, T string) *MSDSCourse {
	// Both of them should have a value
	//if T == "" || S == "" {
	if P == "" || N == "" {
		return nil
	}
	// Give LastAccess a value
	//LastAccess := strconv.FormatInt(time.Now().Unix(), 10)
	CPREREQ := strconv.FormatInt(time.Now().Unix(), 10)
	return &MSDSCourse{CID: I, CNAME: N, CPREREQ: CPREREQ}
}

func insert(pS *MSDSCourse) error {
	// If it already exists, do not add it
	_, ok := index[(*pS).CID]
	if ok {
		return fmt.Errorf("%s already exists", pS.CID)
	}

	//*&pS.LastAccess = strconv.FormatInt(time.Now().Unix(), 10)
	*&pS.CPREREQ = strconv.FormatInt(time.Now().Unix(), 10)
	data = append(data, *pS)
	// Update the index
	_ = createIndex()

	err := saveCSVFile(CSVFILE)
	if err != nil {
		return err
	}
	return nil
}

func deleteEntry(key string) error {
	i, ok := index[key]
	if !ok {
		return fmt.Errorf("%s cannot be found!", key)
	}
	data = append(data[:i], data[i+1:]...)
	// Update the index - key does not exist any more
	delete(index, key)

	err := saveCSVFile(CSVFILE)
	if err != nil {
		return err
	}
	return nil
}

func search(key string) *MSDSCourse {
	i, ok := index[key]
	if !ok {
		return nil
	}
	data[i].CPREREQ = strconv.FormatInt(time.Now().Unix(), 10)
	return &data[i]
}

func matchTel(s string) bool {
	t := []byte(s)
	re := regexp.MustCompile(`\d+$`)
	return re.Match(t)
}

func list() string {
	var all string
	for _, k := range data {
		all = all + k.CID + " " + k.CNAME + " " + k.CPREREQ + "\n"
	}
	return all
}

func main() {
	http.HandleFunc("/", index)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	log.Printf("Listening on port %s", port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatal(err)
	}


	err := readCSVFile(CSVFILE)
	if err != nil {
		fmt.Println(err)
		return
	}

	err = createIndex()
	if err != nil {
		fmt.Println("Cannot create index.")
		return
	}

	mux := http.NewServeMux()
	s := &http.Server{
		Addr:         PORT,
		Handler:      mux,
		IdleTimeout:  10 * time.Second,
		ReadTimeout:  time.Second,
		WriteTimeout: time.Second,
	}

	mux.Handle("/list", http.HandlerFunc(listHandler))
	mux.Handle("/insert/", http.HandlerFunc(insertHandler))
	mux.Handle("/insert", http.HandlerFunc(insertHandler))
	mux.Handle("/search", http.HandlerFunc(searchHandler))
	mux.Handle("/search/", http.HandlerFunc(searchHandler))
	mux.Handle("/delete/", http.HandlerFunc(deleteHandler))
	mux.Handle("/status", http.HandlerFunc(statusHandler))
	mux.Handle("/", http.HandlerFunc(defaultHandler))

	fmt.Println("Ready to serve at", PORT)
	err = s.ListenAndServe()
	if err != nil {
		fmt.Println(err)
		return
}
}

func index(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Hello, World!")
}
