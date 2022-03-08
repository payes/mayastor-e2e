#!/usr/bin/env groovy

import java.util.concurrent.LinkedBlockingQueue
import java.util.concurrent.TimeUnit

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

def GetMCP(branch) {
  checkout([
    $class: 'GitSCM',
    branches: [[name: "*/${branch}"]],
    doGenerateSubmoduleConfigurations: false,
      extensions: [[
        $class: 'RelativeTargetDirectory',
        relativeTargetDir: "mayastor-control-plane"
      ]],
      submoduleCfg: [],
        userRemoteConfigs:
        [[url: "https://github.com/mayadata-io/mayastor-control-plane", credentialsId: "github-checkout"]]
    ])
}

def GetTestTag() {
  def tag = sh(
    script: 'printf $(date +"%Y-%m-%d-%H-%M-%S")',
    returnStdout: true
  )
  return tag
}

def BuildMCPImages(Map params) {
  def mayastorBranch = params['mayastorBranch']
  def mayastorRev = params['mayastorRev']
  def mcpBranch = params['mcpBranch']
  def mcpRev = params['mcpRev']
  def test_tag = params['test_tag']
  def build_flags = ""
  if (params.containsKey('build_flags')) {
        build_flags = params['build_flags']
  }

  println params

  if (!mayastorBranch?.trim() || !mcpBranch?.trim()) {
    throw new Exception("Empty branch parameters: mayastor branch is ${mayastorBranch}, mcp branch is ${mcpBranch}")
  }

  GetMayastor(mayastorBranch)

  // e2e tests are the most demanding step for space on the disk so we
  // test the free space here rather than repeating the same code in all
  // stages.
  sh "cd Mayastor && ./scripts/reclaim-space.sh 10"
  if (mayastorRev?.trim()) {
      sh "cd Mayastor && git checkout ${mayastorRev}"
  }
  sh "cd Mayastor && git submodule update --init"
  sh "cd Mayastor && git status"

  // Build images (REGISTRY is set in jenkin's global configuration).
  // Note: We might want to build and test dev images that have more
  // assertions instead but that complicates e2e tests a bit.
  // Build mayastor and mayastor-csi
  sh "cd Mayastor && ./scripts/release.sh $build_flags --registry \"${env.REGISTRY}\" --alias-tag \"$test_tag\" "

  // Build mayastor control plane
  GetMCP(mcpBranch)
  if (mcpRev?.trim()) {
      sh "cd mayastor-control-plane && git checkout ${mcpRev}"
  }
  sh "cd mayastor-control-plane && git submodule update --init"
  sh "cd mayastor-control-plane && git status"
  sh "cd mayastor-control-plane && ./scripts/release.sh $build_flags --registry \"${env.REGISTRY}\" --alias-tag \"$test_tag\" "

  // Build the install image
  sh "./scripts/create-install-image.sh $build_flags --alias-tag \"$test_tag\" --mayastor Mayastor --mcp mayastor-control-plane --registry \"${env.REGISTRY}\""

  // Limit any side-effects
  sh "rm -Rf Mayastor/"
  sh "rm -Rf mayastor-control-plane/"
}


def BuildCluster(e2e_build_cluster_job, e2e_environment) {
  def uuid = UUID.randomUUID()
  return build(
    job: "${e2e_build_cluster_job}",
    propagate: true,
    wait: true,
    parameters: [
      [
        $class: 'StringParameterValue',
        name: "ENVIRONMENT",
        value: "${e2e_environment}"
      ],
      [
        $class: 'StringParameterValue',
        name: "UUID",
        value: "${uuid}"
      ]
    ]
  )
}

