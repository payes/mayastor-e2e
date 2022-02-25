#!/usr/bin/env groovy

def GetDefaultJobParameters (k8sv, build_images, image_tag ) {
    return [
            [ $class: 'StringParameterValue',  name: "build_prefix",       value: "develop" ],
            [ $class: 'StringParameterValue',  name: "mayastorBranch",     value: "develop" ],
            [ $class: 'StringParameterValue',  name: "mcpBranch",          value: "develop" ],
            [ $class: 'StringParameterValue',  name: "xray_test_plan",     value: "MQ-2743" ],
            [ $class: 'BooleanParameterValue', name: "xray_send_report",   value: true ],
            [ $class: 'StringParameterValue',  name: "test_profile",       value: "regression" ],
            [ $class: 'StringParameterValue',  name: "parallelism",        value: "10" ],
            [ $class: 'StringParameterValue',   name: "kubernetes_version", value: k8sv ],
            [ $class: 'BooleanParameterValue',  name: "build_images",       value: build_images ],
            [ $class: 'StringParameterValue',   name: "image_tag",          value: image_tag ]
    ]
}

return this
