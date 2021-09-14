package lib

import (
	"context"
	"fmt"

	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"

	coreV1 "k8s.io/api/core/v1"
)

// CreateNamespace create the given namespace
func CreateNamespace(clientset kubernetes.Clientset, namespace string) error {
	nsSpec := &coreV1.Namespace{ObjectMeta: metaV1.ObjectMeta{Name: namespace}}
	_, err := clientset.CoreV1().Namespaces().Create(context.TODO(), nsSpec, metaV1.CreateOptions{})
	if err != nil {
		fmt.Println(err)
		return err
	}
	return nil
}