def BuildK8sVersionedCluster(e2e_build_cluster_job, e2e_environment, kubernetes_verion) {
  def uuid = UUID.randomUUID()
  return build(
    job: "${e2e_build_cluster_job}",
    propagate: true,
    wait: true,
    parameters: [
      [
        $class: 'StringParameterValue',
        name: "ENVIRONMENT",
        value: "${e2e_environment}"
      ],
      [
        $class: 'StringParameterValue',
        name: "UUID",
        value: "${uuid}"
      ],
      [
        $class: 'StringParameterValue',
        name: "KUBERNETES_VERSION",
        value: "${kubernetes_verion}"
      ]
    ]
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

def GetIdentityFile(e2e_environment, k8s_job) {
  copyArtifacts(
    projectName: "${k8s_job.getProjectName()}",
    selector: specific("${k8s_job.getNumber()}"),
    filter: "${e2e_environment}/id_rsa",
    target: "",
    fingerprintArtifacts: true
  )
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
def LokiInstall(tag, test) {
  String loki_run_id = GetLokiRunId()
  sh 'kubectl apply -f ./loki/promtail_namespace_e2e.yaml'
  sh 'kubectl apply -f ./loki/promtail_rbac_e2e.yaml'
  sh 'kubectl apply -f ./loki/promtail_configmap_e2e.yaml'
  def cmd = "run=\"${loki_run_id}\" version=\"${tag}\" test=\"${test}\" envsubst -no-unset < ./loki/promtail_daemonset_e2e.template.yaml | kubectl apply -f -"
  sh "nix-shell --run '${cmd}'"
}

// Uninstall Loki tag
def LokiUninstall(tag, test) {
  String loki_run_id = GetLokiRunId()
  def cmd = "run=\"${loki_run_id}\" version=\"${tag}\" test=\"${test}\" envsubst -no-unset < ./loki/promtail_daemonset_e2e.template.yaml | kubectl delete -f -"
  sh "nix-shell --run '${cmd}'"
  sh 'kubectl delete -f ./loki/promtail_configmap_e2e.yaml'
  sh 'kubectl delete -f ./loki/promtail_rbac_e2e.yaml'
  sh 'kubectl delete -f ./loki/promtail_namespace_e2e.yaml'
}

// Install Loki on the cluster with kubernetes version 1.22 and above
def LokiInstallV1(tag, test) {
  String loki_run_id = GetLokiRunId()
  sh 'kubectl apply -f ./loki/promtail_namespace_e2e.yaml'
  sh 'kubectl apply -f ./loki/promtail_rbac_e2e_v1.yaml'
  sh 'kubectl apply -f ./loki/promtail_configmap_e2e.yaml'
  def cmd = "run=\"${loki_run_id}\" version=\"${tag}\" test=\"${test}\" envsubst -no-unset < ./loki/promtail_daemonset_e2e.template.yaml | kubectl apply -f -"
  sh "nix-shell --run '${cmd}'"
}

def GetTestList(profile) {
  def list = sh(
    script: "scripts/e2e-get-test-list.sh '${profile}'",
    returnStdout: true
  )
  return list
}

def RunOneTestPerCluster(e2e_test,
                          test_tag,
                          loki_run_id,
                          e2e_build_cluster_job,
                          e2e_destroy_cluster_job,
                          e2e_environment,
                          e2e_reports_dir,
                          kubernetes_verion) {
    def failed_tests=""
    def k8s_job=""
    def testset = "install ${e2e_test} uninstall"
    println testset
    println kubernetes_verion
    if (kubernetes_verion != "" && kubernetes_verion != null){
      echo "e2e_kubernetes_version=${kubernetes_verion}"
      k8s_job = BuildK8sVersionedCluster(e2e_build_cluster_job, e2e_environment, kubernetes_verion)
    } else {
      echo "e2e_kubernetes_version=1.21.8"
      k8s_job = BuildCluster(e2e_build_cluster_job, e2e_environment)
    }
    GetClusterAdminConf(e2e_environment, k8s_job)
    def session_id = e2e_test.replaceAll(",", "-")

    GetIdentityFile(e2e_environment, k8s_job)

    def cmd = "./scripts/e2e-test.sh --device /dev/sdb --tag \"${test_tag}\" --logs --onfail stop --tests \"${testset}\" --loki_run_id \"${loki_run_id}\" --loki_test_label \"${e2e_test}\" --reportsdir \"${env.WORKSPACE}/${e2e_reports_dir}\" --registry \"${env.REGISTRY}\" --session \"${session_id}\" --ssh_identity \"${env.WORKSPACE}/${e2e_environment}/id_rsa\" "
    withCredentials([
      usernamePassword(credentialsId: 'GRAFANA_API', usernameVariable: 'grafana_api_user', passwordVariable: 'grafana_api_pw'),
      string(credentialsId: 'HCLOUD_TOKEN', variable: 'HCLOUD_TOKEN')
    ]) {
      if (kubernetes_verion != "" && kubernetes_verion != null && kubernetes_verion != "1.21.7"){
          LokiInstallV1(test_tag, e2e_test)
      } else {
          LokiInstall(test_tag, e2e_test)
      }
      try {
        sh "nix-shell --run '${cmd}'"
      } catch(err) {
          failed_tests = e2e_test
      }
    }
    HandleUninstallReports(e2e_reports_dir, e2e_test)

    DestroyCluster(e2e_destroy_cluster_job, k8s_job)
    return failed_tests
}

def BuildTestsQueue(profile) {
  def list = sh(
    script: "scripts/e2e-get-test-list.sh '${profile}'",
    returnStdout: true
  )
  def tests = list.split()
  LinkedBlockingQueue testsQueue = [] as LinkedBlockingQueue
  //loop over list
  for (int i = 0; i < tests.size(); i++) {
    testsQueue.add(tests[i])
  }
  println testsQueue
  return testsQueue
}

def SendXrayReport(xray_testplan, test_tag, e2e_reports_dir) {
  xray_test_execution_type = '10059'
  def xray_projectkey = xray_testplan.split('-')[0]
  def pipeline = GetJobBaseName()
  def summary = "Pipeline: ${pipeline}, test plan: ${xray_testplan}, git branch: ${env.BRANCH_name}, tested image tag: ${test_tag}"

  // Ensure there is only one install junit report
  // for XRay with failed reports having priority.
  DeDupeInstallReports(e2e_reports_dir)

  try {
    step([
      $class: 'XrayImportBuilder',
      endpointName: '/junit/multipart',
      importFilePath: "${e2e_reports_dir}/**/*.xml",
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

// See below, for now we delete all uninstall reports.
// When fixed, this function should call the correct behaviour,
// e.g. AlterUninstallReports()
def HandleUninstallReports(e2e_reports_dir, e2e_test) {
    DeleteUninstallReports(e2e_reports_dir)
}

// deprecated until we decide how to keep the number of test results predictable
def AlterUninstallReports(e2e_reports_dir, e2e_test) {
    def junit_name = "e2e.uninstall-junit.xml"
    def junit_find_path = "${e2e_reports_dir}/*"
    def junit_full_find_path = " ${junit_find_path}/${junit_name}"
    def junit = sh(
        script: "find ${junit_find_path} -maxdepth 1 -name ${junit_name} | head -1",
        returnStdout: true
    )
    junit = junit.trim()
    if (junit != "") {
        sh "sed -i 's/classname=\"Basic Teardown Suite\"/classname=\"${e2e_test}\"/g' ${junit_full_find_path}"
        sh "sed -i 's/<testcase name=\"Mayastor setup should teardown using yamls\"/<testcase name=\"${e2e_test} should teardown using yamls\"/g' ${junit_full_find_path}"
    } else {
        println("no uninstall junit files found")
    }
}

def DeleteUninstallReports(e2e_reports_dir) {
    def junit_name = "e2e.uninstall-junit.xml"
    def junit_find_path = "${e2e_reports_dir}/*"
    def junit_full_find_path = " ${junit_find_path}/${junit_name}"
    sh "rm -f ${junit_full_find_path}"
}

def DeDupeInstallReports(e2e_reports_dir) {
    def junit_name = "e2e.install-junit.xml"
    def install_junit_find_path = " ${e2e_reports_dir}/*"
    def install_junit_full_find_path = " ${install_junit_find_path}/${junit_name}"

    def install_sample_junit = sh(
        script: "find ${install_junit_find_path} -maxdepth 1 -name ${junit_name} | head -1",
        returnStdout: true
    )
    install_sample_junit = install_sample_junit.trim()
    if (install_sample_junit == "") {
        println("no install junit files found")
        return
    }
    // find the first failed install report and use it if it exists
    def install_fail_junit = sh(
        script: "grep -lR  '<failure type=\"Failure\">'  ${install_junit_full_find_path} | head -1 || true",
        returnStdout: true
    )
    install_fail_junit = install_fail_junit.trim()
    if (install_fail_junit != "") {
        install_sample_junit = install_fail_junit
    }
    // rename it to a safe name and delete all others
    sh "mv ${install_sample_junit} ${install_sample_junit}.sample.xml"
    sh "rm ${install_junit_full_find_path}"
}

// Note Groovy passes parameters by value so any changes to params will not be reflected
// in the callers version of "params", the exceptions are the objects like the queue objects
def RunParallelStage(Map params) {
    // for simplicity "unwrap" required members of the map
    def tests_queue = params['tests_queue']
    def failed_tests_queue = params['failed_tests_queue']
    def e2e_image_tag = params['e2e_image_tag']
    def e2e_build_cluster_job = params['e2e_build_cluster_job']
    def e2e_destroy_cluster_job = params['e2e_destroy_cluster_job']
    def e2e_environment = params['e2e_environment']
    def e2e_reports_dir = params['e2e_reports_dir']
    def kubernetes_verion = params['kubernetes_verion']

    def loki_run_id = GetLokiRunId()

    while (tests_queue.size() > 0) {
        def e2e_test = tests_queue.poll(60L, TimeUnit.SECONDS)
        if (e2e_test != "" && e2e_test != null){
            def failed_test = RunOneTestPerCluster(e2e_test,
                                                    e2e_image_tag,
                                                    loki_run_id,
                                                    e2e_build_cluster_job,
                                                    e2e_destroy_cluster_job,
                                                    e2e_environment,
                                                    e2e_reports_dir,
                                                    kubernetes_verion)

            if (failed_test != "") {
                failed_tests_queue.add(failed_test)
            }
        }
    }
}

// Note Groovy passes parameters by value so any changes to params will not be reflected
// in the callers version of "params", the exceptions are the objects like the queue objects
def PostParallelStage(Map params, run_uuid) {
    // for simplicity "unwrap" required members of the map
    def junit_stash_queue = params['junit_stash_queue']
    def artefacts_stash_queue = params['artefacts_stash_queue']
    def e2e_reports_dir =  params['e2e_reports_dir']

    try {
        files = sh(script: "find ${e2e_reports_dir} -name *.xml", returnStdout: true).split()
        if (files.size() != 0) {
            stash_name = "junit-${run_uuid}"
            stash includes: "${e2e_reports_dir}/**/*.xml", name: stash_name
            junit_stash_queue.add(stash_name)
        } else {
            println "no junit files"
        }
    } catch (err) {
    }

    try {
        files = sh(script: "find artifacts/ -type f", returnStdout: true).split()
        if (files.size() != 0) {
            stash_name = "arts-${run_uuid}"
            stash includes: 'artifacts/**/**', name: stash_name
            artefacts_stash_queue.add(stash_name)
        } else {
            println "no files to archive"
        }
    } catch(err) {
    }
}

// Note Groovy passes parameters by value so any changes to params will not be reflected
// in the callers version of "params", the exceptions are the objects like the queue objects
def ParallelArchiveArtefacts(Map params) {
    // for simplicity "unwrap" all members of the map
    def artefacts_stash_queue = params['artefacts_stash_queue']

    while (artefacts_stash_queue.size() > 0) {
        def stash_name = artefacts_stash_queue.poll()
        unstash name: stash_name
    }
    archiveArtifacts 'artifacts/**/**'
}

// Note Groovy passes parameters by value so any changes to params will not be reflected
// in the callers version of "params", the exceptions are the objects like the queue objects
def ParallelJunit(Map params) {
    // for simplicity "unwrap" all members of the map
    def e2e_image_tag = params['e2e_image_tag']
    def e2e_reports_dir = params['e2e_reports_dir']
    def junit_stash_queue = params['junit_stash_queue']
    def xray_send_report = params['xray_send_report']
    def xray_test_plan = params['xray_test_plan']

    while (junit_stash_queue.size() > 0) {
        def stash_name = junit_stash_queue.poll()
        unstash name: stash_name
    }

    junit testResults: "${e2e_reports_dir}/**/*.xml", skipPublishingChecks: true

    if (xray_send_report == true) {
        // xray_send_report alters the report artifacts
        // so should run after archiveArtifacts and junit
        SendXrayReport(xray_test_plan, e2e_image_tag, e2e_reports_dir)
    }
}

// Note Groovy passes parameters by value so any changes to params will not be reflected
// in the callers version of "params", the exceptions are the objects like the queue objects
def PopulateTestQueue(Map params) {
  def tests_queue = params['tests_queue']
  def e2e_test_profile = params['e2e_test_profile']

  def list = sh(
    script: "scripts/e2e-get-test-list.sh '${e2e_test_profile}'",
    returnStdout: true
  )
  def tests = list.split()
  //loop over list
  for (int i = 0; i < tests.size(); i++) {
    tests_queue.add(tests[i])
  }
}

def StashMayastorBinaries(Map params) {
    def artefacts_stash_queue = params['artefacts_stash_queue']
    def test_tag = params['e2e_image_tag']
    def bin_dir = "./artifacts/binaries/${test_tag}"

    sh "./scripts/get-mayastor-binaries.py --tag ${test_tag} --registry ${env.REGISTRY} --outputdir ${bin_dir}"
    def stash_name = 'arts-bin'
    stash includes: 'artifacts/**/**', name: stash_name
    artefacts_stash_queue.add(stash_name)
}

def CoverageReport(Map params) {
    def artefacts_stash_queue = params['artefacts_stash_queue']
    def test_tag = params['e2e_image_tag']
    def mayastorBranch = params['mayastorBranch']
    def mcpBranch = params['mcpBranch']

    GetMayastor(mayastorBranch)
    GetMCP(mcpBranch)

    while (artefacts_stash_queue.size() > 0) {
        def stash_name = artefacts_stash_queue.poll()
        unstash name: stash_name
    }

    def mayastor_dir = "${env.WORKSPACE}/Mayastor"
    def mcp_dir = "${env.WORKSPACE}/mayastor-control-plane"
    def data_dir = "${env.WORKSPACE}/artifacts/coverage/data"
    def bin_dir = "${env.WORKSPACE}/artifacts/binaries/${test_tag}"
    def report_dir = "${env.WORKSPACE}/artifacts/coverage/report"

    sh "./scripts/get-mayastor-binaries.py --tag ${test_tag} --registry ${env.REGISTRY} --outputdir ${bin_dir}"

    sh "./scripts/coverage-report.sh -d ${data_dir} -b ${bin_dir} -M ${mayastor_dir} -C ${mcp_dir} -r ${report_dir}"

    stash_name = 'artefcats-with-coverage-report'
    stash includes: 'artifacts/**/**', name: stash_name
    artefacts_stash_queue.add(stash_name)
}


return this
