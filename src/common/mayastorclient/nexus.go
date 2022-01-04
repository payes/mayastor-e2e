package mayastorclient

import (
	"context"
	"fmt"
	mayastorGrpc "mayastor-e2e/common/mayastorclient/grpc"

	"google.golang.org/grpc"

	logf "sigs.k8s.io/controller-runtime/pkg/log"
)

// MayastorNexus Mayastor Nexus data
type MayastorNexus struct {
	Name      string                  `protobuf:"bytes,1,opt,name=name,proto3" json:"name,omitempty"`
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
			logf.Log.Info("ListNexuses", "error on close", err)
		}
	}(conn)

	c := mayastorGrpc.NewMayastorClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), ctxTimeout)
	defer cancel()

	var response *mayastorGrpc.ListNexusV2Reply
	retryBackoff(func() error {
		response, err = c.ListNexusV2(ctx, &null)
		return err
	})

	if err == nil {
		if response != nil {
			for _, nexus := range response.NexusList {
				ni := MayastorNexus{
					Name:      nexus.Name,
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
			logf.Log.Info("ListNexuses", "error", err)
		}
	} else {
		err = niceError(err)
		logf.Log.Info("ListNexuses", "error", err)
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
	ctx, cancel := context.WithTimeout(context.Background(), ctxTimeout)
	defer cancel()

	faultRequest := mayastorGrpc.FaultNexusChildRequest{
		Uuid: Uuid,
		Uri:  Uri,
	}
	var response *mayastorGrpc.Null
	retryBackoff(func() error {
		response, err = c.FaultNexusChild(ctx, &faultRequest)
		return err
	})

	if err == nil {
		if response == nil {
			err = fmt.Errorf("nil response to FaultNexusChild")
		}
	} else {
		err = niceError(err)
		logf.Log.Info("FaultNexusChild", "error", err)
	}

	return err
}

// FindNexus given a list of node ip addresses, return the MayastorNexus with matching uuid
// returns accumulated errors if gRPC communication failed.
func FindNexus(uuid string, addrs []string) (*MayastorNexus, error) {
	var accErr error
	for _, address := range addrs {
		nexusInfos, err := listNexuses(address)
		if err == nil {
			for _, ni := range nexusInfos {
				if ni.Uuid == uuid {
					return &ni, nil
				}
			}
		} else {
			if accErr != nil {
				accErr = fmt.Errorf("%v;%v", accErr, niceError(err))
			} else {
				accErr = err
			}
		}
	}
	return nil, accErr
}
