package mayastorclient

import (
	"context"
	"errors"
	"fmt"
	"time"

	mayastorGrpc "mayastor-e2e/common/mayastorclient/grpc"

	"google.golang.org/grpc"
	"k8s.io/apimachinery/pkg/runtime/schema"

	k8serrors "k8s.io/apimachinery/pkg/api/errors"
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
	defer func(conn *grpc.ClientConn) {
		err := conn.Close()
		if err != nil {
			logf.Log.Info("listPool", "error on close", err)
		}
	}(conn)

	c := mayastorGrpc.NewMayastorClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var response *mayastorGrpc.ListPoolsReply
	for ito := 0; ito < len(backOffTimes); ito += 1 {
		response, err = c.ListPools(ctx, &null)
		if !errors.Is(err, context.DeadlineExceeded) {
			break
		}
		time.Sleep(backOffTimes[ito])
	}

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
		err = niceError(err)
		logf.Log.Info("listPool", "error", err)
	}
	return poolInfos, err
}

func GetPool(name, addr string) (*MayastorPool, error) {
	poolInfo, err := listPool(addr)
	if err != nil {
		return nil, err
	}
	for _, pool := range poolInfo {
		if pool.Name == name {
			return &pool, nil
		}
	}
	return nil, k8serrors.NewNotFound(schema.GroupResource{}, "")
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
