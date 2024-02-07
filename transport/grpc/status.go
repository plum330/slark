package grpc

import (
	"go.opentelemetry.io/otel/codes"
	grpccodes "google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func statusCode(status *status.Status) (codes.Code, string) {
	switch status.Code() {
	case grpccodes.Unknown,
		grpccodes.DeadlineExceeded,
		grpccodes.Unimplemented,
		grpccodes.Internal,
		grpccodes.Unavailable,
		grpccodes.DataLoss:
		return codes.Error, status.Message()
	default:
		return codes.Unset, ""
	}
}
