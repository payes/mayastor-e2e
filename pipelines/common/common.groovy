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
def LokiInstall(tag, loki_run_id) {
  sh 'kubectl apply -f ./loki/promtail_namespace_e2e.yaml'
  sh 'kubectl apply -f ./loki/promtail_rbac_e2e.yaml'
  sh 'kubectl apply -f ./loki/promtail_configmap_e2e.yaml'
  def cmd = "run=\"${loki_run_id}\" version=\"${tag}\" envsubst -no-unset < ./loki/promtail_daemonset_e2e.template.yaml | kubectl apply -f -"
  sh "nix-shell --run '${cmd}'"
}

// Unnstall Loki
def LokiUninstall(tag, loki_run_id) {
  def cmd = "run=\"${loki_run_id}\" version=\"${tag}\" envsubst -no-unset < ./loki/promtail_daemonset_e2e.template.yaml | kubectl delete -f -"
  sh "nix-shell --run '${cmd}'"
  sh 'kubectl delete -f ./loki/promtail_configmap_e2e.yaml'
  sh 'kubectl delete -f ./loki/promtail_rbac_e2e.yaml'
  sh 'kubectl delete -f ./loki/promtail_namespace_e2e.yaml'
}

return this