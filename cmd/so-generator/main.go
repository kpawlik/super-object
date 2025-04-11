package main

import (
	"bufio"
	"bytes"
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/kpawlik/om"
	so "github.com/kpawlik/superobject"
)


func init() {
	var(
		soSource string
		soCompose string
		soDest string
	)
	flag.StringVar(&soSource, "source", "", "Path to source superobject def file")
	flag.StringVar(&soCompose, "compose", "", "Path to compose superobject def file")
	flag.StringVar(&soDest, "dest", "", "Path to destination of superobject with combined fields. Output file will be created if it does not exist")
	flag.Parse()
	if soSource == "" || soCompose == "" || soDest == "" {
		flag.PrintDefaults()
		os.Exit(1)
	}

}	

func readDef(path string) (def *om.OrderedMap, err error){
	var file *os.File
	file, err = os.OpenFile(path, os.O_RDWR, 0644)
	if err != nil {
		return
	}
	defer file.Close()
	if def, err = so.ReadFeatureDef(bufio.NewReader(file)); err != nil {
		err = fmt.Errorf("failed to read feature definition from %s: %w", path, err)
		return	
	}
	return
}

func main() {
	var (
		err error
		source *om.OrderedMap
		compose *om.OrderedMap
		fieldsNames = []string{}
		file *os.File

	)
	sourcePath := flag.CommandLine.Lookup("source").Value.String()
	destPath := flag.CommandLine.Lookup("dest").Value.String()
	composePath := flag.CommandLine.Lookup("compose").Value.String()
	if source, err = readDef(sourcePath); err != nil {
		log.Fatal(err)
	}
	// add default group if it does not exist
	if !so.IsGroupExists(source, "Default"){
		defaultGroupFields := so.GetFields(source, so.GeomExcludedFields)
		defaultFields := make([]string, len(defaultGroupFields))
		for i, field := range defaultGroupFields {
			defaultFields[i] = field.Name
		}
		so.AddGroup(source, "Default", defaultFields)
	}
	// read compose definition
	if compose, err = readDef(composePath); err != nil {
		log.Fatal(err)
	}
	// buffer for methods body
	methods := bytes.NewBuffer([]byte{})
	fields := so.GetFields(compose, nil)
	// add fields to source superobject and generate methods
	for _, f := range fields {
		fieldName := f.Name
		featureName := f.FeatureName
		calcFieldName := fmt.Sprintf("calc__%s__%s", featureName, fieldName)
		if so.IsFieldExists(source, calcFieldName) {
			so.UpdateField(source, calcFieldName, f.ExternalName, f.Type)
			method := so.GetMethodBody(calcFieldName, featureName, fieldName)
			methods.WriteString(method)
		}else{
			so.AddField(source, calcFieldName, f.ExternalName, f.Type)
		}
		fieldsNames = append(fieldsNames, calcFieldName)
	}
	// add new fields group if needed
	composeExternalName := compose.Map["external_name"].(string)
	if so.IsGroupExists(source, composeExternalName) {
		so.UpdateGroup(source, composeExternalName, fieldsNames)
	}else{
		so.AddGroup(source, composeExternalName, fieldsNames)
	}
	// write new superobject definition to file
	file, err = os.OpenFile(destPath, os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		log.Fatalf("failed to open file %s: %v", destPath, err)	
	}
	defer file.Close()
	so.WriteFeatureDef(bufio.NewWriter(file), source)
	// write methods to file
	methodsPath := fmt.Sprintf("%s_methods.txt", destPath)
	file, err = os.OpenFile(methodsPath, os.O_APPEND|os.O_WRONLY, 0644)
	if os.IsNotExist(err){
		os.WriteFile(methodsPath, methods.Bytes(), 0644)
	}else{
		file.Write(methods.Bytes())
		file.Close()
	}
}