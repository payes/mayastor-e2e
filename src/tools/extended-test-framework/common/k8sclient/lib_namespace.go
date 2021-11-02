package k8sclient

import (
	"context"
	"fmt"

	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	coreV1 "k8s.io/api/core/v1"
)

// CreateNamespace create the given namespace
func CreateNamespace(namespace string) error {
	nsSpec := &coreV1.Namespace{ObjectMeta: metaV1.ObjectMeta{Name: namespace}}
	_, err := gKubeInt.CoreV1().Namespaces().Create(context.TODO(), nsSpec, metaV1.CreateOptions{})
	if err != nil {
		return fmt.Errorf("failed to create namespace, err: %v", err)
	}
	return nil
}
