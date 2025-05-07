package main

import (
	"encoding/csv"
	"encoding/json"
	"flag"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	so "github.com/kpawlik/superobject"
)

var (
	//fieldsToCheck = []string{}
	fieldsToCheck = []string{"type", "external_name"}
	ignoreFields  = []string{}
)

type ResultBothWay struct {
	featureName string
	fieldName string
	d1Fields  map[string]string
	d2Fields  map[string]string
	stateInD1 string
	stateInD2 string
}

type FeatureDef map[string]any

type Exporter struct {
	writer *csv.Writer
}

func (e *Exporter) WriteHeader() {
	row := []string{}

	dir1Name := flag.CommandLine.Lookup("name1").Value.String()
	dir2Name := flag.CommandLine.Lookup("name2").Value.String()
	row = append(row, "Feature")
	row = append(row, "Field")
	row = append(row, dir1Name)
	row = append(row, dir2Name)
	for _, field := range fieldsToCheck {
		row = append(row, fmt.Sprintf("%s (%s)", field, dir1Name))
		row = append(row, fmt.Sprintf("%s (%s)", field, dir2Name))
	}
	e.writer.Write(row)
	e.writer.Flush()
}

func (e *Exporter) WriteSeparator() {
	rowLength := 4 + len(fieldsToCheck)*2
	row := make([]string, rowLength)
	e.writer.Write(row)
	e.writer.Flush()
}

func (e *Exporter) WriteRow(result *ResultBothWay) {
	row := []string{result.featureName, result.fieldName, result.stateInD1, result.stateInD2}
	for _, fieldName := range fieldsToCheck {
		row = append(row, result.d1Fields[fieldName])
		row = append(row, result.d2Fields[fieldName])
	}
	e.writer.Write(row)
	e.writer.Flush()
}


func init() {
	var (
		dir1, dir2, dir1Name, dir2Name string
		displayNonExists               bool
		fieldsToCheckStr, ignoreFieldsStr string
	)
	flag.StringVar(&dir1, "dir1", "", "Dir 1")
	flag.StringVar(&dir2, "dir2", "", "Dir 2")
	flag.StringVar(&dir1Name, "name1", "Dir 1", "Dir 1 Name")
	flag.StringVar(&dir2Name, "name2", "Dir 2", "Dir 2 Name")
	flag.StringVar(&fieldsToCheckStr, "fields-to-check", "" , "Fields to check")
	flag.StringVar(&ignoreFieldsStr, "ignore-fields", "" , "Ignore fields")
	if fieldsToCheckStr != "" {
		fieldsToCheck = strings.Split(fieldsToCheckStr, ",")
	}
	if ignoreFieldsStr != "" {
		ignoreFields = strings.Split(ignoreFieldsStr, ",")
	}
	flag.BoolVar(&displayNonExists, "not-exists", false, "display not existing files")
	flag.Parse()
}



func main() {
	dir1 := flag.CommandLine.Lookup("dir1").Value.String()
	dir2 := flag.CommandLine.Lookup("dir2").Value.String()
	displayNonExists := flag.CommandLine.Lookup("not-exists").Value.String() == "true"
	compareBothWay(dir1, dir2, displayNonExists)

}

func compareBothWay(dir1, dir2 string, displayNonExists bool) {
	var (
		entries1 []fs.DirEntry
		entry1   fs.DirEntry
		err      error
		bytes    []byte
		feature1 FeatureDef
		feature2 FeatureDef
	)
	entries1, err = os.ReadDir(dir1)
	so.HandleErr(err)
	exporter := &Exporter{writer: csv.NewWriter(os.Stdout)}
	exporter.WriteHeader()
	for _, entry1 = range entries1 {
		if entry1.IsDir() {
			continue
		}
		fileName := entry1.Name()
		if !strings.HasSuffix(fileName, ".def") {
			continue
		}
		filepath1 := filepath.Join(dir1, fileName)
		filepath2 := filepath.Join(dir2, fileName)
		bytes, err = os.ReadFile(filepath1)
		so.HandleErr(err)
		err = json.Unmarshal(bytes, &feature1)
		if err != nil{
			err = fmt.Errorf("error unmarshal file %s [%w]", filepath1, err)
			so.HandleErr(err)
		}
		
		bytes, err = os.ReadFile(filepath2)
		if os.IsNotExist(err) {
			if displayNonExists {
				fmt.Printf("File %s not exists\n", filepath2)
			}
			continue
		}
		so.HandleErr(err)
		so.HandleErr(json.Unmarshal(bytes, &feature2))
		res := compareFieldsBothWay(feature1, feature2)
		if len(res) > 0 {
			csvExportBothWay(exporter, res)
			exporter.WriteSeparator()
		}
	}
}

func fieldNames(fields []any) (results []string) {
	for _, intFields1 := range fields {
		field1, _ := intFields1.(map[string]any)
		fieldName1 := field1["name"].(string)
		results = append(results, fieldName1)
	}
	return
}

func getFieldDef(fields []any, fieldName string) map[string]any {
	for _, intFields1 := range fields {
		field1, _ := intFields1.(map[string]any)
		fieldName1 := field1["name"].(string)
		if fieldName1 == fieldName {
			return field1
		}
	}
	return nil
}

func compareFieldsBothWay(feature1 FeatureDef, feature2 FeatureDef) (results []*ResultBothWay) {
	allFields := map[string]*ResultBothWay{}
	fields1 := feature1["fields"]
	fields2 := feature2["fields"]
	mFields1, _ := fields1.([]any)
	mFields2, _ := fields2.([]any)
	fields := fieldNames(mFields1)
	featureName := feature1["name"].(string)
	fields = append(fields, fieldNames(mFields2)...)

	for _, fieldName := range fields {
		ignore := false
		for _, ignoreField := range ignoreFields{
			if strings.HasPrefix(fieldName, ignoreField) {
				ignore = true
				break
			}
		}
		if ignore{
			continue
		}

		allFields[fieldName] = &ResultBothWay{
			fieldName: fieldName,
			featureName: featureName,
			d1Fields:  map[string]string{},
			d2Fields: map[string]string{},
		}
	}

	for fieldName, result := range allFields {
		field1 := getFieldDef(mFields1, fieldName)
		field2 := getFieldDef(mFields2, fieldName)
		if field1 == nil && field2 != nil {
			result.stateInD2 = "added"
			results = append(results, result)
			continue
		}
		if field1 != nil && field2 == nil {
			result.stateInD1 = "removed"
			results = append(results, result)
			continue
		}
		
		different := false
		for _, fieldName := range fieldsToCheck {
			if field1[fieldName] != nil && 
				field2[fieldName] != nil && 
				field1[fieldName].(string) != field2[fieldName].(string) {
				result.d1Fields[fieldName] = field1[fieldName].(string)
				result.d2Fields[fieldName] = field2[fieldName].(string)
				different = true
			}
		}
		if different {
			results = append(results, result)
		}
	}
	return
}

func csvExportBothWay(exporter *Exporter, results []*ResultBothWay) {
	for _, result := range results {
		if result.stateInD1 != "" {
			exporter.WriteRow(result)
		}
	}
	for _, result := range results {
		if result.stateInD2 != "" {
			exporter.WriteRow(result)
		}
	}
	for _, result := range results {
		if result.stateInD1 != "" || result.stateInD2 != "" {
			continue
		}
		exporter.WriteRow(result)
	}
}
