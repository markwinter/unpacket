package unpacket

import (
	"encoding/binary"
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"time"
)

const (
	structTag = "unpack"
)

var (
	buf2 = make([]byte, 2)
	buf4 = make([]byte, 4)
	buf8 = make([]byte, 8)
)

func parseUnpackTag(tag string) (int, int, error) {
	offset := -1
	length := -1

	for _, part := range strings.Split(tag, ",") {
		if strings.HasPrefix(part, "offset=") {
			o, err := strconv.Atoi(strings.TrimPrefix(part, "offset="))
			if err != nil {
				return 0, 0, fmt.Errorf("invalid offset")
			}
			offset = o
		} else if strings.HasPrefix(part, "length=") {
			l, err := strconv.Atoi(strings.TrimPrefix(part, "length="))
			if err != nil {
				return 0, 0, fmt.Errorf("invalid length")
			}
			length = l
		}
	}

	if offset == -1 || length == -1 {
		return 0, 0, fmt.Errorf("missing offset or length")
	}

	return offset, length, nil
}

func setFieldFromBytes(fieldValue reflect.Value, data []byte, byteOrder binary.ByteOrder) error {
	if fieldValue.Type() == reflect.TypeOf(time.Duration(0)) {
		// Handle time.Duration with fewer than 8 bytes (e.g., 6 bytes)
		if len(data) < 6 {
			return fmt.Errorf("invalid length for time.Duration")
		}
		// Pad with 0s to make it an 8-byte slice
		paddedData := append(make([]byte, 8-len(data)), data...)
		val := byteOrder.Uint64(paddedData)
		fieldValue.Set(reflect.ValueOf(time.Duration(val)))
		return nil
	}

	switch fieldValue.Kind() {
	case reflect.Uint8:
		if len(data) < 1 {
			return fmt.Errorf("invalid length for uint8")
		}
		val := data[0]
		fieldValue.SetUint(uint64(val))
	case reflect.Uint16:
		if len(data) < 2 {
			return fmt.Errorf("invalid length for uint16")
		}
		val := byteOrder.Uint16(data[:2])
		fieldValue.SetUint(uint64(val))
	case reflect.Uint32:
		if len(data) < 4 {
			return fmt.Errorf("invalid length for uint32")
		}
		val := byteOrder.Uint32(data[:4])
		fieldValue.SetUint(uint64(val))
	case reflect.Uint64:
		if len(data) < 8 {
			return fmt.Errorf("invalid length for uint64")
		}
		val := byteOrder.Uint64(data[:8])
		fieldValue.SetUint(val)
	case reflect.Int8:
		if len(data) < 1 {
			return fmt.Errorf("invalid length for int8")
		}
		val := int8(data[0])
		fieldValue.SetInt(int64(val))
	case reflect.Int16:
		if len(data) < 2 {
			return fmt.Errorf("invalid length for int16")
		}
		val := int16(byteOrder.Uint16(data[:2]))
		fieldValue.SetInt(int64(val))
	case reflect.Int32:
		if len(data) < 4 {
			return fmt.Errorf("invalid length for int32")
		}
		val := int32(byteOrder.Uint32(data[:4]))
		fieldValue.SetInt(int64(val))
	case reflect.Int64:
		fmt.Printf("HERE")
		if len(data) < 8 {
			return fmt.Errorf("invalid length for int64")
		}
		val := int64(byteOrder.Uint64(data[:8]))
		fieldValue.SetInt(val)
	case reflect.Bool:
		if len(data) < 1 {
			return fmt.Errorf("invalid length for bool")
		}
		val := data[0] != 0
		fieldValue.SetBool(val)
	case reflect.String:
		str := strings.TrimSpace(string(data))
		fieldValue.SetString(str)
	default:
		return fmt.Errorf("unsupported field type: %s", fieldValue.Type())
	}
	return nil
}

func fieldToBytes(fieldValue reflect.Value, length int, order binary.ByteOrder) ([]byte, error) {
	if fieldValue.Type() == reflect.TypeOf(time.Duration(0)) {
		buf := make([]byte, 8)
		duration := uint64(fieldValue.Int())
		order.PutUint64(buf, duration)
		return buf[8-length:], nil
	}

	switch fieldValue.Kind() {
	case reflect.Uint8:
		return []byte{byte(fieldValue.Uint())}, nil
	case reflect.Uint16:
		order.PutUint16(buf2, uint16(fieldValue.Uint()))
		return buf2, nil
	case reflect.Uint32:
		order.PutUint32(buf4, uint32(fieldValue.Uint()))
		return buf4, nil
	case reflect.Uint64:
		order.PutUint64(buf8, fieldValue.Uint())
		return buf8, nil
	case reflect.Int8:
		return []byte{byte(fieldValue.Int())}, nil
	case reflect.Int16:
		order.PutUint16(buf2, uint16(fieldValue.Int()))
		return buf2, nil
	case reflect.Int32:
		order.PutUint32(buf4, uint32(fieldValue.Int()))
		return buf4, nil
	case reflect.Int64:
		order.PutUint64(buf8, uint64(fieldValue.Int()))
		return buf8, nil
	case reflect.String:
		str := fieldValue.String()
		if len(str) > length {
			str = str[:length] // Trim if the string is too long
		}
		buf := make([]byte, length)
		copy(buf, []byte(str)) // Pad with zeroes if too short
		return buf, nil
	}

	return nil, fmt.Errorf("unsupported field type: %s", fieldValue.Type())
}

// Unpack reads the data bytes and unpacks them into the given struct
func Unpack(data []byte, order binary.ByteOrder, datastruct any) error {
	val := reflect.ValueOf(datastruct).Elem()

	if val.Kind() != reflect.Struct {
		return errors.New("not supplied a struct")
	}

	for i := 0; i < val.NumField(); i++ {
		field := val.Type().Field(i)
		fieldValue := val.Field(i)

		tag, exists := field.Tag.Lookup(structTag)
		if !exists {
			continue
		}

		offset, length, err := parseUnpackTag(tag)
		if err != nil {
			return fmt.Errorf("failed parsing unpack tag for field [%s]: %w", field.Name, err)
		}

		if offset+length > len(data) {
			return fmt.Errorf("data out of bounds")
		}

		fieldBytes := data[offset : offset+length]

		if err := setFieldFromBytes(fieldValue, fieldBytes, order); err != nil {
			return fmt.Errorf("failed to set field %s: %w", field.Name, err)
		}
	}

	return nil
}

func Pack(order binary.ByteOrder, dataStruct any) ([]byte, error) {
	v := reflect.ValueOf(dataStruct).Elem()
	if v.Kind() != reflect.Struct {
		return nil, fmt.Errorf("expected a struct, but got %s", v.Kind())
	}

	t := v.Type()

	// Determine the size of the byte array by finding the largest offset + length
	maxOffset := 0
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)

		tag := field.Tag.Get(structTag)
		if tag == "" {
			continue
		}

		offset, length, err := parseUnpackTag(tag)
		if err != nil {
			return nil, err
		}

		if offset+length > maxOffset {
			maxOffset = offset + length
		}
	}

	result := make([]byte, maxOffset)

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		fieldValue := v.Field(i)

		tag := field.Tag.Get(structTag)
		if tag == "" {
			continue
		}

		offset, length, err := parseUnpackTag(tag)
		if err != nil {
			return nil, err
		}

		fieldBytes, err := fieldToBytes(fieldValue, length, order)
		if err != nil {
			return nil, fmt.Errorf("failed to convert field [%s] to bytes: %v", field.Name, err)
		}

		copy(result[offset:offset+length], fieldBytes)
	}

	return result, nil
}
