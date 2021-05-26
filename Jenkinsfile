#!/usr/bin/env groovy

// On-demand E2E infra configuration
// https://mayadata.atlassian.net/wiki/spaces/MS/pages/247332965/Test+infrastructure#On-Demand-E2E-K8S-Clusters

def e2e_build_cluster_job='k8s-build-cluster' // Jenkins job to build cluster
def e2e_destroy_cluster_job='k8s-destroy-cluster' // Jenkins job to destroy cluster
// Environment to run e2e test in (job param of $e2e_build_cluster_job)
def e2e_environment="hcloud-kubeadm"
// Global variable to pass current k8s job between stages
def k8s_job=""

e2e_image_tag='nightly'
e2e_local_registry=true
e2e_reports_dir='artifacts/reports/'
e2e_test_profile = 'selfci'

run_linter=true

xray_send_report=true
xray_projectkey='MQ'
xray_self_ci_testplan='MQ-482'
xray_test_execution_type='10059'

// In the case of multi-branch pipelines, the pipeline
// name a.k.a. job base name, will be the
// 2nd-to-last item of env.JOB_NAME which
// consists of identifiers separated by '/' e.g.
//     first/second/pipeline/branch
// In the case of a non-multibranch pipeline, the pipeline
// name is env.JOB_NAME. This caters for all eventualities.
def getJobBaseName() {
  def jobSections = env.JOB_NAME.tokenize('/') as String[]
  return jobSections.length < 2 ? env.JOB_NAME : jobSections[ jobSections.length - 2 ]
}

def getTag() {
  return e2e_image_tag
}

def getTestPlan() {
  return xray_self_ci_testplan
}

// Install Loki on the cluster
def lokiInstall(tag, loki_run_id) {
  sh 'kubectl apply -f ./loki/promtail_namespace_e2e.yaml'
  sh 'kubectl apply -f ./loki/promtail_rbac_e2e.yaml'
  sh 'kubectl apply -f ./loki/promtail_configmap_e2e.yaml'
  def cmd = "run=\"${loki_run_id}\" version=\"${tag}\" envsubst -no-unset < ./loki/promtail_daemonset_e2e.template.yaml | kubectl apply -f -"
  sh "nix-shell --run '${cmd}'"
}

// Unnstall Loki
def lokiUninstall(tag, loki_run_id) {
  def cmd = "run=\"${loki_run_id}\" version=\"${tag}\" envsubst -no-unset < ./loki/promtail_daemonset_e2e.template.yaml | kubectl delete -f -"
  sh "nix-shell --run '${cmd}'"
  sh 'kubectl delete -f ./loki/promtail_configmap_e2e.yaml'
  sh 'kubectl delete -f ./loki/promtail_rbac_e2e.yaml'
  sh 'kubectl delete -f ./loki/promtail_namespace_e2e.yaml'
}

String job_base_name = getJobBaseName()
String loki_run_id = job_base_name + "-" + env.BRANCH_NAME + "-" + env.BUILD_NUMBER

