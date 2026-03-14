package node

import (
	"context"

	"github.com/wwvpn/wwvpn/internal/node/pb"
)

type NodeClient interface {
	AddPeer(ctx context.Context, protocol, clientPubKey, deviceUUID string) (*pb.AddPeerResponse, error)
	RemovePeer(ctx context.Context, protocol, clientPubKey, vlessUUID string) error
	ListPeers(ctx context.Context) (int32, error)
	HealthCheck(ctx context.Context) (bool, string, error)
	Close() error
}

type NodeClientFactory func(address string) (NodeClient, error)

func DefaultClientFactory(address string) (NodeClient, error) {
	return NewClient(address)
}
