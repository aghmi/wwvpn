package main

import (
	"flag"
	"log"
	"net"

	"google.golang.org/grpc"
	"github.com/wwvpn/wwvpn/internal/node"
	"github.com/wwvpn/wwvpn/internal/node/pb"
)

func main() {
	port := flag.String("port", "50051", "gRPC listen port")
	awgInterface := flag.String("awg-interface", "awg0", "AmneziaWG interface name")
	awgPort := flag.Int("awg-port", 51820, "AmneziaWG listen port")
	awgPrivKey := flag.String("awg-privkey", "", "AmneziaWG private key")
	awgPubKey := flag.String("awg-pubkey", "", "AmneziaWG public key")
	jc := flag.Int("jc", 5, "AmneziaWG Jc")
	jmin := flag.Int("jmin", 50, "AmneziaWG Jmin")
	jmax := flag.Int("jmax", 1000, "AmneziaWG Jmax")
	s1 := flag.Int("s1", 50, "AmneziaWG S1")
	s2 := flag.Int("s2", 50, "AmneziaWG S2")
	h1 := flag.Int("h1", 1234567890, "AmneziaWG H1")
	h2 := flag.Int("h2", 987654321, "AmneziaWG H2")
	h3 := flag.Int("h3", 1122334455, "AmneziaWG H3")
	h4 := flag.Int("h4", 5544332211, "AmneziaWG H4")
	xrayPubKey := flag.String("xray-pubkey", "", "XRay Reality public key")
	xrayShortID := flag.String("xray-shortid", "", "XRay Reality short ID")
	xraySNI := flag.String("xray-sni", "www.microsoft.com", "XRay Reality SNI")
	xrayPort := flag.Int("xray-port", 443, "XRay listen port")
	flag.Parse()

	_ = awgPrivKey

	awgConfig := node.AWGConfig{
		Interface:  *awgInterface,
		ListenPort: *awgPort,
		PublicKey:  *awgPubKey,
		Jc:         *jc,
		Jmin:       *jmin,
		Jmax:       *jmax,
		S1:         *s1,
		S2:         *s2,
		H1:         *h1,
		H2:         *h2,
		H3:         *h3,
		H4:         *h4,
	}

	xrayConfig := node.XRayConfig{
		PublicKey: *xrayPubKey,
		ShortID:  *xrayShortID,
		SNI:      *xraySNI,
		Port:     *xrayPort,
	}

	awgManager := node.NewAWGManager(awgConfig)
	xrayManager := node.NewXRayManager(xrayConfig)
	agentServer := node.NewAgentServer(awgManager, xrayManager)

	lis, err := net.Listen("tcp", ":"+*port)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	grpcServer := grpc.NewServer()
	pb.RegisterNodeAgentServer(grpcServer, agentServer)

	log.Printf("node-agent listening on :%s", *port)
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
