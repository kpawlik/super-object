package superobject

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"slices"
	"strings"
	"text/template"

	"github.com/kpawlik/om"
)

var (
	DefaultExcludedFields = []string{"reference_set", "reference", "linestring", "point", "polygon"}
	GeomExcludedFields    = []string{"linestring", "point", "polygon"}
	methodTemplateText    = `
	/**
	 Method for calculated field. 
	 @returns {any} value of field {{.FieldName}} from feature {{.FeatureName}}
	 */
    async {{.MethodName}}(){
        return await this.getSuperObjectFieldValue("{{.FeatureName}}", "{{.FieldName}}");
    }
`
	methodTemplate = template.New("methodTemplate")
)

type Method struct {
	MethodName  string
	FeatureName string
	FieldName   string
}

type Field struct {
	FeatureName  string
	Name         string
	ExternalName string
	Type         string
	Unit         string
}

func init() {
	methodTemplate = template.Must(methodTemplate.Parse(methodTemplateText))
}

// Return true if field already exists in the feature definition
// featureDef: the feature definition to add the field to
// fieldName: the name of the field to add
func IsFieldExists(featureDef *om.OrderedMap, fieldName string) bool {
	fields := featureDef.Map["fields"].([]any)
	for _, field := range fields {
		if field.(*om.OrderedMap).Map["name"] == fieldName {
			return true
		}
	}
	return false
}

// AddField adds a new field to the feature definition
// featureDef: the feature definition to add the field to
// fieldName: the name of the field to add
// externalName: the external name of the field to add
// fieldType: the type of the field to add
func AddField(featureDef *om.OrderedMap, fieldName string, externalName string, fieldType string, unit string) {
	field := om.NewOrderedMap()
	field.Set("name", fieldName)
	field.Set("external_name", externalName)
	field.Set("type", fieldType)
	field.Set("value", fmt.Sprintf("method(%s)", fieldName))
	if unit != "" {
		field.Set("unit", unit)
	}
	fields := featureDef.Map["fields"].([]any)
	fields = append(fields, field)
	featureDef.Set("fields", fields)
}

// UpdateField updates an existing field in the feature definition
// featureDef: the feature definition to add the field to
// fieldName: the name of the field to add
// externalName: the external name of the field to add
// fieldType: the type of the field to add
func UpdateField(featureDef *om.OrderedMap, fieldName string, externalName string, fieldType string, unit string) {
	fields := featureDef.Map["fields"].([]any)
	for i, iField := range fields {
		field := iField.(*om.OrderedMap)
		if field.Map["name"] == fieldName {
			field.Set("external_name", externalName)
			field.Set("type", fieldType)
			field.Set("value", fmt.Sprintf("method(%s)", fieldName))
			if unit != "" {
				field.Set("unit", unit)
			}
			fields[i] = field
		}
	}
	featureDef.Set("fields", fields)
}

// Check if group already exists in the feature definition
// featureDef: the feature definition to check
// groupName: the name of the group to check
func IsGroupExists(featureDef *om.OrderedMap, groupName string) bool {
	groups := featureDef.Map["groups"].([]any)
	for _, iGroup := range groups {
		if iGroup.(*om.OrderedMap).Map["name"] == groupName {
			return true
		}
	}
	return false
}

// AddGroup adds a new group to the feature definition
// featureDef: the feature definition to add the group to
// groupName: the name of the group to add
// fields: list of fields to add to group
func AddGroup(featureDef *om.OrderedMap, groupName string, fields []string) {
	groups := featureDef.Map["groups"].([]any)
	group := om.NewOrderedMap()
	group.Set("name", groupName)
	group.Set("visible", true)
	group.Set("expanded", false)
	group.Set("fields", fields)
	groups = append(groups, group)
	featureDef.Set("groups", groups)
}

// UpdateGroup
func UpdateGroup(featureDef *om.OrderedMap, groupName string, fields []string) {
	groups := featureDef.Map["groups"].([]any)
	for i, iGroup := range groups {
		group := iGroup.(*om.OrderedMap)
		if group.Map["name"] == groupName {
			group.Set("fields", fields)
			groups[i] = group
			break
		}
	}
	featureDef.Set("groups", groups)
}

// Get list of fields from feature definition. Exclude fields with prefix "myw_"
// and fields with type "reference_set", "reference", "linestring", "point", "polygon"
func GetFields(featureDef *om.OrderedMap, excluded []string) (fields []Field) {
	if excluded == nil {
		excluded = DefaultExcludedFields
	}
	featureName := featureDef.Map["name"].(string)
	fieldsDefs := featureDef.Map["fields"].([]any)
	for _, fieldDef := range fieldsDefs {
		field := fieldDef.(*om.OrderedMap)
		fieldName := field.Map["name"].(string)
		if strings.HasPrefix(fieldName, "myw_") {
			continue
		}
		externalName := field.Map["external_name"].(string)
		fieldType := field.Map["type"].(string)
		if slices.Contains(excluded, fieldType) {
			continue
		}
		unitValue := ""
		if unit := field.Map["unit"]; unit != nil {
			unitValue = unit.(string)
		}
		fields = append(fields, Field{
			FeatureName:  featureName,
			Name:         fieldName,
			ExternalName: externalName,
			Type:         fieldType,
			Unit:         unitValue,
		})
	}
	return fields
}

// Reads the feature definition from a file
func ReadFeatureDef(reader *bufio.Reader) (feature *om.OrderedMap, err error) {
	var (
		buff []byte
	)
	feature = om.NewOrderedMap()
	if buff, err = io.ReadAll(reader); err != nil {
		err = fmt.Errorf("failed to read feature definition: %w", err)
		return
	}

	if err = json.Unmarshal(buff, feature); err != nil {
		err = fmt.Errorf("failed to unmarshal feature definition: %w", err)
		return
	}
	return
}

// Writes the feature definition to a file
// The function replaces all occurrences of "\u0026" with "&" in the JSON output
func WriteFeatureDef(writer *bufio.Writer, feature *om.OrderedMap) (err error) {
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
	if _, err = writer.Write(res); err != nil {
		err = fmt.Errorf("failed to write feature definition: %w", err)
		return
	}
	return
}

// GetMethodBody generates the method body for a field in a feature definition
func GetMethodBody(methodName string, featureName string, fieldName string) (body string) {
	buff := bytes.NewBuffer([]byte{})
	methodTemplate.Execute(buff, Method{
		MethodName:  methodName,
		FeatureName: featureName,
		FieldName:   fieldName,
	})
	body = buff.String()
	return body

}
