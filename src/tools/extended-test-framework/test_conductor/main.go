package main

import (
	"context"
	"fmt"
	"io/ioutil"
	"strings"

	apps "k8s.io/api/apps/v1"
	core "k8s.io/api/core/v1"
	rbac "k8s.io/api/rbac/v1"

	errors "github.com/pkg/errors"

	"mayastor-e2e/common/custom_resources"

	"mayastor-e2e/tools/extended-test-framework/client"
	"mayastor-e2e/tools/extended-test-framework/client/test_director"

	"mayastor-e2e/tools/extended-test-framework/models"

	"mayastor-e2e/lib"

	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"

	"k8s.io/apimachinery/pkg/runtime/serializer"

	"time"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
)

var gClientSet kubernetes.Clientset
var gDynamicClientSet dynamic.Interface

var nameSpace = "default"

func banner() {
	fmt.Println("test_conductor started")
}

func getPod(podName string, namespace string) (*v1.Pod, error) {
	pods, err := gClientSet.CoreV1().Pods(namespace).List(context.TODO(), metaV1.ListOptions{})
	if err != nil {
		fmt.Println("list failed")
		return nil, err
	}
	for _, pod := range pods.Items {
		if pod.Name == podName {
			fmt.Println("found pod ", podName)
			return &pod, nil
		}
	}
	fmt.Println("not found pod ", podName)
	return nil, errors.New("pod not found")
}

func getTestPlans(client *client.Extended) {
	testPlanParams := test_director.NewGetTestPlansParams()
	pTestPlansOk, err := client.TestDirector.GetTestPlans(testPlanParams)

	if err != nil {
		fmt.Printf("failed to get plans %v %v\n", err, pTestPlansOk)
	} else {
		fmt.Printf("got plans payload %v items %d\n", pTestPlansOk.Payload, len(pTestPlansOk.Payload))
		for _, tp := range pTestPlansOk.Payload {
			fmt.Printf("plan name %s\n", *tp.Name)
			fmt.Printf("plan key %v\n", tp.Key)
		}
	}
}

func sendTestPlan(client *client.Extended, name string, id *models.JiraKey, isActive bool) {

	testPlanRunSpec := models.TestRunSpec{}
	testPlanRunSpec.TestKey = id
	testPlanRunSpec.Data = "test"

	testPlanParams := test_director.NewPutTestPlanByIDParams()
	testPlanParams.ID = string(*id)
	testPlanParams.Body = &testPlanRunSpec

	pPutTestPlansOk, err := client.TestDirector.PutTestPlanByID(testPlanParams)

	if err != nil {
		fmt.Printf("failed to put plans %v %v\n", err, pPutTestPlansOk)
	} else {
		fmt.Printf("put plans payload %v\n", pPutTestPlansOk.Payload)
		fmt.Printf("plan name: %s ", *pPutTestPlansOk.Payload.Name)
		fmt.Printf("plan ID: %s\n", pPutTestPlansOk.Payload.Key)
	}
}

func createNamespace(namespace string) {
	nsSpec := &v1.Namespace{ObjectMeta: metaV1.ObjectMeta{Name: "mayastor"}}
	_, err := gClientSet.CoreV1().Namespaces().Create(context.TODO(), nsSpec, metaV1.CreateOptions{})
	if err != nil {
		fmt.Println(err)
		return
	}
}

