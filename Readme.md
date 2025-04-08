# Generate fields and methods for Super Objects

## Usage 

```bash
go run cmd/so-generator/main.go 
  -compose string
        Path to compose superobject def file
  -dest string
        Path to destination of superobject with combined fields. Output file will be created if it does not exist
  -source string
        Path to source superobject def file
```

## How it works

- read super object definition `source`
- read compose object definition `compose`
- get all stored fields (except: myw_*, geometry, relations) from `compose`
- append fields from `compose` to `source`
- generate calc methods body for each field
- store result definition file as `dest` file
- store methods as `dest_methods.txt` file

## Example usage

```bash
# init source + component
go run main.go -source $DEFS/eo_connector_point_inst.def -compose $DEFS/eo_cable.def -dest $DEFS/eo_connector_point_inst_res.def
# result + second component
go run main.go -source $DEFS/eo_connector_point_inst_res.def -compose $DEFS/eo_cable_exi_phase.def -dest $DEFS/eo_connector_point_inst_res.def

```
