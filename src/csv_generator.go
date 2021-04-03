package src

import (
	"encoding/csv"
	"html"
	"log"
	"os"
	"strconv"
	"strings"
)

var csv_header = []string{"contributor", "created", "reviewed"}

func GenerateCsv(outpath string, counter *CommitCounter) {
	if _, err := os.Stat(csvOutputPath); os.IsNotExist(err) {
		os.Mkdir(csvOutputPath, 0755)
	}

	file, err := os.Create(csvOutputPath + "/result.csv")
	if err != nil {
		panic(err)
	}

	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	usedNameMap := map[string]int{}
	writer.Write(csv_header)

	for key, value := range counter.CommiterCounterMap {
		usedNameMap[key]++

		values := strings.Split(key, " ")
		email := html.UnescapeString(values[2])
		name := values[0] + " " + values[1] + " " + email

		writer.Write([]string{name, strconv.Itoa(value), strconv.Itoa(counter.ReviewerCounterMap[key])})
	}

	for key, value := range counter.ReviewerCounterMap {
		if _, ok := usedNameMap[key]; ok {
			continue
		}
		values := strings.Split(key, " ")
		email := html.UnescapeString(values[2])
		name := values[0] + " " + values[1] + " " + email

		writer.Write([]string{name, strconv.Itoa(counter.CommiterCounterMap[key]), strconv.Itoa(value)})
	}
}

func createAndWriteFile(file_name string, content string) {
	if _, err := os.Stat(outputPath); os.IsNotExist(err) {
		os.Mkdir(outputPath, 0755)
	}

	f, err := os.Create(outputPath + file_name)

	if err != nil {
		log.Fatal(err)
	}

	defer f.Close()

	_, err2 := f.WriteString(content)

	if err2 != nil {
		log.Fatal(err2)
	}
}
