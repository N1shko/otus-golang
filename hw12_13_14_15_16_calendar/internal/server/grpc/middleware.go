package internalgrpc

import (
	"context"
	"fmt"
	"net"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/peer"
	"google.golang.org/grpc/status"
)

func (s *Server) LoggingUnaryInterceptor(
	ctx context.Context,
	req any,
	info *grpc.UnaryServerInfo,
	handler grpc.UnaryHandler,
) (any, error) {
	start := time.Now()

	clientIP := ""
	if p, ok := peer.FromContext(ctx); ok {
		host, _, err := net.SplitHostPort(p.Addr.String())
		if err != nil {
			clientIP = p.Addr.String()
		} else {
			clientIP = host
		}
	}

	resp, err := handler(ctx, req)

	latency := fmt.Sprintf("%d", time.Since(start).Milliseconds())
	date := time.Now().UTC().Format("[02/Jan/2006:15:04:05 -0700]")
	userAgent := `""`
	if md, ok := metadata.FromIncomingContext(ctx); ok {
		if ua, ok := md["user-agent"]; ok && len(ua) > 0 {
			userAgent = fmt.Sprintf("%q", ua[0])
		}
	}
	code := codes.OK
	if err != nil {
		if s, ok := status.FromError(err); ok {
			code = s.Code()
		}
	}
	s.logger.Info(
		"Request",
		"client_ip", clientIP,
		"date", date,
		"method", info.FullMethod,
		"proto", "grpc",
		"code", code,
		"latency", latency,
		"user_agent", userAgent,
	)

	return resp, err
}
