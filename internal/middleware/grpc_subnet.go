package middleware

import (
	"context"
	"fmt"
	"net"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

func GRPCSubnetInterceptor(subNetMask string) (grpc.UnaryServerInterceptor, error) {
	var trustedNet *net.IPNet
	if subNetMask != "" {
		var err error
		_, trustedNet, err = net.ParseCIDR(subNetMask)
		if err != nil {
			return nil, fmt.Errorf("invalid trusted subnet CIDR %q: %w", subNetMask, err)
		}
	}

	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		if trustedNet == nil {
			return handler(ctx, req)
		}

		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			return nil, status.Error(codes.PermissionDenied, "missing metadata")
		}

		ips := md.Get("x-real-ip")
		if len(ips) == 0 {
			return nil, status.Error(codes.PermissionDenied, "missing x-real-ip header")
		}

		clientIP := net.ParseIP(ips[0])
		if clientIP == nil || !trustedNet.Contains(clientIP) {
			return nil, status.Error(codes.PermissionDenied, "ip is not in trusted subnet")
		}

		return handler(ctx, req)
	}, nil
}
