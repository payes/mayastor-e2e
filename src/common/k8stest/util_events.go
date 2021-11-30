package k8stest

import (
	"context"

	v1 "k8s.io/api/core/v1"
	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// GetEvents retrieves events for a specific namespace
func GetEvents(nameSpace string, listOptions metaV1.ListOptions) (*v1.EventList, error) {
	return gTestEnv.KubeInt.CoreV1().Events(nameSpace).List(context.TODO(), listOptions)
}
