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
		return fmt.Errorf("failed to read file %s, error: %v", fileName, err)
	}
	s := string(b)

	stringSlice := strings.Split(s, "\n---\n")

	scheme := runtime.NewScheme()
	err = apps.AddToScheme(scheme)
	if err != nil {
		return fmt.Errorf("failed to add apps scheme, error: %v", err)
	}

	err = core.AddToScheme(scheme)
	if err != nil {
		return fmt.Errorf("failed to add core scheme error: %v", err)
	}

	err = rbac.AddToScheme(scheme)
	if err != nil {
		return fmt.Errorf("failed to add rbac scheme error: %v", err)
	}

	for _, obj_str := range stringSlice {
		factory := serializer.NewCodecFactory(scheme)
		decoder := factory.UniversalDeserializer()
		obj, _, err := decoder.Decode([]byte(obj_str), nil, nil)

		if err != nil {
			return fmt.Errorf("Error while decoding YAML object, error: %v", err)
		}

		switch o := obj.(type) {
		case *apps.DaemonSet:
			daemonsetClient := clientset.AppsV1().DaemonSets("mayastor")
			_, err := daemonsetClient.Create(context.TODO(), o, metaV1.CreateOptions{})
			if err != nil {
				return fmt.Errorf("Failed to create daemonset, error: %v", err)
			}

		case *apps.Deployment:
			deploymentClient := clientset.AppsV1().Deployments("mayastor")
			_, err := deploymentClient.Create(context.TODO(), o, metaV1.CreateOptions{})
			if err != nil {
				return fmt.Errorf("Failed to create deployment, error %v", err)
			}

		case *apps.StatefulSet:
			statefulSetClient := clientset.AppsV1().StatefulSets("mayastor")
			_, err := statefulSetClient.Create(context.TODO(), o, metaV1.CreateOptions{})
			if err != nil {
				return fmt.Errorf("Failed to create stateful set, error: %v", err)
			}

		case *rbac.Role:
			roleClient := clientset.RbacV1().Roles("mayastor")
			_, err := roleClient.Create(context.TODO(), o, metaV1.CreateOptions{})
			if err != nil {
				return fmt.Errorf("Failed to create role, error: %v", err)
			}

		case *rbac.RoleBinding:
			roleBindingClient := clientset.RbacV1().RoleBindings("mayastor")
			_, err := roleBindingClient.Create(context.TODO(), o, metaV1.CreateOptions{})
			if err != nil {
				return fmt.Errorf("Failed to create role binding, error %v", err)
			}

		case *rbac.ClusterRole: /**/
			clusterRoleClient := clientset.RbacV1().ClusterRoles()
			_, err := clusterRoleClient.Create(context.TODO(), o, metaV1.CreateOptions{})
			if err != nil {
				return fmt.Errorf("Failed to create cluster role, error %v", err)
			}

		case *rbac.ClusterRoleBinding:
			clusterRoleBindingClient := clientset.RbacV1().ClusterRoleBindings()
			_, err := clusterRoleBindingClient.Create(context.TODO(), o, metaV1.CreateOptions{})
			if err != nil {
				return fmt.Errorf("Failed to create cluster role binding, error: %v", err)
			}

		case *core.ServiceAccount:
			serviceAccountClient := clientset.CoreV1().ServiceAccounts("mayastor")
			_, err := serviceAccountClient.Create(context.TODO(), o, metaV1.CreateOptions{})
			if err != nil {
				return fmt.Errorf("Failed to create service account, error: %v", err)
			}

		case *core.ConfigMap:
			configMapClient := clientset.CoreV1().ConfigMaps("mayastor")
			_, err := configMapClient.Create(context.TODO(), o, metaV1.CreateOptions{})
			if err != nil {
				return fmt.Errorf("Failed to create config map, error: %v", err)
			}

		case *core.Service:
			serviceClient := clientset.CoreV1().Services("mayastor")
			_, err := serviceClient.Create(context.TODO(), o, metaV1.CreateOptions{})
			if err != nil {
				return fmt.Errorf("Failed to create service, error: %v", err)
			}

		default:
			return fmt.Errorf("Unsupported object %+v", o)
		}
	}
	return nil
}
