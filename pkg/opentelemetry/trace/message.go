package trace

import (
	"context"
	"go.opentelemetry.io/otel/attribute"
	semconv "go.opentelemetry.io/otel/semconv/v1.17.0"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/protobuf/proto"
)

// Semantic conventions for attribute keys for gRPC.
const (
	// RPCNameKey Name of message transmitted or received.
	RPCNameKey = attribute.Key("name")

	// RPCMessageTypeKey Type of message transmitted or received.
	RPCMessageTypeKey = attribute.Key("message.type")

	// RPCMessageIDKey Identifier of message transmitted or received.
	RPCMessageIDKey = attribute.Key("message.id")

	// RPCMessageCompressedSizeKey The compressed size of the message transmitted or received in bytes.
	RPCMessageCompressedSizeKey = attribute.Key("message.compressed_size")

	// RPCMessageUncompressedSizeKey The uncompressed size of the message transmitted or received in bytes.
	RPCMessageUncompressedSizeKey = attribute.Key("message.uncompressed_size")
)

// Semantic conventions for common RPC attributes.
var (
	// RPCSystemGRPC Semantic convention for gRPC as the remoting system.
	RPCSystemGRPC = semconv.RPCSystemGRPC

	// RPCNameMessage Semantic convention for a message named message.
	RPCNameMessage = RPCNameKey.String("message")

	// RPCMessageTypeSent Semantic conventions for RPC message types.
	RPCMessageTypeSent     = RPCMessageTypeKey.String("SENT")
	RPCMessageTypeReceived = RPCMessageTypeKey.String("RECEIVED")
)

var (
	MessageSent     = messageType(RPCMessageTypeSent)
	MessageReceived = messageType(RPCMessageTypeReceived)
)

type messageType attribute.KeyValue

// Event adds an event of the messageType to the span associated with the
// passed context with id and size (if message is a proto message).
func (m messageType) Event(ctx context.Context, id int, message any) {
	span := trace.SpanFromContext(ctx)
	if p, ok := message.(proto.Message); ok {
		span.AddEvent("message", trace.WithAttributes(
			attribute.KeyValue(m),
			RPCMessageIDKey.Int(id),
			RPCMessageUncompressedSizeKey.Int(proto.Size(p)),
		))
	} else {
		span.AddEvent("message", trace.WithAttributes(
			attribute.KeyValue(m),
			RPCMessageIDKey.Int(id),
		))
	}
}
