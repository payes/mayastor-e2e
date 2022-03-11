#!/usr/bin/env groovy

// params must have the following fields
//  xray_test_plan  : string
//  build_images    : boolean
//  image_tag       : string
//  datacore_bolt   : boolean
def GetDefaultJobParameters (k8sv, params) {
    return [
            [ $class: 'StringParameterValue',  name: "build_prefix",       value: "develop" ],
            [ $class: 'StringParameterValue',  name: "dataplaneBranch",    value: "develop" ],
            [ $class: 'StringParameterValue',  name: "controlplaneBranch", value: "develop" ],
            [ $class: 'StringParameterValue',  name: "xray_test_plan",     value: params.xray_test_plan ],
            [ $class: 'BooleanParameterValue', name: "xray_send_report",   value: true ],
            [ $class: 'StringParameterValue',  name: "test_profile",       value: "regression" ],
            [ $class: 'StringParameterValue',  name: "parallelism",        value: "10" ],
            [ $class: 'StringParameterValue',  name: "kubernetes_version", value: k8sv ],
            [ $class: 'BooleanParameterValue', name: "build_images",       value: params.build_images ],
            [ $class: 'StringParameterValue',  name: "image_tag",          value: params.image_tag ],
            [ $class: 'BooleanParameterValue', name: "datacore_bolt",      value: params.datacore_bolt ],
    ]
}

return this
