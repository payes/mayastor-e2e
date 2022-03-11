#!/usr/bin/env groovy

import java.util.concurrent.LinkedBlockingQueue
import java.util.concurrent.TimeUnit

def unwrap(Map params, key) {
    if (params.containsKey(key)) {
        return params[key]
    }
    error("value not defined for ${key}")
}

def GetProductSettings (datacore_bolt) {
    if (datacore_bolt == true) {
        return [
            dataplane_dir: "bolt-data-plane",
            dataplane_repo_url: "git@github.com:DataCoreSoftware/bolt-data-plane.git",
            controlplane_dir: "bolt-control-plane",
            controlplane_repo_url: "git@github.com:DataCoreSoftware/bolt-control-plane.git",
            github_credentials: 'BOLT_CICD_GITHUB_SSH_KEY',
        ]
    }
    return [
        dataplane_dir: "Mayastor",
        dataplane_repo_url: "https://github.com/openebs/Mayastor",
        controlplane_dir: "mayastor-control-plane",
        controlplane_repo_url: "https://github.com/openebs/mayastor-control-plane",
        github_credentials: 'github-checkout',
    ]
}

def GetE2ESettings() {
    def e2e_dir = "mayastor-e2e"
    return [
        e2e_dir: e2e_dir,
        e2e_repo_url: "https://github.com/mayadata-io/mayastor-e2e.git",
        e2e_reports_dir: "${e2e_dir}/artifacts/reports/",
        e2e_artifacts_dir: "${e2e_dir}/artifacts/"
    ]
}

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

def CheckoutRepo(branch, relativeTargetDir, url, credentials, commitOrTag) {
  sh """
    rm -rf ${relativeTargetDir}
    """
  bct = "*/${branch}"
  if ( commitOrTag != "" ) {
    bct = commitOrTag
  }
  println("branches name = $bct")
  checkout([
    $class: 'GitSCM',
    branches: [[name: "${bct}"]],
    doGenerateSubmoduleConfigurations: false,
    extensions: [
        [
            $class: 'CloneOption',
            noTags: false,
        ],
        [
            $class: 'RelativeTargetDirectory',
            relativeTargetDir: relativeTargetDir
        ],
        [
            $class: 'SubmoduleOption',
            disableSubmodules: false,
            parentCredentials: true,
            recursiveSubmodules: true,
            reference: '',
            trackingSubmodules: false
        ]
    ],
    submoduleCfg: [],
    userRemoteConfigs: [[
        url: url,
        credentialsId: credentials
      ]]
    ])
}

// Checks out the specified branch of E2E repo
def CheckoutE2E(params) {
  def branch = unwrap(params,'e2eBranch')
  def relativeTargetDir = unwrap(params,'e2e_dir')
  def url = unwrap(params,'e2e_repo_url')
  def credentials = unwrap(params,'github_credentials')
  def commitOrTag = unwrap(params, 'e2eCommitOrTag')

  CheckoutRepo(branch, relativeTargetDir, url, credentials, commitOrTag)
}

// Checks out the specified branch of the data plane repo
def CheckoutDataPlane(params) {
  def branch = unwrap(params,'dataplaneBranch')
  def relativeTargetDir = unwrap(params,'dataplane_dir')
  def url = unwrap(params,'dataplane_repo_url')
  def credentials = unwrap(params,'github_credentials')
  def commitOrTag = unwrap(params, 'dataplaneCommitOrTag')

  CheckoutRepo(branch, relativeTargetDir, url, credentials, commitOrTag)
}

// Checks out the specified branch of the control plane repo
def CheckoutControlPlane(params) {
  def branch = unwrap(params,'controlplaneBranch')
  def relativeTargetDir = unwrap(params,'controlplane_dir')
  def url = unwrap(params,'controlplane_repo_url')
  def credentials = unwrap(params,'github_credentials')
  def commitOrTag = unwrap(params, 'controlplaneCommitOrTag')

  CheckoutRepo(branch, relativeTargetDir, url, credentials, commitOrTag)
}

