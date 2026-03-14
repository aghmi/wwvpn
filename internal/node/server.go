package node

import (
	"context"

	"github.com/wwvpn/wwvpn/internal/node/pb"
)

const version = "0.1.0"

type AgentServer struct {
	pb.UnimplementedNodeAgentServer
	awg  *AWGManager
	xray *XRayManager
}

func NewAgentServer(awg *AWGManager, xray *XRayManager) *AgentServer {
	return &AgentServer{awg: awg, xray: xray}
}

func (s *AgentServer) AddPeer(ctx context.Context, req *pb.AddPeerRequest) (*pb.AddPeerResponse, error) {
	resp := &pb.AddPeerResponse{}

	switch req.Protocol {
	case "amnezia_wg":
		assignedIP, err := s.awg.AddPeer(req.ClientPublicKey)
		if err != nil {
			return nil, err
		}

		resp.ServerPublicKey = s.awg.PublicKey()
		resp.Endpoint = s.awg.Endpoint()
		resp.AllowedIps = "0.0.0.0/0"
		resp.Dns = "1.1.1.1,8.8.8.8"
		resp.AssignedIp = assignedIP + "/32"

		jc, jmin, jmax, s1, s2, h1, h2, h3, h4 := s.awg.ObfuscationParams()
		resp.Jc = int32(jc)
		resp.Jmin = int32(jmin)
		resp.Jmax = int32(jmax)
		resp.S1 = int32(s1)
		resp.S2 = int32(s2)
		resp.H1 = int32(h1)
		resp.H2 = int32(h2)
		resp.H3 = int32(h3)
		resp.H4 = int32(h4)

	case "vless_reality":
		vlessUUID, err := s.xray.AddUser(req.DeviceUuid)
		if err != nil {
			return nil, err
		}

		pubKey, shortID, sni, port := s.xray.Params()
		resp.VlessUuid = vlessUUID
		resp.VlessPublicKey = pubKey
		resp.VlessShortId = shortID
		resp.VlessSni = sni
		resp.VlessPort = int32(port)
	}

	return resp, nil
}

func (s *AgentServer) RemovePeer(ctx context.Context, req *pb.RemovePeerRequest) (*pb.RemovePeerResponse, error) {
	switch req.Protocol {
	case "amnezia_wg":
		if err := s.awg.RemovePeer(req.ClientPublicKey); err != nil {
			return &pb.RemovePeerResponse{Success: false}, err
		}
	case "vless_reality":
		if err := s.xray.RemoveUser(req.VlessUuid); err != nil {
			return &pb.RemovePeerResponse{Success: false}, err
		}
	}

	return &pb.RemovePeerResponse{Success: true}, nil
}

func (s *AgentServer) ListPeers(ctx context.Context, req *pb.ListPeersRequest) (*pb.ListPeersResponse, error) {
	count := s.awg.PeerCount() + s.xray.UserCount()
	return &pb.ListPeersResponse{ActiveCount: int32(count)}, nil
}

func (s *AgentServer) HealthCheck(ctx context.Context, req *pb.HealthRequest) (*pb.HealthResponse, error) {
	return &pb.HealthResponse{
		Healthy: true,
		Version: version,
	}, nil
}
