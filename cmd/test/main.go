package main

import (
	"bytes"
	"fmt"
	"os"

	so "github.com/kpawlik/superobject"
)

func main() {
	soName := "eo_connector_point_inst"
	sObj,err := so.ReadFeatureDef(fmt.Sprintf("test/%s.def", soName))
	if err != nil {
		panic(err)
	}
	composeName := "eo_cable"
	co, err:= so.ReadFeatureDef(fmt.Sprintf("test/%s.def", composeName))
	if err != nil {	
		panic(err)
	}
	methods := bytes.NewBuffer([]byte{})
	fields := so.GetFields(co)
	for _, f := range fields {
		calcFieldName := fmt.Sprintf("calc__%s__%s", f["feature_name"], f["name"])
		so.AddField(sObj, calcFieldName, f["external_name"], f["type"])
		method := so.GetMethodBody(calcFieldName, f["feature_name"], f["name"])
		methods.WriteString(method)
	}
	so.WriteFeatureDef(fmt.Sprintf("test/%s_result.def", soName), sObj)
	methodsFile := fmt.Sprintf("test/%s_methods.txt", soName)
	file, err := os.OpenFile(methodsFile, os.O_APPEND|os.O_WRONLY, 0644)
	if os.IsNotExist(err){
		os.WriteFile(fmt.Sprintf("test/%s_methods.txt", soName), methods.Bytes(), 0644)
	}else{
		file.Write(methods.Bytes())
		file.Close()
	}
	
}
