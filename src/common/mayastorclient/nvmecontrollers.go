package mayastorclient

import (
	"context"
	"fmt"
	"time"

	mayastorGrpc "mayastor-e2e/common/mayastorclient/grpc"

	"google.golang.org/grpc"

	logf "sigs.k8s.io/controller-runtime/pkg/log"
)

// MayastorReplica Mayastor Replica data
type NvmeController struct {
	Name    string                           `protobuf:"bytes,1,opt,name=name,proto3" json:"name,omitempty"`
	State   mayastorGrpc.NvmeControllerState `protobuf:"varint,2,opt,name=state,proto3,enum=mayastor.NvmeControllerState" json:"state,omitempty"`
	Size    uint64                           `protobuf:"varint,3,opt,name=size,proto3" json:"size,omitempty"`
	BlkSize uint32                           `protobuf:"varint,4,opt,name=blk_size,json=blkSize,proto3" json:"blk_size,omitempty"`
}

type NvmeControllerArray []NvmeController

func (msr NvmeController) String() string {
	return fmt.Sprintf("Name=%s; State=%s; ; Size=%d; BlkSize=%d;",
		msr.Name, msr.State, msr.Size, msr.BlkSize)
}

func listNvmeController(address string) ([]NvmeController, error) {
	var nvmeControllers []NvmeController
	var err error
	addrPort := fmt.Sprintf("%s:%d", address, mayastorPort)
	conn, err := grpc.Dial(addrPort, grpc.WithInsecure())
	if err != nil {
		logf.Log.Info("listReplica", "error", err)
		return nvmeControllers, err
	}
	defer func(conn *grpc.ClientConn) {
		err := conn.Close()
		if err != nil {
			logf.Log.Info("listReplicas", "error on close", err)
		}
	}(conn)

	c := mayastorGrpc.NewMayastorClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	response, err := c.ListNvmeControllers(ctx, &null)
	if err == nil {
		if response != nil {
			for _, nvmeController := range response.Controllers {
				nc := NvmeController{
					Name:    nvmeController.Name,
					State:   nvmeController.State,
					Size:    nvmeController.Size,
					BlkSize: nvmeController.BlkSize,
				}
				nvmeControllers = append(nvmeControllers, nc)
			}
		} else {
			err = fmt.Errorf("nil response for ListReplicas on %s", address)
			logf.Log.Info("listReplicas", "error", err)
		}
	} else {
		logf.Log.Info("listReplicas", "error", err)
	}
	return nvmeControllers, err
}

// ListNvmeControllers given a list of node ip addresses, enumerate the set of nvmeControllers on mayastor using gRPC on each of those nodes
// returns accumulated errors if gRPC communication failed.
func ListNvmeControllers(addrs []string) ([]NvmeController, error) {
	var accErr error
	var nvmeControllers []NvmeController
	for _, address := range addrs {
		nvmeController, err := listNvmeController(address)
		if err == nil {
			nvmeControllers = append(nvmeControllers, nvmeController...)
		} else {
			if accErr != nil {
				accErr = fmt.Errorf("%v;%v", accErr, err)
			} else {
				accErr = err
			}
		}
	}
	return nvmeControllers, accErr
}