def GetTestTag() {
  def tag = sh(
    script: 'printf $(date +"%Y-%m-%d-%H-%M-%S")',
    returnStdout: true
  )
  return tag
}

def BuildImages2(Map params) {
  def dataplaneCommitOrTag = unwrap(params,'dataplaneCommitOrTag')
  def controlplaneCommitOrTag = unwrap(params,'controlplaneCommitOrTag')
  def test_tag = unwrap(params,'test_tag')
  def dataplane_dir = unwrap(params,'dataplane_dir')
  def controlplane_dir = unwrap(params,'controlplane_dir')
  def product = unwrap(params,'product')
  def build_flags = ""
  if (params.containsKey('build_flags')) {
        build_flags = params['build_flags']
  }

  PrettyPrintMap("BuildImages2 params", params)

  CheckoutDataPlane(params)

  // e2e tests are the most demanding step for space on the disk so we
  // test the free space here rather than repeating the same code in all
  // stages.
  sh "cd ${dataplane_dir} && ./scripts/reclaim-space.sh 10"
  sh "cd ${dataplane_dir} && git status"

  // Build images (REGISTRY is set in jenkin's global configuration).
  // Note: We might want to build and test dev images that have more
  // assertions instead but that complicates e2e tests a bit.
  // Build mayastor and mayastor-csi
  sh "cd ${dataplane_dir} && ./scripts/release.sh $build_flags --registry \"${env.REGISTRY}\" --alias-tag \"$test_tag\" "

  // Build mayastor control plane
  CheckoutControlPlane(params)
  sh "cd ${controlplane_dir} && git status"
  sh "cd ${controlplane_dir} && ./scripts/release.sh $build_flags --registry \"${env.REGISTRY}\" --alias-tag \"$test_tag\" "

  // NOTE: create-install-image.sh should be part of the Jenkins repo
  // not mayastor-e2e
  // Build the install image
  sh "./scripts/create-install-image.sh $build_flags --alias-tag \"$test_tag\" --mayastor ${dataplane_dir} --mcp ${controlplane_dir} --registry \"${env.REGISTRY}\" --product \"${product}\" "

  // Limit any side-effects
  sh "rm -Rf ${dataplane_dir}/"
  sh "rm -Rf ${controlplane_dir}/"
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

def BuildK8sVersionedCluster(e2e_build_cluster_job, e2e_environment, kubernetes_version) {
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
        value: "${kubernetes_version}"
      ]
    ]
  )
}