func deployYaml(fileName string) {
	b, err := ioutil.ReadFile(fileName)
	if err != nil {
		fmt.Println(err)
		return
	}
	s := string(b)
	fmt.Printf("%q \n\n\nvim ../", s)

	stringSlice := strings.Split(s, "\n---\n")

	scheme := runtime.NewScheme()
	err = apps.AddToScheme(scheme)
	if err != nil {
		fmt.Printf("failed to add apps scheme %v\n\n\n", err)
	}

	err = core.AddToScheme(scheme)
	if err != nil {
		fmt.Printf("failed to add core scheme %v\n\n\n", err)
	}

	err = rbac.AddToScheme(scheme)
	if err != nil {
		fmt.Printf("failed to add rbac scheme %v\n\n\n", err)
	}

	for _, obj_str := range stringSlice {
		fmt.Printf("about to deserialize\n\n\n")
		factory := serializer.NewCodecFactory(scheme)
		decoder := factory.UniversalDeserializer()
		obj, _ /*groupVersionKind*/, err := decoder.Decode([]byte(obj_str), nil, nil)

		if err != nil {
			fmt.Printf("Error while decoding YAML object. Err was: %s\n\n\n", err)
			return
		}

		fmt.Printf("about to switch\n\n\n")
		switch o := obj.(type) {
		case *v1.Pod:
			fmt.Printf("Attempting to create pod\n\n\n")
			// o is a pod

		case *apps.DaemonSet:
			fmt.Printf("Attempting to create daemonset\n\n\n")
			daemonsetClient := gClientSet.AppsV1().DaemonSets("mayastor")
			_, err := daemonsetClient.Create(context.TODO(), o, metaV1.CreateOptions{})
			if err != nil {
				fmt.Printf("Failed to create daemonset. Err was: %s\n\n\n", err)
				return
			}

		case *apps.Deployment:
			fmt.Printf("Attempting to create deployment\n\n\n")
			deploymentClient := gClientSet.AppsV1().Deployments("mayastor")
			_, err := deploymentClient.Create(context.TODO(), o, metaV1.CreateOptions{})
			if err != nil {
				fmt.Printf("Failed to create deployment. Err was: %s\n\n\n", err)
				return
			}

		case *apps.StatefulSet:
			fmt.Printf("Attempting to create stateful set\n\n\n")
			statefulSetClient := gClientSet.AppsV1().StatefulSets("mayastor")
			_, err := statefulSetClient.Create(context.TODO(), o, metaV1.CreateOptions{})
			if err != nil {
				fmt.Printf("Failed to create stateful set. Err was: %s\n\n\n", err)
				return
			}

		case *rbac.Role:
			fmt.Printf("Attempting to create role\n\n\n")
			roleClient := gClientSet.RbacV1().Roles("mayastor")
			_, err := roleClient.Create(context.TODO(), o, metaV1.CreateOptions{})
			if err != nil {
				fmt.Printf("Failed to create role. Err was: %s\n\n\n", err)
				return
			}

		case *rbac.RoleBinding:
			fmt.Printf("Attempting to create role binding\n\n\n")
			roleBindingClient := gClientSet.RbacV1().RoleBindings("mayastor")
			_, err := roleBindingClient.Create(context.TODO(), o, metaV1.CreateOptions{})
			if err != nil {
				fmt.Printf("Failed to create role binding. Err was: %s\n\n\n", err)
				return
			}

		case *rbac.ClusterRole: /**/
			fmt.Printf("Attempting to create cluster role\n\n\n")
			clusterRoleClient := gClientSet.RbacV1().ClusterRoles()
			_, err := clusterRoleClient.Create(context.TODO(), o, metaV1.CreateOptions{})
			if err != nil {
				fmt.Printf("Failed to create cluster role. Err was: %s\n\n\n", err)
				return
			}

		case *rbac.ClusterRoleBinding: /**/
			fmt.Printf("Attempting to create cluster role binding\n\n\n")
			clusterRoleBindingClient := gClientSet.RbacV1().ClusterRoleBindings()
			_, err := clusterRoleBindingClient.Create(context.TODO(), o, metaV1.CreateOptions{})
			if err != nil {
				fmt.Printf("Failed to create cluster role binding. Err was: %s\n\n\n", err)
				return
			}

		case *core.ServiceAccount: /**/
			fmt.Printf("Attempting to create service account\n\n\n")
			serviceAccountClient := gClientSet.CoreV1().ServiceAccounts("mayastor")
			_, err := serviceAccountClient.Create(context.TODO(), o, metaV1.CreateOptions{})
			if err != nil {
				fmt.Printf("Failed to create service account. Err was: %s\n\n\n", err)
				return
			}

		case *core.ConfigMap: /**/
			fmt.Printf("Attempting to create config map\n\n\n")
			configMapClient := gClientSet.CoreV1().ConfigMaps("mayastor")
			_, err := configMapClient.Create(context.TODO(), o, metaV1.CreateOptions{})
			if err != nil {
				fmt.Printf("Failed to create config map. Err was: %s\n\n\n", err)
				return
			}

		case *core.Service: /**/
			fmt.Printf("Attempting to create service\n\n\n")
			serviceClient := gClientSet.CoreV1().Services("mayastor")
			_, err := serviceClient.Create(context.TODO(), o, metaV1.CreateOptions{})
			if err != nil {
				fmt.Printf("Failed to create service. Err was: %s\n\n\n", err)
				return
			}

		default:
			fmt.Printf("object %+v \n\n\n", o)
			//o is unknown for us
		}
	}
}

