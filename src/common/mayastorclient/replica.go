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
type MayastorReplica struct {
	Uuid  string                            `protobuf:"bytes,1,opt,name=uuid,proto3" json:"uuid,omitempty"`                                       // uuid of the replica
	Pool  string                            `protobuf:"bytes,2,opt,name=pool,proto3" json:"pool,omitempty"`                                       // name of the pool
	Thin  bool                              `protobuf:"varint,3,opt,name=thin,proto3" json:"thin,omitempty"`                                      // thin provisioning
	Size  uint64                            `protobuf:"varint,4,opt,name=size,proto3" json:"size,omitempty"`                                      // size of the replica in bytes
	Share mayastorGrpc.ShareProtocolReplica `protobuf:"varint,5,opt,name=share,proto3,enum=mayastor.ShareProtocolReplica" json:"share,omitempty"` // protocol used for exposing the replica
	Uri   string                            `protobuf:"bytes,6,opt,name=uri,proto3" json:"uri,omitempty"`                                         // uri usable by nexus to access it
}

type MayastorReplicaArray []MayastorReplica

func (msr MayastorReplicaArray) Len() int           { return len(msr) }
func (msr MayastorReplicaArray) Less(i, j int) bool { return msr[i].Uuid < msr[j].Uuid }
func (msr MayastorReplicaArray) Swap(i, j int)      { msr[i], msr[j] = msr[j], msr[i] }

func (msr MayastorReplica) String() string {
	return fmt.Sprintf("Uuid=%s; Pool=%s; Thin=%v; Size=%d; Share=%s; Uri=%s;",
		msr.Uuid, msr.Pool, msr.Thin, msr.Size, msr.Share, msr.Uri)
}

func listReplica(address string) ([]MayastorReplica, error) {
	var replicaInfos []MayastorReplica
	var err error
	addrPort := fmt.Sprintf("%s:%d", address, mayastorPort)
	conn, err := grpc.Dial(addrPort, grpc.WithInsecure())
	if err != nil {
		logf.Log.Info("listReplica", "error", err)
		return replicaInfos, err
	}
	c := mayastorGrpc.NewMayastorClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	response, err := c.ListReplicas(ctx, &null)
	if err == nil {
		if response != nil {
			for _, replica := range response.Replicas {
				ri := MayastorReplica{
					Uuid:  replica.Uuid,
					Pool:  replica.Pool,
					Thin:  replica.Thin,
					Size:  replica.Size,
					Share: replica.Share,
					Uri:   replica.Uri,
				}
				replicaInfos = append(replicaInfos, ri)
			}
		} else {
			err = fmt.Errorf("nil response for ListReplicas on %s", address)
			logf.Log.Info("listReplicas", "error", err)
		}
	} else {
		logf.Log.Info("listReplicas", "error", err)
	}
	closeErr := conn.Close()
	if closeErr != nil {
		logf.Log.Info("listReplicas", "error on close", closeErr)
	}
	return replicaInfos, err
}

// ListReplicas given a list of node ip addresses, enumerate the set of replicas on mayastor using gRPC on each of those nodes
// returns accumulated errors if gRPC communication failed.
func ListReplicas(addrs []string) ([]MayastorReplica, error) {
	var accErr error
	var replicaInfos []MayastorReplica
	for _, address := range addrs {
		replicaInfo, err := listReplica(address)
		if err == nil {
			replicaInfos = append(replicaInfos, replicaInfo...)
		} else {
			if accErr != nil {
				accErr = fmt.Errorf("%v;%v", accErr, err)
			} else {
				accErr = err
			}
		}
	}
	return replicaInfos, accErr
}
