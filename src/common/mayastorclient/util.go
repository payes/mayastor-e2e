package mayastorclient

import (
	"context"
	"fmt"
	mayastorGrpc "mayastor-e2e/common/mayastorclient/grpc"
	"time"

	"google.golang.org/grpc"
	grpcCodes "google.golang.org/grpc/codes"
	grpcStatus "google.golang.org/grpc/status"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
)

func isDeadlineExceeded(err error) bool {
	status, ok := grpcStatus.FromError(err)
	if !ok {
		return false
	}
	if status.Code() == grpcCodes.DeadlineExceeded {
		return true
	}
	return false
}

func niceError(err error) error {
	if err != nil {
		if isDeadlineExceeded(err) {
			// stop huge print out of error on deadline exceeded
			return grpcStatus.Error(grpcCodes.DeadlineExceeded, fmt.Sprintf("%v", context.DeadlineExceeded))
		}
	}
	return err
}

var canConnect bool = false

func mayastorInfo(address string) (*mayastorGrpc.MayastorInfoRequest, error) {
	var err error
	addrPort := fmt.Sprintf("%s:%d", address, mayastorPort)
	conn, err := grpc.Dial(addrPort, grpc.WithInsecure())
	if err != nil {
		return nil, err
	}
	defer func(conn *grpc.ClientConn) {
		_ = conn.Close()
	}(conn)

	c := mayastorGrpc.NewMayastorClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	info, err := c.GetMayastorInfo(ctx, &null)
	return info, err
}

// CheckAndSetConnect call to cache connectable state to Mayastor instances on the cluster under test
// Just dialing does not work, we need to make a simple gRPC call  (GetMayastorInfo)
func CheckAndSetConnect(nodes []string) {
	allConnected := true
	logf.Log.Info("Checking gRPC connections to Mayastor on", "nodes", nodes)
	if len(nodes) != 0 {
		for _, node := range nodes {
			info, err := mayastorInfo(node)
			if err != nil || info == nil {
				allConnected = false
				logf.Log.Info("gRPC connect failed", "node", node, "err", err)
				break
			}
		}
		canConnect = allConnected
		logf.Log.Info("gRPC connect to Mayastor instances", "enabled", canConnect)
	}
}

// CanConnect retrieve the cached connectable state to Mayastor instances on the cluster under test
func CanConnect() bool {
	return canConnect
}
