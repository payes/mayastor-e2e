package mayastorclient

import (
	"context"
	"fmt"
	"time"

	mayastorGrpc "mayastor-e2e/common/mayastorclient/grpc"

	"google.golang.org/grpc"

	logf "sigs.k8s.io/controller-runtime/pkg/log"
)

// MayastorNexus Mayastor Nexus data
type MayastorNexus struct {
	Uuid      string                  `protobuf:"bytes,1,opt,name=uuid,proto3" json:"uuid,omitempty"`
	Size      uint64                  `protobuf:"varint,2,opt,name=size,proto3" json:"size,omitempty"`
	State     mayastorGrpc.NexusState `protobuf:"varint,3,opt,name=state,proto3,enum=mayastor.NexusState" json:"state,omitempty"`
	Children  []*mayastorGrpc.Child   `protobuf:"bytes,4,rep,name=children,proto3" json:"children,omitempty"`
	DeviceUri string                  `protobuf:"bytes,5,opt,name=device_uri,json=deviceUri,proto3" json:"device_uri,omitempty"`
	Rebuilds  uint32                  `protobuf:"varint,6,opt,name=rebuilds,proto3" json:"rebuilds,omitempty"`
}

func (msn MayastorNexus) String() string {
	descChildren := "["
	for _, child := range msn.Children {
		descChildren = fmt.Sprintf("%s(%v); ", descChildren, child)
	}
	descChildren += "]"
	return fmt.Sprintf("Uuid=%s; Size=%d; State=%v; DeviceUri=%s, Rebuilds=%d; Children=%v",
		msn.Uuid, msn.Size, msn.State, msn.DeviceUri, msn.Rebuilds, descChildren)
}

func listNexuses(address string) ([]MayastorNexus, error) {
	var nexusInfos []MayastorNexus
	var err error

	addrPort := fmt.Sprintf("%s:%d", address, mayastorPort)
	conn, err := grpc.Dial(addrPort, grpc.WithInsecure())
	if err != nil {
		logf.Log.Info("listNexuses", "error", err)
		return nexusInfos, err
	}
	defer func(conn *grpc.ClientConn) {
		err := conn.Close()
		if err != nil {
			logf.Log.Info("ListPools", "error on close", err)
		}
	}(conn)

	c := mayastorGrpc.NewMayastorClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	response, err := c.ListNexus(ctx, &null)
	if err == nil {
		if response != nil {
			for _, nexus := range response.NexusList {
				ni := MayastorNexus{
					Uuid:      nexus.Uuid,
					Size:      nexus.Size,
					State:     nexus.State,
					Children:  nexus.Children,
					DeviceUri: nexus.DeviceUri,
					Rebuilds:  nexus.Rebuilds,
				}
				nexusInfos = append(nexusInfos, ni)
			}
		} else {
			err = fmt.Errorf("nil response for ListNexus on %s", address)
			logf.Log.Info("ListPools", "error", err)
		}
	} else {
		logf.Log.Info("ListPools", "error", err)
	}
	return nexusInfos, err
}

// ListNexuses given a list of node ip addresses, enumerate the set of nexuses on mayastor using gRPC on each of those nodes
// returns accumulated errors if gRPC communication failed.
func ListNexuses(addrs []string) ([]MayastorNexus, error) {
	var accErr error
	var nexusInfos []MayastorNexus
	for _, address := range addrs {
		nexusInfo, err := listNexuses(address)
		if err == nil {
			nexusInfos = append(nexusInfos, nexusInfo...)
		} else {
			if accErr != nil {
				accErr = fmt.Errorf("%v;%v", accErr, err)
			} else {
				accErr = err
			}
		}
	}
	return nexusInfos, accErr
}

func FaultNexusChild(address string, Uuid string, Uri string) error {
	var err error
	addrPort := fmt.Sprintf("%s:%d", address, mayastorPort)
	conn, err := grpc.Dial(addrPort, grpc.WithInsecure())
	if err != nil {
		logf.Log.Info("FaultNexusChild", "error", err)
		return err
	}
	defer func(conn *grpc.ClientConn) {
		err := conn.Close()
		if err != nil {
			logf.Log.Info("FaultNexusChild", "error on close", err)
		}
	}(conn)

	c := mayastorGrpc.NewMayastorClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	faultRequest := mayastorGrpc.FaultNexusChildRequest{
		Uuid: Uuid,
		Uri:  Uri,
	}
	response, err := c.FaultNexusChild(ctx, &faultRequest)
	if err == nil {
		if response == nil {
			err = fmt.Errorf("nil response to FaultNexusChild")
		}
	} else {
		logf.Log.Info("FaultNexusChild", "error", err)
	}

	return err
}