pipeline {
  agent none
  options {
    timeout(time: 3, unit: 'HOURS')
  }

  stages {
    stage('init') {
      agent { label 'nixos-mayastor' }
      steps {
        step([
          $class: 'GitHubSetCommitStatusBuilder',
          contextSource: [
            $class: 'ManuallyEnteredCommitContextSource',
            context: 'continuous-integration/jenkins/branch'
          ],
          statusMessage: [ content: 'Pipeline started' ]
        ])
      }
    }
    stage('linter') {
      agent { label 'nixos-mayastor' }
      when {
        beforeAgent true
        expression { run_linter == true }
      }
      steps {
        sh 'nix-shell --run "./scripts/go-checks.sh -n"'
      }
    }
    stage('test') {
      stages {
        stage('build e2e cluster') {
          agent { label 'nixos' }
          steps {
            script {
              k8s_job=build(
                job: "${e2e_build_cluster_job}",
                propagate: true,
                wait: true,
                parameters: [[
                  $class: 'StringParameterValue',
                  name: "ENVIRONMENT",
                  value: "${e2e_environment}"
                ]]
              )
            }
          }
        }
        stage('run e2e') {
          agent { label 'nixos' }
          environment {
            KUBECONFIG = "${env.WORKSPACE}/${e2e_environment}/modules/k8s/secrets/admin.conf"
          }
          steps {
            // FIXME(arne-rusek): move hcloud's config to top-level dir in TF scripts
            sh """
              mkdir -p "${e2e_environment}/modules/k8s/secrets"
            """
            copyArtifacts(
              projectName: "${k8s_job.getProjectName()}",
              selector: specific("${k8s_job.getNumber()}"),
              filter: "${e2e_environment}/modules/k8s/secrets/admin.conf",
              target: "",
              fingerprintArtifacts: true
            )
            sh 'kubectl get nodes -o wide'

            script {
              sh "mkdir -p ./${e2e_reports_dir}"
              sh 'rm -Rf Mayastor'
              // we need the Mayastor repo, but only for the deployment yamls
              withCredentials([
                usernamePassword(credentialsId: 'github-checkout', usernameVariable: 'ghuser', passwordVariable: 'ghpw')
              ]) {
                sh "git clone https://github.com/openebs/Mayastor.git"
                sh 'cd Mayastor && git checkout develop'
              }

              def tag = getTag()
              def cmd = "./scripts/e2e-test.sh --device /dev/sdb --tag \"${tag}\" --logs --profile \"${e2e_test_profile}\" --loki_run_id \"${loki_run_id}\" --mayastor \"${env.WORKSPACE}/Mayastor\" --reportsdir \"${env.WORKSPACE}/${e2e_reports_dir}\" "
              // building images also means using the CI registry
              if (e2e_local_registry == true) {
                cmd = cmd + " --registry \"" + env.REGISTRY + "\""
              }

              withCredentials([
                usernamePassword(credentialsId: 'GRAFANA_API', usernameVariable: 'grafana_api_user', passwordVariable: 'grafana_api_pw'),
                string(credentialsId: 'HCLOUD_TOKEN', variable: 'HCLOUD_TOKEN')
              ]) {
                lokiInstall(tag, loki_run_id)
                sh "nix-shell --run '${cmd}'"
                lokiUninstall(tag, loki_run_id) // so that, if we keep the cluster, the next Loki instance can use different parameters
              }
            }
          }
          post {
            always {
              archiveArtifacts 'artifacts/**/*.*'
              // handle junit results on success or failure
              junit "${e2e_reports_dir}/*.xml"
              script {
                // destroy the cluster
                build(
                  job: "${e2e_destroy_cluster_job}",
                  propagate: true,
                  wait: true,
                  parameters: [
                    [
                      $class: 'RunParameterValue',
                      name: "BUILD",
                      runId:"${k8s_job.getProjectName()}#${k8s_job.getNumber()}"
                    ]
                  ]
                )
                try {
                  if (xray_send_report == true) {
                    def xray_testplan = getTestPlan()
                    def tag = getTag()
                    step([
                      $class: 'XrayImportBuilder',
                      endpointName: '/junit/multipart',
                      importFilePath: "${e2e_reports_dir}/*.xml",
                      importToSameExecution: 'true',
                      projectKey: "${xray_projectkey}",
                      testPlanKey: "${xray_testplan}",
                      serverInstance: "${env.JIRASERVERUUID}",
                      inputInfoSwitcher: 'fileContent',
                      importInfo: """{
                        "fields": {
                          "summary": "Build #${env.BUILD_NUMBER}, branch: ${env.BRANCH_name}, tag: ${tag}",
                          "project": {
                            "key": "${xray_projectkey}"
                          },
                          "issuetype": {
                            "id": "${xray_test_execution_type}"
                          },
                          "description": "Results for build #${env.BUILD_NUMBER} at ${env.BUILD_URL}"
                        }
                      }"""
                    ])
                  }
                } catch (err) {
                  echo 'XRay failed'
                  echo err.getMessage()
                  // send Slack message to inform of XRay failure
                  slackSend(
                    channel: '#mayastor-e2e',
                    color: 'danger',
                    message: "E2E failed to send XRay reports - <$self_url|$self_name>."
                  )
                }
              }
            }// always
          }//post
        }//stage 'run e2e'
      }//stages
    }//stage 'test'
  }//stages

  // The main motivation for post block is that if all stages were skipped
  // (which happens when running cron job and branch != develop) then we don't
  // want to set commit status in github (jenkins will implicitly set it to
  // success).
  post {
    always {
      node(null) {
        script {
          // If no tests were run then we should neither be updating commit
          // status in github nor send any slack messages
          if (currentBuild.result != null) {
            step([
              $class: 'GitHubCommitStatusSetter',
              errorHandlers: [[$class: "ChangingBuildStatusErrorHandler", result: "UNSTABLE"]],
              contextSource: [
                $class: 'ManuallyEnteredCommitContextSource',
                context: 'continuous-integration/jenkins/branch'
              ],
              statusResultSource: [
                $class: 'ConditionalStatusResultSource',
                results: [
                  [$class: 'AnyBuildResult', message: 'Pipeline result', state: currentBuild.getResult()]
                ]
              ]
            ])
          }
        }
      }
    }//always
  }
}
