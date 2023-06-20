package form

import (
	"encoding/base64"
	"fmt"
	"github.com/go-slark/slark/errors"
	utils "github.com/go-slark/slark/pkg"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/types/known/durationpb"
	"google.golang.org/protobuf/types/known/fieldmaskpb"
	"net/url"
	"strconv"
	"strings"
	"time"

	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/reflect/protoregistry"
	"google.golang.org/protobuf/types/known/structpb"
	"google.golang.org/protobuf/types/known/timestamppb"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

func paddingFieldValues(v protoreflect.Message, paths, values []string) error {
	pLen := len(paths)
	vLen := len(values)
	if pLen == 0 || vLen == 0 {
		return errors.BadRequest("field path or value miss", "FIELD_PATH_OR_VALUE_MISS")
	}

	var fd protoreflect.FieldDescriptor
	for i, field := range paths {
		fd = getFieldDescriptor(v, field)
		if fd == nil {
			return nil
		}

		if i == pLen-1 {
			break
		}

		if fd.Message() == nil || fd.Cardinality() == protoreflect.Repeated {
			if fd.IsMap() && pLen > 1 {
				return paddingMapField(fd, v.Mutable(fd).Map(), []string{paths[1]}, values)
			}
			return fmt.Errorf("path %s is not message", field)
		}

		v = v.Mutable(fd).Message()
	}
	if of := fd.ContainingOneof(); of != nil {
		if f := v.WhichOneof(of); f != nil {
			return fmt.Errorf("field already set for oneof %s", of.FullName().Name())
		}
	}
	switch {
	case fd.IsList():
		return paddingRepeatedField(fd, v.Mutable(fd).List(), values)
	case fd.IsMap():
		return paddingMapField(fd, v.Mutable(fd).Map(), paths, values)
	}
	if vLen > 1 {
		return fmt.Errorf("too many values for field %s: %s", fd.FullName().Name(), strings.Join(values, ", "))
	}
	return paddingField(fd, v, values[0])
}

func getFieldDescriptor(v protoreflect.Message, field string) protoreflect.FieldDescriptor {
	fields := v.Descriptor().Fields()
	fd := getDescriptorByField(fields, field)
	if fd != nil {
		return fd
	}

	if v.Descriptor().FullName() == "google.protobuf.Struct" {
		fd = fields.ByNumber(1)
	} else if len(field) > 2 && strings.HasSuffix(field, "[]") {
		fd = getDescriptorByField(fields, strings.TrimSuffix(field, "[]"))
	} else {
		// TODO
	}
	return fd
}

func getDescriptorByField(fields protoreflect.FieldDescriptors, field string) protoreflect.FieldDescriptor {
	fd := fields.ByName(protoreflect.Name(field))
	if fd == nil {
		fd = fields.ByJSONName(field)
	}
	return fd
}

func paddingField(fd protoreflect.FieldDescriptor, v protoreflect.Message, value string) error {
	if len(value) == 0 {
		return nil
	}
	val, err := parseField(fd, value)
	if err != nil {
		return err
	}
	v.Set(fd, val)
	return nil
}

func paddingRepeatedField(fd protoreflect.FieldDescriptor, list protoreflect.List, values []string) error {
	for _, value := range values {
		v, err := parseField(fd, value)
		if err != nil {
			return err
		}
		list.Append(v)
	}
	return nil
}

func paddingMapField(fd protoreflect.FieldDescriptor, mp protoreflect.Map, paths, values []string) error {
	l := len(paths)
	if l == 0 {
		return errors.BadRequest("paths invalid", "PATHS_INVALID")
	}
	key, err := parseField(fd.MapKey(), paths[l-1])
	if err != nil {
		return err
	}
	l = len(values)
	if l == 0 {
		return errors.BadRequest("values invalid", "VALUES_INVALID")
	}
	value, err := parseField(fd.MapValue(), values[l-1])
	if err != nil {
		return err
	}
	mp.Set(key.MapKey(), value)
	return nil
}

func parse(msg proto.Message, values url.Values) error {
	var err error
	for key, value := range values {
		err = paddingFieldValues(msg.ProtoReflect(), strings.Split(key, "."), value)
		if err != nil {
			return err
		}
	}
	return err
}

func parseField(fd protoreflect.FieldDescriptor, value string) (protoreflect.Value, error) {
	kind := fd.Kind()
	f, ok := filedParseFunc[kind]
	if !ok {
		panic(fmt.Sprintf("unknown field kind: %v", kind))
	}
	return f(value, fd)
}

var filedParseFunc = map[protoreflect.Kind]func(string, protoreflect.FieldDescriptor) (protoreflect.Value, error){
	protoreflect.BoolKind:     parseBoolKind,
	protoreflect.EnumKind:     parseEnumKind,
	protoreflect.Int32Kind:    parse32Kind,
	protoreflect.Sint32Kind:   parse32Kind,
	protoreflect.Sfixed32Kind: parse32Kind,
	protoreflect.Int64Kind:    parse64Kind,
	protoreflect.Sint64Kind:   parse64Kind,
	protoreflect.Sfixed64Kind: parse64Kind,
	protoreflect.Uint32Kind:   parseU32Kind,
	protoreflect.Fixed32Kind:  parseU32Kind,
	protoreflect.Uint64Kind:   parseU64Kind,
	protoreflect.Fixed64Kind:  parseU64Kind,
	protoreflect.FloatKind:    parseFloat32Kind,
	protoreflect.DoubleKind:   parseFloat64Kind,
	protoreflect.StringKind:   parseStringKind,
	protoreflect.BytesKind:    parseBytesKind,
	protoreflect.MessageKind:  parseMessage,
	protoreflect.GroupKind:    parseMessage,
}

func parseBoolKind(value string, _ protoreflect.FieldDescriptor) (protoreflect.Value, error) {
	v, err := strconv.ParseBool(value)
	if err != nil {
		return protoreflect.Value{}, err
	}
	return protoreflect.ValueOfBool(v), nil
}

func parseEnumKind(value string, fd protoreflect.FieldDescriptor) (protoreflect.Value, error) {
	name := fd.Enum().FullName()
	enum, err := protoregistry.GlobalTypes.FindEnumByName(name)
	if err != nil {
		if errors.Is(err, protoregistry.NotFound) {
			return protoreflect.Value{}, fmt.Errorf("enum %s not found", name)
		}
		return protoreflect.Value{}, fmt.Errorf("failed to find enum: %+v", err)
	}

	v := enum.Descriptor().Values().ByName(protoreflect.Name(value))
	if v == nil {
		i, err := strconv.ParseInt(value, 10, 32)
		if err != nil {
			return protoreflect.Value{}, fmt.Errorf("%s is invalid", value)
		}
		v = enum.Descriptor().Values().ByNumber(protoreflect.EnumNumber(i))
		if v == nil {
			return protoreflect.Value{}, fmt.Errorf("%s is invalid", value)
		}
	}
	return protoreflect.ValueOfEnum(v.Number()), nil
}

func parse32Kind(value string, _ protoreflect.FieldDescriptor) (protoreflect.Value, error) {
	v, err := strconv.ParseInt(value, 10, 32)
	if err != nil {
		return protoreflect.Value{}, err
	}
	return protoreflect.ValueOfInt32(int32(v)), nil
}

func parse64Kind(value string, _ protoreflect.FieldDescriptor) (protoreflect.Value, error) {
	v, err := strconv.ParseInt(value, 10, 64)
	if err != nil {
		return protoreflect.Value{}, err
	}
	return protoreflect.ValueOfInt64(v), nil
}

func parseU32Kind(value string, _ protoreflect.FieldDescriptor) (protoreflect.Value, error) {
	v, err := strconv.ParseUint(value, 10, 32)
	if err != nil {
		return protoreflect.Value{}, err
	}
	return protoreflect.ValueOfUint32(uint32(v)), nil
}

func parseU64Kind(value string, _ protoreflect.FieldDescriptor) (protoreflect.Value, error) {
	v, err := strconv.ParseUint(value, 10, 64)
	if err != nil {
		return protoreflect.Value{}, err
	}
	return protoreflect.ValueOfUint64(v), nil
}

func parseFloat32Kind(value string, _ protoreflect.FieldDescriptor) (protoreflect.Value, error) {
	v, err := strconv.ParseFloat(value, 32)
	if err != nil {
		return protoreflect.Value{}, err
	}
	return protoreflect.ValueOfFloat32(float32(v)), nil
}

func parseFloat64Kind(value string, _ protoreflect.FieldDescriptor) (protoreflect.Value, error) {
	v, err := strconv.ParseFloat(value, 64)
	if err != nil {
		return protoreflect.Value{}, err
	}
	return protoreflect.ValueOfFloat64(v), nil
}

func parseStringKind(value string, _ protoreflect.FieldDescriptor) (protoreflect.Value, error) {
	return protoreflect.ValueOfString(value), nil
}

func parseBytesKind(value string, _ protoreflect.FieldDescriptor) (protoreflect.Value, error) {
	v, err := base64.StdEncoding.DecodeString(value)
	if err != nil {
		return protoreflect.Value{}, err
	}
	return protoreflect.ValueOfBytes(v), nil
}

func parseMessage(value string, fd protoreflect.FieldDescriptor) (protoreflect.Value, error) {
	md := fd.Message()
	f, ok := msgParseFunc[md.FullName()]
	if !ok {
		return protoreflect.Value{}, fmt.Errorf("unsupported message type: %s", string(md.FullName()))
	}
	msg, err := f(value)
	if err != nil {
		return protoreflect.Value{}, err
	}
	return protoreflect.ValueOfMessage(msg.ProtoReflect()), nil
}

var msgParseFunc = map[protoreflect.FullName]func(string) (proto.Message, error){
	"google.protobuf.Timestamp":   parseTimestamp,
	"google.protobuf.Duration":    parseDuration,
	"google.protobuf.DoubleValue": parseFloat64,
	"google.protobuf.FloatValue":  parseFloat32,
	"google.protobuf.Int64Value":  parseInt64,
	"google.protobuf.Int32Value":  parseInt32,
	"google.protobuf.UInt64Value": parseUint64,
	"google.protobuf.UInt32Value": parseUint32,
	"google.protobuf.BoolValue":   parseBool,
	"google.protobuf.StringValue": parseString,
	"google.protobuf.BytesValue":  parseBytes,
	"google.protobuf.FieldMask":   parseFieldMask,
	"google.protobuf.Value":       parseValue,
	"google.protobuf.Struct":      parseStruct,
}

func parseTimestamp(v string) (proto.Message, error) {
	var msg proto.Message
	if v == "null" {
		return msg, nil
	}
	t, err := time.Parse(time.RFC3339Nano, v)
	if err != nil {
		return nil, err
	}
	return timestamppb.New(t), nil
}

func parseDuration(v string) (proto.Message, error) {
	var msg proto.Message
	if v == "null" {
		return msg, nil
	}
	d, err := time.ParseDuration(v)
	if err != nil {
		return nil, err
	}
	return durationpb.New(d), nil
}

func parseFloat64(v string) (proto.Message, error) {
	f, err := strconv.ParseFloat(v, 64)
	if err != nil {
		return nil, err
	}
	return wrapperspb.Double(f), nil
}

func parseFloat32(v string) (proto.Message, error) {
	f, err := strconv.ParseFloat(v, 32)
	if err != nil {
		return nil, err
	}
	return wrapperspb.Float(float32(f)), nil
}

func parseInt64(v string) (proto.Message, error) {
	i, err := strconv.ParseInt(v, 10, 64)
	if err != nil {
		return nil, err
	}
	return wrapperspb.Int64(i), nil
}

func parseInt32(v string) (proto.Message, error) {
	i, err := strconv.ParseInt(v, 10, 32)
	if err != nil {
		return nil, err
	}
	return wrapperspb.Int32(int32(i)), nil
}

func parseUint64(v string) (proto.Message, error) {
	u, err := strconv.ParseUint(v, 10, 64)
	if err != nil {
		return nil, err
	}
	return wrapperspb.UInt64(u), nil
}

func parseUint32(v string) (proto.Message, error) {
	u, err := strconv.ParseUint(v, 10, 32)
	if err != nil {
		return nil, err
	}
	return wrapperspb.UInt32(uint32(u)), nil
}

func parseBool(v string) (proto.Message, error) {
	b, err := strconv.ParseBool(v)
	if err != nil {
		return nil, err
	}
	return wrapperspb.Bool(b), nil
}

func parseString(v string) (proto.Message, error) {
	return wrapperspb.String(v), nil
}

func parseBytes(v string) (proto.Message, error) {
	s, err := base64.StdEncoding.DecodeString(v)
	if err != nil {
		s, err = base64.URLEncoding.DecodeString(v)
		if err != nil {
			return nil, err
		}
	}
	return wrapperspb.Bytes(s), nil
}

func parseFieldMask(v string) (proto.Message, error) {
	fm := &fieldmaskpb.FieldMask{}
	for _, fv := range strings.Split(v, ",") {
		fm.Paths = append(fm.Paths, utils.SnakeCase(fv))
	}
	return fm, nil
}

func parseValue(v string) (proto.Message, error) {
	value, err := structpb.NewValue(v)
	if err != nil {
		return nil, err
	}
	return value, nil
}

func parseStruct(v string) (proto.Message, error) {
	s := &structpb.Struct{}
	if err := protojson.Unmarshal([]byte(v), s); err != nil {
		return nil, err
	}
	return s, nil
}
