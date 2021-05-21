package k8stest

import (
	"context"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

// getRegStr adjust the registry string if necessary to generate the correct image path
// we need to generate 2 types of image strings
//	1) no registry specified, image is pulled from dockerhub
//		- "mayadata/xxxx:tag"
//	2) test registry is specified, image is pulled from there
//		- "ci-registry.mayastor-ci.mayadata.io/mayadata/xxxx:tag"
// adjust the registry by adding the "/" separator to end of test registry
func getRegStr(registry string) string {
	if registry != "" {
		return registry + "/"
	}
	return ""
}

// MayastorDsPatch is used to patch mayastor daemonset
// registry from where mayastor images are retrieved
// imageTag will be the type of image ci or develop, etc
// namespace is the workspace where mayastor pods are deployed
func MayastorDsPatch(registry string, imageTag string, namespace string) error {
	strRegistry := getRegStr(registry)
	daemonsets := gTestEnv.KubeInt.AppsV1().DaemonSets(namespace)
	patch := []byte(`{ "spec": { "template": { "spec": { "containers":  [{"name": "mayastor","image":` + ` "` + strRegistry + `mayadata/mayastor:` + imageTag + `","args": ["-l3","-N$(MY_NODE_NAME)","-g$(MY_POD_IP)","-nnats:4222","-y/var/local/mayastor/config.yaml"]}]}}}}`)
	_, err := daemonsets.Patch(context.TODO(), "mayastor", types.StrategicMergePatchType, patch, metav1.PatchOptions{})
	if err != nil {
		return err
	}
	return nil
}

// MayastorMoacPatch is used to patch moac deployment
// registry from where mayastor images are retrieved
// imageTag will be the type of image ci or develop, etc
// namespace is the workspace where mayastor pods are deployed
func MayastorMoacPatch(registry string, imageTag string, namespace string) error {
	strRegistry := getRegStr(registry)
	deployments := gTestEnv.KubeInt.AppsV1().Deployments(namespace)
	patch := []byte(`{ "spec": { "template": { "spec": { "containers": [{"name":"moac","image":` + ` "` + strRegistry + `mayadata/moac:` + imageTag + `"}]}}}}`)
	_, err := deployments.Patch(context.TODO(), "moac", types.StrategicMergePatchType, patch, metav1.PatchOptions{})
	if err != nil {
		return err
	}
	return nil
}

// MayastorCsiPatch is used to patch mayastor-csi daemonset
// registry from where mayastor images are retrieved
// imageTag will be the type of image ci or develop, etc
// namespace is the workspace where mayastor pods are deployed
func MayastorCsiPatch(registry string, imageTag string, namespace string) error {
	strRegistry := getRegStr(registry)
	daemonsets := gTestEnv.KubeInt.AppsV1().DaemonSets(namespace)
	patch := []byte(`{ "spec": { "template": { "spec": { "containers": [{"name": "mayastor-csi","image":` + ` "` + strRegistry + `mayadata/mayastor-csi:` + imageTag + `"}]}}}}`)
	_, err := daemonsets.Patch(context.TODO(), "mayastor-csi", types.StrategicMergePatchType, patch, metav1.PatchOptions{})
	if err != nil {
		return err
	}
	return nil
}