def DestroyCluster(e2e_destroy_cluster_job, k8s_job) {
  // This relatively complex algorithm to destroy a cluster
  // because it has been observed that 2 or more concurrent
  // cluster destroy build jobs are conflated and only one
  // cluster is destroyed.
  // To combat this the destroy cluster job creates an artifact
  // with the cluster number destroyed.
  // This function retrieves that artifact, compares the value contained
  // within with the requested value.
  // If the values are unequal the request to destroy the cluster is
  // repeated upto N times.
  def desired = k8s_job.getNumber()
  def actual = ""

  for (int ix = 10; ix > 0; ix--) {
      println("k8s-destroy-cluster #${k8s_job.getNumber()}")
      built = build(
        job: "${e2e_destroy_cluster_job}",
        propagate: false,
        wait: true,
        parameters: [
          [
            $class: 'RunParameterValue',
            name: "BUILD",
            runId:"${k8s_job.getProjectName()}#${k8s_job.getNumber()}"
          ]
        ]
      )
      copyArtifacts(
            projectName: "${built.getFullProjectName()}",
            selector: specific("${built.getNumber()}"),
            filter: "BUILD_NUMBER",
            target: "",
            fingerprintArtifacts: false)

        actual = readFile('BUILD_NUMBER').trim()
        if ("$actual" == "$desired") {
            println("k8s-destroy-cluster #$actual destroyed")
            return
        }

        println("k8s-destroy-cluster mismatch desired=$desired actual=$actual retries=${ix-1}")
        sleep 10
    }

    println("WARNING!!!! Failed to destroy cluster $desired")
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

def GetSelfUrlName() {
  // Job name for multi-branch is Mayastor/<branch> however
  // in URL jenkins requires /job/ in between for url to work
  def urlized_job_name=JOB_NAME.replaceAll("/", "/job/")
  def self_url="${JENKINS_URL}job/${urlized_job_name}/${BUILD_NUMBER}"
  def self_name="${JOB_NAME}#${BUILD_NUMBER}"
  return [ self_url, self_name ]
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

  def (self_url, self_name) = GetSelfUrlName()
  def build_cluster_run_url="${JENKINS_URL}job/${k8s_job.getProjectName()}/${k8s_job.getNumber()}"
  def build_cluster_destroy_url="${JENKINS_URL}job/${e2e_destroy_cluster_job}/buildWithParameters?BUILD=${k8s_job.getProjectName()}%23${k8s_job.getNumber()}"
  def kubeconfig_url="${JENKINS_URL}job/${k8s_job.getProjectName()}/${k8s_job.getNumber()}/artifact/hcloud-kubeadm/modules/k8s/secrets/admin.conf"
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

def RunOneTestPerCluster(e2e_test, loki_run_id, params) {
    // for simplicity "unwrap" required members of the map
    def e2e_image_tag = unwrap(params,'e2e_image_tag')
    def e2e_build_cluster_job = unwrap(params,'e2e_build_cluster_job')
    def e2e_destroy_cluster_job = unwrap(params,'e2e_destroy_cluster_job')
    def e2e_environment = unwrap(params,'e2e_environment')
    def e2e_reports_dir = unwrap(params,'e2e_reports_dir')
    def e2e_dir = unwrap(params,'e2e_dir')
    def kubernetes_version = unwrap(params,'kubernetes_version')
    def test_platform = unwrap(params,'test_platform')
    def product = unwrap(params,'product')

    def failed_tests=""
    def k8s_job=""
    def testset = "install ${e2e_test} uninstall"

    if (kubernetes_version == "" || kubernetes_version == null) {
        error("kubernetes version is not specified")
    }

    println("RunOneTestPerCluster ${e2e_test}")
    for (int i=0; i<3; i++) {
      echo "try build cluster ${i}"
      try {
          k8s_job = BuildK8sVersionedCluster(e2e_build_cluster_job, e2e_environment, kubernetes_version)
      } catch (err) {
        // if infrastructure has problems then do not retry
        println("Test infrastructure failed to create cluster")
        return [e2e_test, true]
      }
      GetClusterAdminConf(e2e_environment, k8s_job)
      try {
            // retry 10 times with 60 seconds gap for all nodes in the cluster
            // to synchronize on environments.
            sh "nix-shell --run './scripts/get_cluster_env.py -v --retry 10 --sleeptime 60'"
            break
      } catch(err) {
        DestroyCluster(e2e_destroy_cluster_job, k8s_job)
        k8s_job = ""
      }
    }

    if (k8s_job == "") {
        println("Failed to create cluster with uniform environment")
        return [e2e_test, true]
    }

    def session_id = e2e_test.replaceAll(",", "-")

    GetIdentityFile(e2e_environment, k8s_job)

    def reports_dir = "${env.WORKSPACE}/${e2e_reports_dir}/${e2e_test}"
    def envs_txt_file = "${reports_dir}/xray_test_environments.txt"
    def envs_yaml_file = "${reports_dir}/test_environments.yaml"
    // Record environment settings to files
    sh """
        mkdir -p "${reports_dir}"
        nix-shell --run './scripts/get_cluster_env.py --platform "${test_platform}" --oxray  "${envs_txt_file}" --oyaml "${envs_yaml_file}"'
    """
    
    def cmd = "cd ${e2e_dir} && ./scripts/e2e-test.sh --device /dev/sdb --tag \"${e2e_image_tag}\"  --onfail stop --tests \"${testset}\" --loki_run_id \"${loki_run_id}\" --loki_test_label \"${e2e_test}\" --reportsdir \"${env.WORKSPACE}/${e2e_reports_dir}\" --registry \"${env.REGISTRY}\" --session \"${session_id}\" --ssh_identity \"${env.WORKSPACE}/${e2e_environment}/id_rsa\" --product \"${product}\" "
    withCredentials([
      usernamePassword(credentialsId: 'GRAFANA_API', usernameVariable: 'grafana_api_user', passwordVariable: 'grafana_api_pw'),
      string(credentialsId: 'HCLOUD_TOKEN', variable: 'HCLOUD_TOKEN')
    ]) {
      if (kubernetes_version != "" && kubernetes_version != null && kubernetes_version != "1.21.7"){
          LokiInstallV1(e2e_image_tag, e2e_test)
      } else {
          LokiInstall(e2e_image_tag, e2e_test)
      }
      try {
        sh "nix-shell --run '${cmd}'"
      } catch(err) {
          failed_tests = e2e_test
      }
    }
    HandleUninstallReports(e2e_reports_dir, e2e_test)

    DestroyCluster(e2e_destroy_cluster_job, k8s_job)
    return [failed_tests, false]
}

def SendXrayReport(xray_testplan, test_tag, e2e_reports_dir, test_environments) {
  if (xray_testplan == "" ) {
    return
  }
  xray_test_execution_type = '10059'
  def xray_projectkey = xray_testplan.split('-')[0]
  def pipeline = GetJobBaseName()
  def summary = "Pipeline: ${pipeline}, test plan: ${xray_testplan}, git branch: ${env.BRANCH_name}, tested image tag: ${test_tag}"

  // Ensure there is only one install junit report
  // for XRay with failed reports having priority.
  DeDupeInstallReports(e2e_reports_dir)

  def xray_params = [
      $class: 'XrayImportBuilder',
      endpointName: '/junit/multipart',
      importFilePath: "${e2e_reports_dir}/**/*.xml",
      importToSameExecution: 'true',
      projectKey: "${xray_projectkey}",
      testPlanKey: "${xray_testplan}",
      serverInstance: "${env.JIRASERVERUUID}",
      testEnvironments: "${test_environments}",
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
  ]

  println(xray_params)

  try {
    step(xray_params)
  } catch (err) {
    echo 'XRay failed'
    echo err.getMessage()
    def (self_url, self_name) = GetSelfUrlName()
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
        // FIXME: bolt
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
    sh "rm -f ${install_junit_full_find_path}"
}

// Note Groovy passes parameters by value so any changes to params will not be reflected
// in the callers version of "params", the exceptions are the objects like the queue objects
def RunParallelStage(Map params) {
    // for simplicity "unwrap" required members of the map
    def tests_queue = unwrap(params,'tests_queue')
    def failed_tests_queue = unwrap(params,'failed_tests_queue')
    def passed_tests_queue = unwrap(params,'passed_tests_queue')
    def loki_run_id = GetLokiRunId()

    CheckoutE2E(params)
    while (tests_queue.size() > 0) {
        def e2e_test = tests_queue.poll(60L, TimeUnit.SECONDS)
        if (e2e_test != "" && e2e_test != null){
            def (failed_test, failed_cluster) = RunOneTestPerCluster(e2e_test, loki_run_id, params)

            if (failed_cluster == true) {
                // Scale down parallel runs
                println("Failed to create a cluster for test, stopping this parallel stage")
                tests_queue.add(e2e_test)
                break
            }

            if (failed_test != "") {
                failed_tests_queue.add(failed_test)
            } else {
                passed_tests_queue.add(e2e_test)
            }
        }
    }
}

// Note Groovy passes parameters by value so any changes to params will not be reflected
// in the callers version of "params", the exceptions are the objects like the queue objects
def PostParallelStage(Map params, run_uuid) {
    // for simplicity "unwrap" required members of the map
    def junit_stash_queue = unwrap(params,'junit_stash_queue')
    def artefacts_stash_queue = unwrap(params,'artefacts_stash_queue')
    def e2e_reports_dir =  unwrap(params,'e2e_reports_dir')
    def e2e_artifacts_dir = unwrap(params,'e2e_artifacts_dir')

    try {
        files = sh(script: "find ${e2e_reports_dir} -name *.xml", returnStdout: true).split()
        if (files.size() != 0) {
            stash_name = "junit-${run_uuid}"
            stash includes: "${e2e_reports_dir}/**/**", name: stash_name
            junit_stash_queue.add(stash_name)
        } else {
            println "no junit files"
        }
    } catch (err) {
    }

    try {
        files = sh(script: "find ${e2e_artifacts_dir}/ -type f", returnStdout: true).split()
        if (files.size() != 0) {
            stash_name = "arts-${run_uuid}"
            stash includes: "${e2e_artifacts_dir}/**/**", name: stash_name
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
    def artefacts_stash_queue = unwrap(params,'artefacts_stash_queue')
    def passed_tests_queue = unwrap(params,'passed_tests_queue')
    def failed_tests_queue = unwrap(params,'failed_tests_queue')
    def e2e_artifacts_dir = unwrap(params,'e2e_artifacts_dir')
    def artifacts_dir ='artifacts'

    sh """
        rm -rf ${e2e_artifacts_dir}
        rm -rf ${artifacts_dir}
        mkdir -p ${e2e_artifacts_dir}
        mkdir -p ${artifacts_dir}
    """

    if (artefacts_stash_queue.size() > 0) {
        while (artefacts_stash_queue.size() > 0) {
            def stash_name = artefacts_stash_queue.poll()
            unstash name: stash_name
        }
    } else {
        println("No artificats were stashed!")
    }

    def passed = "${passed_tests_queue}"
    def failed = "${failed_tests_queue}"
    writeFile(file: "${artifacts_dir}/passed_tests.txt", text: passed)
    writeFile(file: "${artifacts_dir}/failed_tests.txt", text: failed)

    archiveArtifacts    artifacts: "${e2e_artifacts_dir}/**/**,${artifacts_dir}/**/**",
                        fingerprint: true,
                        allowEmptyArchive: true
}

// Note Groovy passes parameters by value so any changes to params will not be reflected
// in the callers version of "params", the exceptions are the objects like the queue objects
def ParallelJunit(Map params) {
    // for simplicity "unwrap" all members of the map
    def e2e_image_tag = unwrap(params,'e2e_image_tag')
    def e2e_reports_dir = unwrap(params,'e2e_reports_dir')
    def junit_stash_queue = unwrap(params,'junit_stash_queue')
    def xray_send_report = unwrap(params,'xray_send_report')
    def xray_test_plan = unwrap(params,'xray_test_plan')

    if (junit_stash_queue.size() == 0) {
        println "No stashed results found, nothing to do."
        return
    }

    while (junit_stash_queue.size() > 0) {
        def stash_name = junit_stash_queue.poll()
        unstash name: stash_name
    }

    // FIXME: one of
    //  - assert that the content of all test environments files are the same
    //  - create a string which is the union of all test environments
    //  - report each test result individually with its environment
    //  - just use the first file
    def test_environments = sh(
        script: "find ${e2e_reports_dir} -name xray_test_environments.txt  | head -1 | xargs cat",
        returnStdout: true
    )
    test_environments.trim()
    println "test_environments = ${test_environments}"

    junit testResults: "${e2e_reports_dir}/**/*.xml", skipPublishingChecks: true

    if (xray_send_report == true) {
        // xray_send_report alters the report artifacts
        // so should run after archiveArtifacts and junit
        SendXrayReport(xray_test_plan, e2e_image_tag, e2e_reports_dir, test_environments)
    }
}

// Note Groovy passes parameters by value so any changes to params will not be reflected
// in the callers version of "params", the exceptions are the objects like the queue objects
def PopulateTestQueue(Map params) {
  def tests_queue = unwrap(params,'tests_queue')
  def e2e_test_profile = unwrap(params,'e2e_test_profile')
  def e2e_tests = unwrap(params,'e2e_tests')
//  def e2e_artifacts_dir = unwrap(params,'e2e_artifacts_dir')
  def artefacts_stash_queue = unwrap(params,'artefacts_stash_queue')
  def tests=[]
  def artifacts_dir = 'artifacts'

  sh """
    mkdir -p ${artifacts_dir}
  """

  if (e2e_test_profile != "") {
    CheckoutE2E(params)
    def e2e_dir = unwrap(params,'e2e_dir')
    def listsfile = "${e2e_dir}/configurations/testlists.yaml"
    def lstfile = "${artifacts_dir}/testlist"
    def cmdx = "./scripts/testlists.py --profile '${e2e_test_profile}' --lists ${listsfile} --output '${lstfile}' --sort_duration"
    sh(
        script: """
            nix-shell --run '${cmdx}'
        """
    )
    list = readFile(lstfile).trim()
    tests = list.split()
  }

  if (e2e_tests != "") {
    tests = e2e_tests.split()
  }

  println("List of tests: ${tests}")
  //loop over list
  for (int i = 0; i < tests.size(); i++) {
      tests_queue.add(tests[i])
  }
  writeFile(file: "${artifacts_dir}/tests.txt", text: "$tests_queue")

  def stash_name = 'arts-test-list'
  stash includes: "${artifacts_dir}/**/**", name: stash_name
  artefacts_stash_queue.add(stash_name)
}

def StashMayastorBinaries(Map params) {
    def e2e_artifacts_dir = unwrap(params,'e2e_artifacts_dir')
    def artefacts_stash_queue = unwrap(params,'artefacts_stash_queue')
    def test_tag = unwrap(params,'e2e_image_tag')
    def bin_dir = "./${e2e_artifacts_dir}/binaries/${test_tag}"

    sh "./scripts/get-mayastor-binaries.py --tag ${test_tag} --registry ${env.REGISTRY} --outputdir ${bin_dir}"
    def stash_name = 'arts-bin'
    stash includes: "${e2e_artifacts_dir}/**/**", name: stash_name
    artefacts_stash_queue.add(stash_name)
}

def CoverageReport(Map params) {
    def e2e_artifacts_dir = unwrap(params,'e2e_artifacts_dir')
    def artefacts_stash_queue = unwrap(params,'artefacts_stash_queue')
    def test_tag = unwrap(params,'e2e_image_tag')
    def dataplane_dir = unwrap(params['dataplane_dir'])
    def controlplane_dir = unwrap(params['controlplane_dir'])

    CheckoutDataPlane(params)
    CheckoutControlPlane(params)

    while (artefacts_stash_queue.size() > 0) {
        def stash_name = artefacts_stash_queue.poll()
        unstash name: stash_name
    }

    def dataplane_srcs = "${env.WORKSPACE}/${dataplane_dir}"
    def controlplane_srcs = "${env.WORKSPACE}/${controlplane_dir}"
    def data_dir = "${env.WORKSPACE}/${e2e_artifacts_dir}/coverage/data"
    def bin_dir = "${env.WORKSPACE}/${e2e_artifacts_dir}/binaries/${test_tag}"
    def report_dir = "${env.WORKSPACE}/${e2e_artifacts_dir}/coverage/report"

    sh "./scripts/get-mayastor-binaries.py --tag ${test_tag} --registry ${env.REGISTRY} --outputdir ${bin_dir}"

    sh "./scripts/coverage-report.sh -d ${data_dir} -b ${bin_dir} -M ${dataplane_srcs} -C ${controlplane_srcs} -r ${report_dir}"

    stash_name = 'artefacts-with-coverage-report'
    stash includes: "${e2e_artifacts_dir}/**/**", name: stash_name
    artefacts_stash_queue.add(stash_name)
}

def PrettyPrintMap(hmapname, hmap) {
    def smap = hmap.sort()
    def pretty = "${hmapname}:\n"
    for (elem in smap) {
        pretty += "\t${elem.key}: ${elem.value}\n"
    }
    println(pretty)
}

def RunTestJob(job_params, job_branch) {
   def artifacts_dir = 'artifacts'
   def e2e_artifacts_dir = 'mayastor-e2e/artifacts'

   println(job_params)

   sh """
        rm -rf ${artifacts_dir}
        rm -rf ${e2e_artifacts_dir}
        mkdir -p ${artifacts_dir}
        mkdir -p ${e2e_artifacts_dir}
    """

   def built = build(
       job: "generic-system-test/${job_branch}",
       propagate: false,
       wait: true,
       parameters: job_params
   )

   def base_url = built.getAbsoluteUrl()
   withCredentials([
     usernamePassword(credentialsId: 'JENKINS_AUTH_NO_CHALLENGE', usernameVariable: 'JANC_US', passwordVariable: 'JANC_PW'),
     string(credentialsId: 'HCLOUD_TOKEN', variable: 'HCLOUD_TOKEN')
   ]) {
        for (int ix=0; ix < 3; ix++) {
            try {
                def console_url = base_url + "consoleText"
                sh """
                    cd ${artifacts_dir}
                    wget --auth-no-challenge --user $JANC_US --password=$JANC_PW '${console_url}' -O consoleText
                """
                break
            } catch(err) {
                println("Get consoleText failed.")
            }
            sleep 30
        }
        try {
             def console_url = base_url + "consoleFull"
             sh """
                cd ${artifacts_dir}
                wget --auth-no-challenge --user $JANC_US --password=$JANC_PW '${console_url}' -O consoleFull
                tar czf consoleFull.tgz consoleFull
                rm consoleFull
                """
         } catch(err) {
             println("Get consoleFull failed.")
         }
   }

   try {
       copyArtifacts(
           projectName: "${built.getFullProjectName()}",
           selector: specific("${built.getNumber()}"),
           fingerprintArtifacts: true,
           target: "${env.WORKSPACE}"
       )
   } catch(err) {
       println("Failed to copy artificats ${err}")
   }

   archiveArtifacts    artifacts: "${artifacts_dir}/**/**, ${e2e_artifacts_dir}/**/**",
       fingerprint: true,
       allowEmptyArchive: true

   try {
       def testlist = readFile(file: "${artifacts_dir}/tests.txt")
       testlist.trim()

       if (testlist != "[]") {
           // Processing passed_tests.txt is not critical
           try {
               def passed = readFile(file: "${artifacts_dir}/passed_tests.txt")
               passed.trim()
               println("Passed: ${passed}")
           } catch(err) {
               println("Failed to read ${artifacts_dir}/passed_tests.txt : $err")
           }

           // Must be able to process failed.txt for visual consistency
           def failed = readFile(file: "${artifacts_dir}/failed_tests.txt")
           failed.trim()
           if (failed != "[]") {
               error("Failed tests: ${failed}")
           }
        }
   } catch(err) {
       println("Failed to read ${artifacts_dir}/tests.txt : $err")
   }

   if (built.getCurrentResult() != "SUCCESS") {
        error(built.getCurrentResult())
   }
}


return this
