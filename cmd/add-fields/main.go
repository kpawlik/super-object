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
	SourceDir string
	FieldsDiffFile string
)


func init() {
	flag.StringVar(&FeatureDir, "target-dir", "", "Feature dir")
	flag.StringVar(&SourceDir, "source-dir", "", "Source dir")
	flag.StringVar(&FieldsDiffFile, "cmp-file", "", "Fields diff file")	
	flag.Parse()
}

func main() {
	csvContent, err := os.ReadFile(FieldsDiffFile)
	so.HandleErr(err)
	reader := strings.NewReader(string(csvContent))
	csvReader := csv.NewReader(reader)
	fieldsToAdd := make(map[string][]string)
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
		removed := row[3]
		if removed != "added" {
			continue
		}
		if _, ok := fieldsToAdd[feature]; !ok {
			fieldsToAdd[feature] = make([]string, 0)
		}
		fieldsToAdd[feature] = append(fieldsToAdd[feature], field)
	} 
	for feature, fields := range fieldsToAdd {
		featurePath := filepath.Join(FeatureDir, feature + ".def")
		sourcePath := filepath.Join(SourceDir, feature + ".def")
		addFields(featurePath, sourcePath, fields)

	}
	
}

func addFields(featurePath string, sourcePath string, fieldsToAdd []string) {
	var (
		fileContent []byte
		err         error
	)
	fileContent, err = os.ReadFile(featurePath)
	so.HandleErr(err)
	featureOm := om.NewOrderedMap()
	err = featureOm.UnmarshalJSON(fileContent)
	so.HandleErr(err)
	fileContent, err = os.ReadFile(sourcePath)
	so.HandleErr(err)
	sourceOm := om.NewOrderedMap()
	err = sourceOm.UnmarshalJSON(fileContent)
	so.HandleErr(err)
	featureFields:= featureOm.Map["fields"].([]any)
	sourceFields := sourceOm.Map["fields"].([]any)
	for _, sourceField := range sourceFields {
		fieldName := sourceField.(*om.OrderedMap).Map["name"].(string)
		if slices.Contains(fieldsToAdd, fieldName) {
			featureFields = append(featureFields, sourceField)
		}
		
	}
	featureOm.Map["fields"] = featureFields
	fileContent, err = featureOm.MarshalIndent("    ")
	so.HandleErr(err)
	err = os.WriteFile(featurePath, fileContent, 0644)
	so.HandleErr(err)

	
}