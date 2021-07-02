#!/usr/bin/env groovy

// On-demand E2E infra configuration
// https://mayadata.atlassian.net/wiki/spaces/MS/pages/247332965/Test+infrastructure#On-Demand-E2E-K8S-Clusters

def e2e_build_cluster_job='k8s-build-cluster' // Jenkins job to build cluster
def e2e_destroy_cluster_job='k8s-destroy-cluster' // Jenkins job to destroy cluster
// Environment to run e2e test in (job param of $e2e_build_cluster_job)
def e2e_environment="hcloud-kubeadm"
// Global variable to pass current k8s job between stages
def k8s_job=""

e2e_image_tag='selfci'
e2e_local_registry=true
e2e_reports_dir='artifacts/reports/'
e2e_test_profile = 'selfci'

run_linter=true

xray_send_report=false
xray_self_ci_testplan='MQ-482'

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
              common = load "${env.WORKSPACE}/pipelines/common/common.groovy"
              k8s_job = common.BuildCluster(e2e_build_cluster_job, e2e_environment)
            }
          }
        }
        stage('run e2e') {
          agent { label 'nixos' }
          environment {
            KUBECONFIG = "${env.WORKSPACE}/${e2e_environment}/modules/k8s/secrets/admin.conf"
          }
          steps {
            script {
              common = load "./pipelines/common/common.groovy"
              common.GetClusterAdminConf(e2e_environment, k8s_job)
              loki_run_id = common.GetLokiRunId()
              sh "mkdir -p ./${e2e_reports_dir}"

              def cmd = "./scripts/e2e-test.sh --device /dev/sdb --tag \"${e2e_image_tag}\" --logs --profile \"${e2e_test_profile}\" --loki_run_id \"${loki_run_id}\" --reportsdir \"${env.WORKSPACE}/${e2e_reports_dir}\" "
              if (e2e_local_registry == true) {
                cmd = cmd + " --registry \"" + env.REGISTRY + "\""
              }

              withCredentials([
                usernamePassword(credentialsId: 'GRAFANA_API', usernameVariable: 'grafana_api_user', passwordVariable: 'grafana_api_pw'),
                string(credentialsId: 'HCLOUD_TOKEN', variable: 'HCLOUD_TOKEN')
              ]) {
                common.LokiInstall(e2e_image_tag)
                sh "nix-shell --run '${cmd}'"
                common.LokiUninstall(e2e_image_tag)
              }
            }
          }
          post {
            always {
              archiveArtifacts 'artifacts/**/*.*'
              // handle junit results on success or failure
              junit "${e2e_reports_dir}/*.xml"
              script {
                common = load "${env.WORKSPACE}/pipelines/common/common.groovy"
                common.DestroyCluster(e2e_destroy_cluster_job, k8s_job)
                if (xray_send_report == true) {
                  def pipeline = common.GetJobBaseName()
                  def summary = "Pipeline: ${pipeline}, test plan: ${xray_self_ci_testplan}, git branch: ${env.BRANCH_name}, tested image tag: ${e2e_image_tag}"
                  common.SendXrayReport(xray_self_ci_testplan, summary, e2e_reports_dir)
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
              reposSource: [
                $class: 'ManuallyEnteredRepositorySource',
                url: 'https://github.com/mayadata-io/mayastor-e2e'
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
