package main

import (
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

func main() {
	var (
		err error
		source *om.OrderedMap
		compose *om.OrderedMap
		fieldsNames = []string{}

	)
	sourcePath := flag.CommandLine.Lookup("source").Value.String()
	destPath := flag.CommandLine.Lookup("dest").Value.String()
	composePath := flag.CommandLine.Lookup("compose").Value.String()
	if source, err = so.ReadFeatureDef(sourcePath); err != nil {
		log.Fatal(err)
	}
	so.AddDefaultGroup(source)
	if compose, err = so.ReadFeatureDef(composePath); err != nil {
		log.Fatal(err)
	}
	methodsPath := fmt.Sprintf("%s_methods.txt", destPath)
	methods := bytes.NewBuffer([]byte{})
	fields := so.GetFields(compose)
	composeExternalName := compose.Map["external_name"].(string)
	for _, f := range fields {
		fieldName := f["name"]
		featureName := f["feature_name"]
		calcFieldName := fmt.Sprintf("calc__%s__%s", featureName, fieldName)
		so.AddField(source, calcFieldName, f["external_name"], f["type"])
		method := so.GetMethodBody(calcFieldName, featureName, fieldName)
		methods.WriteString(method)
		fieldsNames = append(fieldsNames, calcFieldName)
	}
	so.AddGroup(source, composeExternalName, fieldsNames)
	so.WriteFeatureDef(destPath, source)
	file, err := os.OpenFile(methodsPath, os.O_APPEND|os.O_WRONLY, 0644)
	if os.IsNotExist(err){
		os.WriteFile(methodsPath, methods.Bytes(), 0644)
	}else{
		file.Write(methods.Bytes())
		file.Close()
	}
	log.Printf("Methods written to %s", methodsPath)
	log.Printf("Superobject written to %s", destPath)
}