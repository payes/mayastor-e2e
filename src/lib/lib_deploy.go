package lib

import (
	"context"
	"fmt"
	"io/ioutil"

	"k8s.io/apimachinery/pkg/runtime"

	"strings"

	apps "k8s.io/api/apps/v1"
	core "k8s.io/api/core/v1"
	rbac "k8s.io/api/rbac/v1"

	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"

	"k8s.io/apimachinery/pkg/runtime/serializer"
)

// DeployYaml
func DeployYaml(clientset kubernetes.Clientset, fileName string) error {
	b, err := ioutil.ReadFile(fileName)
	if err != nil {
		fmt.Println(err)
		return err
	}
	s := string(b)
	fmt.Printf("%q \n\n\nvim ../", s)

	stringSlice := strings.Split(s, "\n---\n")

	scheme := runtime.NewScheme()
	err = apps.AddToScheme(scheme)
	if err != nil {
		fmt.Printf("failed to add apps scheme %v\n\n\n", err)
		return err
	}

	err = core.AddToScheme(scheme)
	if err != nil {
		fmt.Printf("failed to add core scheme %v\n\n\n", err)
		return err
	}

	err = rbac.AddToScheme(scheme)
	if err != nil {
		fmt.Printf("failed to add rbac scheme %v\n\n\n", err)
		return err
	}

	for _, obj_str := range stringSlice {
		fmt.Printf("about to deserialize\n\n\n")
		factory := serializer.NewCodecFactory(scheme)
		decoder := factory.UniversalDeserializer()
		obj, _ /*groupVersionKind*/, err := decoder.Decode([]byte(obj_str), nil, nil)

		if err != nil {
			fmt.Printf("Error while decoding YAML object. Err was: %s\n\n\n", err)
			return err
		}

		fmt.Printf("about to switch\n\n\n")
		switch o := obj.(type) {
		case *core.Pod:
			fmt.Printf("Attempting to create pod\n\n\n")
			// o is a pod

		case *apps.DaemonSet:
			fmt.Printf("Attempting to create daemonset\n\n\n")
			daemonsetClient := clientset.AppsV1().DaemonSets("mayastor")
			_, err := daemonsetClient.Create(context.TODO(), o, metaV1.CreateOptions{})
			if err != nil {
				fmt.Printf("Failed to create daemonset. Err was: %s\n\n\n", err)
				return err
			}

		case *apps.Deployment:
			fmt.Printf("Attempting to create deployment\n\n\n")
			deploymentClient := clientset.AppsV1().Deployments("mayastor")
			_, err := deploymentClient.Create(context.TODO(), o, metaV1.CreateOptions{})
			if err != nil {
				fmt.Printf("Failed to create deployment. Err was: %s\n\n\n", err)
				return err
			}

		case *apps.StatefulSet:
			fmt.Printf("Attempting to create stateful set\n\n\n")
			statefulSetClient := clientset.AppsV1().StatefulSets("mayastor")
			_, err := statefulSetClient.Create(context.TODO(), o, metaV1.CreateOptions{})
			if err != nil {
				fmt.Printf("Failed to create stateful set. Err was: %s\n\n\n", err)
				return err
			}

		case *rbac.Role:
			fmt.Printf("Attempting to create role\n\n\n")
			roleClient := clientset.RbacV1().Roles("mayastor")
			_, err := roleClient.Create(context.TODO(), o, metaV1.CreateOptions{})
			if err != nil {
				fmt.Printf("Failed to create role. Err was: %s\n\n\n", err)
				return err
			}

		case *rbac.RoleBinding:
			fmt.Printf("Attempting to create role binding\n\n\n")
			roleBindingClient := clientset.RbacV1().RoleBindings("mayastor")
			_, err := roleBindingClient.Create(context.TODO(), o, metaV1.CreateOptions{})
			if err != nil {
				fmt.Printf("Failed to create role binding. Err was: %s\n\n\n", err)
				return err
			}

		case *rbac.ClusterRole: /**/
			fmt.Printf("Attempting to create cluster role\n\n\n")
			clusterRoleClient := clientset.RbacV1().ClusterRoles()
			_, err := clusterRoleClient.Create(context.TODO(), o, metaV1.CreateOptions{})
			if err != nil {
				fmt.Printf("Failed to create cluster role. Err was: %s\n\n\n", err)
				return err
			}

		case *rbac.ClusterRoleBinding: /**/
			fmt.Printf("Attempting to create cluster role binding\n\n\n")
			clusterRoleBindingClient := clientset.RbacV1().ClusterRoleBindings()
			_, err := clusterRoleBindingClient.Create(context.TODO(), o, metaV1.CreateOptions{})
			if err != nil {
				fmt.Printf("Failed to create cluster role binding. Err was: %s\n\n\n", err)
				return err
			}

		case *core.ServiceAccount: /**/
			fmt.Printf("Attempting to create service account\n\n\n")
			serviceAccountClient := clientset.CoreV1().ServiceAccounts("mayastor")
			_, err := serviceAccountClient.Create(context.TODO(), o, metaV1.CreateOptions{})
			if err != nil {
				fmt.Printf("Failed to create service account. Err was: %s\n\n\n", err)
				return err
			}

		case *core.ConfigMap: /**/
			fmt.Printf("Attempting to create config map\n\n\n")
			configMapClient := clientset.CoreV1().ConfigMaps("mayastor")
			_, err := configMapClient.Create(context.TODO(), o, metaV1.CreateOptions{})
			if err != nil {
				fmt.Printf("Failed to create config map. Err was: %s\n\n\n", err)
				return err
			}

		case *core.Service: /**/
			fmt.Printf("Attempting to create service\n\n\n")
			serviceClient := clientset.CoreV1().Services("mayastor")
			_, err := serviceClient.Create(context.TODO(), o, metaV1.CreateOptions{})
			if err != nil {
				fmt.Printf("Failed to create service. Err was: %s\n\n\n", err)
				return err
			}

		default:
			fmt.Printf("object %+v \n\n\n", o)
			//o is unknown for us
		}
	}
	return nil
}
