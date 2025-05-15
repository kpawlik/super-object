package main

import (
	"encoding/csv"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"

	so "github.com/kpawlik/superobject"
)

var (
	csvPath      string
	databaseName string
	sql = 		`select feature_name, display_name, field_name from myw.dd_field_group fg 
		join myw.dd_field_group_item fgi on fg.id=fgi.container_id 
		where feature_name = '%s'  
		and field_name in (%s);`
)

func init() {
	flag.StringVar(&csvPath, "csv", "", "Path to the CSV file")
	flag.StringVar(&databaseName, "db", "iqgeo-test", "Name of the database")	
	flag.Parse()
	if csvPath == "" {
		panic("CSV path is required")
	}	

}

func main() {
	csvContent, err := os.ReadFile(csvPath)
	so.HandleErr(err)
	reader := strings.NewReader(string(csvContent))
	csvReader := csv.NewReader(reader)
	fieldsToRemove := make(map[string][]string)
	for{
		row, err  := csvReader.Read()
		if err == io.EOF {
			break
		}
		feature := row[0]
		if feature == "" {
			continue
		}
		field := row[1]
		removed := row[2]
		if removed != "removed" {
			continue
		}
		if _, ok := fieldsToRemove[feature]; !ok {
			fieldsToRemove[feature] = make([]string, 0)
		}
		fieldsToRemove[feature] = append(fieldsToRemove[feature], fmt.Sprintf("'%s'", field))
	} 
	for feature, fields := range fieldsToRemove {
		sqlStr := fmt.Sprintf(sql, feature, strings.Join(fields, ", "))
		fmt.Printf("echo %s\n", feature)
		fmt.Printf("echo \"%s\"\n", sqlStr)
		fmt.Printf("psql -d %s -c \"%s\"\n", databaseName, sqlStr)
		fmt.Printf("echo \"=======\"\n")
		fmt.Println()	

	}
}