package superobject

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"slices"
	"strings"

	"github.com/kpawlik/om"
)

// AddField adds a new field to the feature definition
// featureDef: the feature definition to add the field to
// fieldName: the name of the field to add
// externalName: the external name of the field to add
// fieldType: the type of the field to add
func AddField(featureDef *om.OrderedMap, fieldName string, externalName string, fieldType string) {
	field := om.NewOrderedMap()
	field.Set("name", fieldName)
	field.Set("external_name", externalName)
	field.Set("type", fieldType)
	field.Set("value", fmt.Sprintf("method(%s)", fieldName))
	fields := featureDef.Map["fields"].([]any)
	fields = append(fields, field)
	featureDef.Set("fields", fields)
}

// Get list of fields from feature definition. Exclude fields with prefix "myw_"
// and fields with type "reference_set", "reference", "linestring", "point", "polygon"
func GetFields(featureDef *om.OrderedMap) (fields []map[string]string) {
	featureName := featureDef.Map["name"].(string)
	fieldsDefs := featureDef.Map["fields"].([]any)
	fields = make([]map[string]string, 0)
	for _, fieldDef := range fieldsDefs {
		field := fieldDef.(*om.OrderedMap)
		fieldName := field.Map["name"].(string)
		if strings.HasPrefix(fieldName, "myw_") {
			continue
		}
		externalName := field.Map["external_name"].(string)
		fieldType := field.Map["type"].(string)
		if slices.Contains([]string{"reference_set", "reference", "linestring", "point", "polygon"}, fieldType) {
			continue
		}
		fields = append(fields, map[string]string{
			"feature_name":  featureName,
			"name":          fieldName,
			"external_name": externalName,	
			"type":          fieldType,
		})
	}
	return fields
}

// Reads the feature definition from a file
func ReadFeatureDef(path string) (feature *om.OrderedMap, err error){
	var (
		buff []byte
	)
	feature = om.NewOrderedMap()
	if buff, err = os.ReadFile(path); err != nil {
		return
	}
	if err = json.Unmarshal(buff, feature); err != nil {
		return
	}
	return 
}

// Writes the feature definition to a file
// The function replaces all occurrences of "\u0026" with "&" in the JSON output
func WriteFeatureDef(path string, feature *om.OrderedMap) (err error){
	var (
		buf bytes.Buffer
		e   *json.Encoder
	)
	if e = json.NewEncoder(&buf); e == nil {
		return
	}
	e.SetIndent("", "  ")
	e.SetEscapeHTML(false)
	if err = e.Encode(feature); err != nil {
		return
	}
	res := bytes.ReplaceAll(buf.Bytes(), []byte("\\u0026"), []byte("&"))
	if err = os.WriteFile(path, res, 0644); err != nil {
		return
	}
	return 
}

// GetMethodBody generates the method body for a field in a feature definition
func GetMethodBody(methodName string, featureName string, fieldName string) (body string) {
	body = fmt.Sprintf(`
    async %s(){
        return await this.getSuperObjectFieldValue("%s", "%s");
    }
`, methodName, featureName, fieldName)
	return body

}
