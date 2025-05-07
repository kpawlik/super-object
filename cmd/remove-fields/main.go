package main

import (
	"encoding/csv"
	"flag"
	"io"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/kpawlik/om"
	so "github.com/kpawlik/superobject"
)

var(
	FeatureDir string
	FieldsDiffFile string
)


func init() {
	flag.StringVar(&FeatureDir, "target-dir", "", "Feature dir")
	flag.StringVar(&FieldsDiffFile, "cmp-file", "", "Fields diff file")	
	flag.Parse()
}

func main() {
	csvContent, err := os.ReadFile(FieldsDiffFile)
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
		fieldsToRemove[feature] = append(fieldsToRemove[feature], field)
	} 
	for feature, fields := range fieldsToRemove {
		featurePath := filepath.Join(FeatureDir, feature + ".def")
		removeFields(featurePath, fields)

	}
	
}

func removeFields(featurePath string, fieldsToRemove []string) {
	var (
		fileContent []byte
		err         error
	)
	fileContent, err = os.ReadFile(featurePath)
	so.HandleErr(err)
	oMap := om.NewOrderedMap()
	err = oMap.UnmarshalJSON(fileContent)
	so.HandleErr(err)
	featureFields:= oMap.Map["fields"].([]any)
	newFields := make([]any, 0)
	for _, featureField := range featureFields {
		fieldName := featureField.(*om.OrderedMap).Map["name"].(string)
		if slices.Contains(fieldsToRemove, fieldName) {
			continue
		}
		newFields = append(newFields, featureField)
	}
	oMap.Map["fields"] = newFields
	fileContent, err = oMap.MarshalIndent("    ")
	so.HandleErr(err)
	err = os.WriteFile(featurePath, fileContent, 0644)
	so.HandleErr(err)

	
}