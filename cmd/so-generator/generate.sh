TEMPDIR=/shared-data/temp/
GENPATH=/shared-data/bin/super-object/cmd/so-generator


function generate(){
    local main_feature_name=$1
    shift            # Shift all arguments to the left (original $1 gets lost)
    local components_features=("$@") # Rebuild the array with rest of arguments
    local result_file="${main_feature_name}_result.def"
    local features=( $main_feature_name "${components_features[@]}")
    for feature in "${features[@]}"
    do 
        myw_db $MYW_DB_NAME dump $TEMPDIR features $feature
    done
    # copy main definition
    cp "${TEMPDIR}/${main_feature_name}.def" "${TEMPDIR}/${result_file}"
    # add fields from components
    for component in "${components_features[@]}"
    do 
        $GENPATH/so-generator -source "$TEMPDIR/${result_file}" -compose "$TEMPDIR/${component}.def" -dest $TEMPDIR/$result_file
    done
}


## Kabel
#features=(eo_cable_segment_inst eo_cable eo_cable_exi_phase)
main_feature=eo_cable_segment_inst
components=(eo_cable eo_cable_exi_phase)
generate $main_feature "${components[@]}"

## Wire
main_feature=eo_wire_segment_inst
components=(eo_wire eo_wire_exi_phase)
generate $main_feature "${components[@]}"

## Koppeling
main_feature=eo_connector_point_inst
components=(eo_connector_point eo_connector_point_exi_phase)
generate $main_feature "${components[@]}"

## Aansluiting
main_feature=eo_service_point
components=(eo_service_connection)
generate $main_feature "${components[@]}"

## Station
main_feature=sub_substation
components=(sub_substation_boundary)
generate $main_feature "${components[@]}"

## Kast
main_feature=ed_cabinet
components=(stedin_cabinet_spec)
generate $main_feature "${components[@]}"

## Mast
main_feature=ed_pole
components=(ed_cross_arm ed_insulator ed_riser)
generate $main_feature "${components[@]}"