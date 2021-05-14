package mayastorclient

import (
	"context"
	"fmt"
	"time"

	mayastorGrpc "mayastor-e2e/common/mayastorclient/grpc"

	"google.golang.org/grpc"

	logf "sigs.k8s.io/controller-runtime/pkg/log"
)

//TODO: change enum fields to strings?

// MayastorPool Mayastor Pool data
type MayastorPool struct {
	Name     string                 `protobuf:"bytes,1,opt,name=name,proto3" json:"name,omitempty"`
	Disks    []string               `protobuf:"bytes,2,rep,name=disks,proto3" json:"disks,omitempty"`
	State    mayastorGrpc.PoolState `protobuf:"varint,3,opt,name=state,proto3,enum=mayastor.PoolState" json:"state,omitempty"`
	Capacity uint64                 `protobuf:"varint,5,opt,name=capacity,proto3" json:"capacity,omitempty"`
	Used     uint64                 `protobuf:"varint,6,opt,name=used,proto3" json:"used,omitempty"`
}

func (msp MayastorPool) String() string {
	return fmt.Sprintf("Name=%s; Disks=%v; State=%v; Used=%d, Capacity=%d;",
		msp.Name, msp.Disks, msp.State, msp.Used, msp.Capacity)
}

func listPool(address string) ([]MayastorPool, error) {
	var poolInfos []MayastorPool
	var err error
	addrPort := fmt.Sprintf("%s:%d", address, mayastorPort)
	conn, err := grpc.Dial(addrPort, grpc.WithInsecure())
	if err != nil {
		logf.Log.Info("listPool", "error", err)
		return poolInfos, err
	}
	c := mayastorGrpc.NewMayastorClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	response, err := c.ListPools(ctx, &null)
	if err == nil {
		if response != nil {
			for _, pool := range response.Pools {
				pi := MayastorPool{
					Name:     pool.Name,
					Disks:    pool.Disks,
					State:    pool.State,
					Capacity: pool.Capacity,
					Used:     pool.Used,
				}
				poolInfos = append(poolInfos, pi)
			}
		} else {
			err = fmt.Errorf("nil response for ListPools on %s", address)
			logf.Log.Info("listPool", "error", err)
		}
	} else {
		logf.Log.Info("listPool", "error", err)
	}
	closeErr := conn.Close()
	if closeErr != nil {
		logf.Log.Info("listPool", "error on close ", closeErr)
	}
	return poolInfos, err
}

// ListPools given a list of node ip addresses, enumerate the set of pools on mayastor using gRPC on each of those nodes
// returns accumulated errors if gRPC communication failed.
func ListPools(addrs []string) ([]MayastorPool, error) {
	var accErr error
	var poolInfos []MayastorPool
	for _, address := range addrs {
		poolInfo, err := listPool(address)
		if err == nil {
			poolInfos = append(poolInfos, poolInfo...)
		} else {
			if accErr != nil {
				accErr = fmt.Errorf("%v;%v", accErr, err)
			} else {
				accErr = err
			}
		}
	}
	return poolInfos, accErr
}
