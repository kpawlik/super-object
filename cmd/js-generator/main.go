package main

import (
	"bytes"
	"fmt"
	"os"
	"text/template"
)

type Relation struct {
	FeatureName string
	Relation    string
}

type Config struct {
	FileName string
	InternalName string
	ExternalName string
	Relations    []Relation
	Methods	  string
}

var jsTemplate = `
import StedSuperObjectFeature from "./stedSuperObjectFeature";

class StedSuperObject{{.ExternalName}} extends StedSuperObjectFeature  {
    static {
        this.prototype.so_configs = {
			{{range .Relations}}
            "{{.FeatureName}}": {
                "relation": "{{.Relation}}",
            },
			{{end}}
        }
    }
	{{.Methods}}
}

myw.StedSuperObject{{.ExternalName}} = StedSuperObject{{.ExternalName}};
export default StedSuperObject{{.ExternalName}};

`

var setDataModelTemplate = `
import {StedSuperObject{{.ExternalName}}} from "./stedSuperObject{{.ExternalName}}";
myw.featureModels["{{.InternalName}}"] = StedSuperObject{{.ExternalName}};
`


var configs = []Config{
	{
		FileName: "stedSuperObject%s.js",
		InternalName: "eo_connector_segment_inst",
		ExternalName: "Installatiegeleider",
		Relations: []Relation{
			{FeatureName: "eo_connector_segment", Relation: `["existing_assets", "future_assets", "past_assets"]`},
		},
	},
	{
		FileName: "stedSuperObject%s.js",
		InternalName: "eo_3w_power_xfrmr_inst",
		ExternalName: "3wTransformator",
		Relations: []Relation{
			{FeatureName: "eo_3w_power_xfrmr", Relation: `["existing_assets", "future_assets", "past_assets"]`},
			{FeatureName: "eo_3w_power_xfrmr_controller", Relation: `[]`},
		},
	},
	{
		FileName: "stedSuperObject%s.js",
		InternalName: "eo_power_xfrmr_inst",
		ExternalName: "Transformator",
		Relations: []Relation{
			{FeatureName: "eo_power_xfrmr", Relation: `["existing_assets", "future_assets", "past_assets"]`},
			{FeatureName: "eo_power_xfrmr_controller", Relation: `[]`},
		},
	},
	{
		FileName: "stedSuperObject%s.js",
		InternalName: "eo_measuring_eqpt_inst",
		ExternalName: "Meettransformator",
		Relations: []Relation{
			{FeatureName: "eo_measuring_eqpt", Relation: `["existing_assets", "future_assets", "past_assets"]`},
		},
	},
	{
		FileName: "stedSuperObject%s.js",
		InternalName: "eo_protective_eqpt_inst",
		ExternalName: "Beveiliging",
		Relations: []Relation{
			{FeatureName: "eo_protective_eqpt", Relation: `["existing_assets", "future_assets", "past_assets"]`},
		},
	},
	{
		FileName: "stedSuperObject%s.js",
		InternalName: "eo_isolating_eqpt_inst",
		ExternalName: "Schakelcomponent",
		Relations: []Relation{
			{FeatureName: "eo_isolating_eqpt", Relation: `["existing_assets", "future_assets", "past_assets"]`},
		},
	},
	{
		FileName: "stedSuperObject%s.js",
		InternalName: "eo_regulating_eqpt_inst",
		ExternalName: "Energieregeling",
		Relations: []Relation{
			{FeatureName: "eo_regulating_eqpt", Relation: `["existing_assets", "future_assets", "past_assets"]`},
		},
	},

}

func main() {
	
	methodTemplate := template.Must(template.New("methodTemplate").Parse(jsTemplate))
	dmTemplate := template.Must(template.New("methodTemplate").Parse(setDataModelTemplate))
	for _, config := range configs{
		
		buff := bytes.NewBuffer([]byte{})
		methodTemplate.Execute(buff, config)
		body := buff.String()
		os.WriteFile(fmt.Sprintf(config.FileName, config.ExternalName), []byte(body), 0644)

	}
	dmbuff := bytes.NewBuffer([]byte{})
	for _, config := range configs{
		
		buff := bytes.NewBuffer([]byte{})	
		dmTemplate.Execute(buff, config)
		dmbuff.Write(buff.Bytes())

	}
	os.WriteFile("setDM.js", dmbuff.Bytes(), 0644)
	

}