func waitForPoolCrd() bool {
	const timoSleepSecs = 5
	const timoSecs = 60
	for ix := 0; ix < timoSecs; ix += timoSleepSecs {
		_, err := custom_resources.ListMsPools()
		if err != nil {
			logf.Log.Info("WaitForPoolCrd", "error", err)
		} else {
			return true
		}
	}
	return false
}

func createPools(clientset kubernetes.Clientset) {
	mayastorNodes, err := lib.GetMayastorNodeNames(clientset)
	if err == nil {
		for _, node := range mayastorNodes {
			fmt.Println("node ", node)
		}
	} else {

		return
	}

	numMayastorInstances := len(mayastorNodes)

	logf.Log.Info("Install", "# of mayastor instances", numMayastorInstances)

	waitForPoolCrd()

	for _, node := range mayastorNodes {
		_, err := custom_resources.CreateMsPool(node+"-pool", node, []string{"/dev/sdb"})
		if err != nil {
			fmt.Println(err)
		}
	}
}

func installMayastor(clientset kubernetes.Clientset) {

	createNamespace("mayastor")
	deployYaml("moac-rbac.yaml")
	deployYaml("etcd/statefulset.yaml")
	deployYaml("etcd/svc-headless.yaml")
	deployYaml("etcd/svc.yaml")
	deployYaml("nats-deployment.yaml")
	deployYaml("csi-daemonset.yaml")
	deployYaml("moac-deployment.yaml")
	deployYaml("mayastor-daemonset.yaml")

	_ = gDynamicClientSet

	/*

	   ready, err := k8stest.MayastorReady(2, 540)
	   Expect(err).ToNot(HaveOccurred())
	   Expect(ready).To(BeTrue())

	   crdReady := WaitForPoolCrd()
	   Expect(crdReady).To(BeTrue())

	   // Now create configured pools on all nodes.
	   k8stest.CreateConfiguredPools()

	   // Wait for pools to be online
	   const timoSecs = 120
	*/

	createPools(clientset)
}

func runTest(clientset kubernetes.Clientset) {
	installMayastor(clientset)
	time.Sleep(600 * time.Second)
}

func main() {
	banner()
	restConfig, err := rest.InClusterConfig()
	if err != nil {
		fmt.Println("failed to get cluster config")
		return
	}

	// read config file
	config := GetConfig()

	fmt.Printf("config name %s\n", config.ConfigName)
	if restConfig == nil {
		fmt.Println("failed to get kubeint")
		return
	}
	gClientSet = *kubernetes.NewForConfigOrDie(restConfig)
	gDynamicClientSet = dynamic.NewForConfigOrDie(restConfig)

	time.Sleep(10 * time.Second)

	workloadMonitorPod, err := getPod("workload-monitor", "default")
	if err != nil {
		fmt.Println("failed to get workload-monitor")
		return
	}
	fmt.Println("worload-monitor pod IP is", workloadMonitorPod.Status.PodIP)

	// find the test_director
	testDirectorPod, err := getPod("test-director", nameSpace)
	if err != nil {
		fmt.Println("failed to get test-director pod")
		return
	}

	fmt.Println("test-director pod IP is", testDirectorPod.Status.PodIP)
	testDirectorLoc := testDirectorPod.Status.PodIP + ":8080"

	transportConfig := client.DefaultTransportConfig().WithHost(testDirectorLoc)
	client := client.NewHTTPClientWithConfig(nil, transportConfig)

	var jk models.JiraKey = "MQ-002"
	sendTestPlan(client, "test name 2", &jk, false)

	getTestPlans(client)

	jk = "MQ-003"
	sendTestPlan(client, "test name 3", &jk, true)

	getTestPlans(client)

	runTest(gClientSet)

}
