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

func GenerateCsv(outpath string, counter *CommitCounter, csvOutputPath string) {
	if _, err := os.Stat(csvOutputPath); os.IsNotExist(err) {
		err := os.Mkdir(csvOutputPath, 0755)
		if err != nil {
			log.Fatal(err)
		}
	}

	file, err := os.Create(csvOutputPath + "/result.csv")
	if err != nil {
		log.Fatal(err)
	}

	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	usedNameMap := map[string]int{}
	err = writer.Write(csv_header)
	if err != nil {
		log.Fatal(err)
	}

	for key, value := range counter.CommiterCounterMap {
		usedNameMap[key]++

		values := strings.Split(key, " ")

		lastIndex := len(values) - 1
		email := html.UnescapeString(values[lastIndex])
		name := strings.Join(values[:lastIndex], " ") + " " + email

		err := writer.Write([]string{name, strconv.Itoa(value), strconv.Itoa(counter.ReviewerCounterMap[key])})
		if err != nil {
			log.Fatal(err)
		}
	}

	for key, value := range counter.ReviewerCounterMap {
		if _, ok := usedNameMap[key]; ok {
			continue
		}
		values := strings.Split(key, " ")
		lastIndex := len(values) - 1
		email := html.UnescapeString(values[lastIndex])
		name := strings.Join(values[:lastIndex], " ") + " " + email

		err := writer.Write([]string{name, strconv.Itoa(counter.CommiterCounterMap[key]), strconv.Itoa(value)})
		if err != nil {
			log.Fatal(err)
		}
	}
}

func createAndWriteFile(file_name string, content string, outputPath string) {
	if _, err := os.Stat(outputPath); os.IsNotExist(err) {
		err := os.Mkdir(outputPath, 0755)
		if err != nil {
			log.Fatal(err)
		}
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
