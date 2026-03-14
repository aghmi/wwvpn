package node

import (
	"context"
	"fmt"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"github.com/wwvpn/wwvpn/internal/node/pb"
)

type Client struct {
	conn   *grpc.ClientConn
	client pb.NodeAgentClient
}

func NewClient(address string) (*Client, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	conn, err := grpc.DialContext(ctx, address,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithBlock(),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to node agent at %s: %w", address, err)
	}

	return &Client{
		conn:   conn,
		client: pb.NewNodeAgentClient(conn),
	}, nil
}

func (c *Client) Close() error {
	return c.conn.Close()
}

func (c *Client) AddPeer(ctx context.Context, protocol, clientPubKey, deviceUUID string) (*pb.AddPeerResponse, error) {
	return c.client.AddPeer(ctx, &pb.AddPeerRequest{
		Protocol:        protocol,
		ClientPublicKey: clientPubKey,
		DeviceUuid:      deviceUUID,
	})
}

func (c *Client) RemovePeer(ctx context.Context, protocol, clientPubKey, vlessUUID string) error {
	_, err := c.client.RemovePeer(ctx, &pb.RemovePeerRequest{
		Protocol:        protocol,
		ClientPublicKey: clientPubKey,
		VlessUuid:       vlessUUID,
	})
	return err
}

func (c *Client) ListPeers(ctx context.Context) (int32, error) {
	resp, err := c.client.ListPeers(ctx, &pb.ListPeersRequest{})
	if err != nil {
		return 0, err
	}
	return resp.ActiveCount, nil
}

func (c *Client) HealthCheck(ctx context.Context) (bool, string, error) {
	resp, err := c.client.HealthCheck(ctx, &pb.HealthRequest{})
	if err != nil {
		return false, "", err
	}
	return resp.Healthy, resp.Version, nil
}
