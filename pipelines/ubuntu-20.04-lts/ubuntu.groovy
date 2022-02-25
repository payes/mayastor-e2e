#!/usr/bin/env groovy

def GetDefaultJobParameters () {
    return [
            [ $class: 'StringParameterValue',  name: "build_prefix",       value: "develop" ],
            [ $class: 'StringParameterValue',  name: "mayastorBranch",     value: "develop" ],
            [ $class: 'StringParameterValue',  name: "mcpBranch",          value: "develop" ],
            [ $class: 'StringParameterValue',  name: "xray_test_plan",     value: "MQ-2743" ],
            [ $class: 'BooleanParameterValue', name: "xray_send_report",   value: true ],
            [ $class: 'StringParameterValue',  name: "test_profile",       value: "regression" ]
            [ $class: 'StringParameterValue',  name: "parallelism",        value: "10" ]
    ]
}

return this
