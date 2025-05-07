WORKDIR=/shared-data/temp/so-20250506
TEMPDIR=$WORKDIR/temp
OUTDIR=$WORKDIR/out
GENPATH=/shared-data/bin/super-object/cmd/so-generator
MYW_DB_NAME=iqgeo-test

function generate(){
    
    local main_feature_name=$1
    shift
    local components_features=("$@") # Rebuild the array with rest of arguments
    local result_file="${main_feature_name}.def"
    local features=( $main_feature_name "${components_features[@]}")

    
    for feature in "${features[@]}"
    do 
        myw_db $MYW_DB_NAME dump $TEMPDIR features $feature
    done
    # copy main definition
    cp "${TEMPDIR}/${main_feature_name}.def" "${OUTDIR}/${result_file}"
    # add fields from components
    for component in "${components_features[@]}"
    do 
        $GENPATH/so-generator -source "$OUTDIR/${result_file}" -compose "$TEMPDIR/${component}.def" -dest $OUTDIR/$result_file
    done
}

mkdir -p $TEMPDIR
mkdir -p $OUTDIR



## Installatiegeleider
main_feature=eo_connector_segment_inst
components=(eo_connector_segment)
generate $main_feature "${components[@]}"

# Schakelinstallatie
main_feature=eo_composite_switch
components=(eo_composite_switch_spec eo_building)
generate $main_feature "${components[@]}"


## 3wikkelingTransformator
main_feature=eo_3w_power_xfrmr_inst
components=(eo_3w_power_xfrmr eo_3w_power_xfrmr_controller)
generate $main_feature "${components[@]}"


## Transformator
main_feature=eo_power_xfrmr_inst
components=(eo_power_xfrmr eo_power_xfrmr_controller)
generate $main_feature "${components[@]}"


## Meettransformator
main_feature=eo_measuring_eqpt_inst
components=(eo_measuring_eqpt)
generate $main_feature "${components[@]}"


## Beveiliging
main_feature=eo_protective_eqpt_inst
components=(eo_protective_eqpt)
generate $main_feature "${components[@]}"


## Schakelcomponent
main_feature=eo_isolating_eqpt_inst
components=(eo_isolating_eqpt eo_isolating_eqpt_controller)
generate $main_feature "${components[@]}"


## Energieregeling
main_feature=eo_regulating_eqpt_inst
components=(eo_regulating_eqpt)
generate $main_feature "${components[@]}"