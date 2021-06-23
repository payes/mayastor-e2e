#!/usr/bin/env groovy

// In the case of multi-branch pipelines, the pipeline
// name a.k.a. job base name, will be the
// 2nd-to-last item of env.JOB_NAME which
// consists of identifiers separated by '/' e.g.
//     first/second/pipeline/branch
// In the case of a non-multibranch pipeline, the pipeline
// name is env.JOB_NAME. This caters for all eventualities.
def GetJobBaseName() {
  def jobSections = env.JOB_NAME.tokenize('/') as String[]
  return jobSections.length < 2 ? env.JOB_NAME : jobSections[ jobSections.length - 2 ]
}

def GetLokiRunId() {
  String job_base_name = GetJobBaseName()
  return job_base_name + "-" + env.BRANCH_NAME + "-" + env.BUILD_NUMBER
}

// Checks out the specified branch of Mayastor to the current repo
def GetMayastor(branch) {
  checkout([
    $class: 'GitSCM',
    branches: [[name: "*/${branch}"]],
    doGenerateSubmoduleConfigurations: false,
      extensions: [[
        $class: 'RelativeTargetDirectory',
        relativeTargetDir: "Mayastor"
      ]],
      submoduleCfg: [],
        userRemoteConfigs:
        [[url: "https://github.com/openebs/Mayastor", credentialsId: "github-checkout"]]
    ])
}

def GetMoac(branch) {
  checkout([
    $class: 'GitSCM',
    branches: [[name: "*/${branch}"]],
    doGenerateSubmoduleConfigurations: false,
      extensions: [[
        $class: 'RelativeTargetDirectory',
        relativeTargetDir: "moac"
      ]],
      submoduleCfg: [],
        userRemoteConfigs:
        [[url: "https://github.com/openebs/moac", credentialsId: "github-checkout"]]
    ])
}

def GetTestTag() {
  def tag = sh(
    script: 'printf $(date +"%Y-%m-%d-%H-%M-%S")',
    returnStdout: true
  )
  return tag
}

def BuildImages(mayastorBranch, moacBranch, test_tag) {
  GetMayastor(mayastorBranch)

  // e2e tests are the most demanding step for space on the disk so we
  // test the free space here rather than repeating the same code in all
  // stages.
  sh "cd Mayastor && ./scripts/reclaim-space.sh 10"

  // Build images (REGISTRY is set in jenkin's global configuration).
  // Note: We might want to build and test dev images that have more
  // assertions instead but that complicates e2e tests a bit.
  // Build mayastor and mayastor-csi
  sh "cd Mayastor && ./scripts/release.sh --registry \"${env.REGISTRY}\" --alias-tag \"$test_tag\" "

  // Build moac
  GetMoac(moacBranch)
  sh "cd moac && ./scripts/release.sh --registry \"${env.REGISTRY}\" --alias-tag \"$test_tag\" "

  // Build the install image
  sh "./scripts/create-install-image.sh --alias-tag \"$test_tag\" --mayastor Mayastor --moac moac --registry \"${env.REGISTRY}\""

  // Limit any side-effects
  sh "rm -Rf Mayastor/"
  sh "rm -Rf moac/"
}

def BuildCluster(e2e_build_cluster_job, e2e_environment) {
  return build(
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

def DestroyCluster(e2e_destroy_cluster_job, k8s_job) {
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
}

def GetClusterAdminConf(e2e_environment, k8s_job) {
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
}

def WarnOrphanCluster(k8s_job) {
  withCredentials([string(credentialsId: 'HCLOUD_TOKEN', variable: 'HCLOUD_TOKEN')]) {
    e2e_nodes=sh(
      script: """
        nix-shell -p hcloud --run 'hcloud server list' | grep -e '-${k8s_job.getNumber()} ' | awk '{ print \$2" "\$4 }'
              """,
        returnStdout: true
    ).trim()
  }

  // Job name for multi-branch is Mayastor/<branch> however
  // in URL jenkins requires /job/ in between for url to work
  urlized_job_name=JOB_NAME.replaceAll("/", "/job/")
  self_url="${JENKINS_URL}job/${urlized_job_name}/${BUILD_NUMBER}"
  self_name="${JOB_NAME}#${BUILD_NUMBER}"
  build_cluster_run_url="${JENKINS_URL}job/${k8s_job.getProjectName()}/${k8s_job.getNumber()}"
  build_cluster_destroy_url="${JENKINS_URL}job/${e2e_destroy_cluster_job}/buildWithParameters?BUILD=${k8s_job.getProjectName()}%23${k8s_job.getNumber()}"
  kubeconfig_url="${JENKINS_URL}job/${k8s_job.getProjectName()}/${k8s_job.getNumber()}/artifact/hcloud-kubeadm/modules/k8s/secrets/admin.conf"
  slackSend(
    channel: '#mayastor-e2e',
    color: 'danger',
    message: "E2E k8s cluster <$build_cluster_run_url|#${k8s_job.getNumber()}> left running due to failure of " +
             "<$self_url|$self_name>. Investigate using <$kubeconfig_url|kubeconfig>, or ssh as root to:\n" +
             "```$e2e_nodes```\n" +
             "And then <$build_cluster_destroy_url|destroy> the cluster.\n" +
             "Note: you need to click `proceed` and will get an empty page when using destroy link. " +
             "(<https://mayadata.atlassian.net/wiki/spaces/MS/pages/247332965/Test+infrastructure#On-Demand-E2E-K8S-Clusters|doc>)"
  )
}

// Install Loki on the cluster
def LokiInstall(tag) {
  String loki_run_id = GetLokiRunId()
  sh 'kubectl apply -f ./loki/promtail_namespace_e2e.yaml'
  sh 'kubectl apply -f ./loki/promtail_rbac_e2e.yaml'
  sh 'kubectl apply -f ./loki/promtail_configmap_e2e.yaml'
  def cmd = "run=\"${loki_run_id}\" version=\"${tag}\" envsubst -no-unset < ./loki/promtail_daemonset_e2e.template.yaml | kubectl apply -f -"
  sh "nix-shell --run '${cmd}'"
}

// Unnstall Loki
def LokiUninstall(tag) {
  String loki_run_id = GetLokiRunId()
  def cmd = "run=\"${loki_run_id}\" version=\"${tag}\" envsubst -no-unset < ./loki/promtail_daemonset_e2e.template.yaml | kubectl delete -f -"
  sh "nix-shell --run '${cmd}'"
  sh 'kubectl delete -f ./loki/promtail_configmap_e2e.yaml'
  sh 'kubectl delete -f ./loki/promtail_rbac_e2e.yaml'
  sh 'kubectl delete -f ./loki/promtail_namespace_e2e.yaml'
}

def SendXrayReport(xray_testplan, summary, e2e_reports_dir) {
  xray_projectkey = 'MQ'
  xray_test_execution_type = '10059'

  try {
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
          "summary": "${summary}",
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

return this