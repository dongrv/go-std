package toolkit

import (
	"encoding/binary"
	"errors"
)

var ErrWrongEndianLength = errors.New("wrong target endian length")

// PutBigEndian 大端模式写入数据
func PutBigEndian(buf []byte, len int, n int) error {
	switch len {
	case 1:
		buf[0] = byte(n)
	case 2:
		binary.BigEndian.PutUint16(buf[:2], uint16(n))
	case 4:
		binary.BigEndian.PutUint32(buf[:4], uint32(n))
	case 8:
		binary.BigEndian.PutUint64(buf[:8], uint64(n))
	default:
		return ErrWrongEndianLength
	}
	return nil
}

// GetBigEndian 大端模式获取数据值
func GetBigEndian(buf []byte, len int) (int, error) {
	switch len {
	case 1:
		return int(buf[0]), nil
	case 2:
		return int(binary.BigEndian.Uint16(buf[:2])), nil
	case 4:
		return int(binary.BigEndian.Uint32(buf[:4])), nil
	case 8:
		return int(binary.BigEndian.Uint64(buf[:8])), nil
	default:
		return 0, ErrWrongEndianLength
	}
}

// PutLittleEndian 大端模式写入数据
func PutLittleEndian(buf []byte, len int, n int) error {
	switch len {
	case 1:
		buf[0] = byte(n)
	case 2:
		binary.LittleEndian.PutUint16(buf[:2], uint16(n))
	case 4:
		binary.LittleEndian.PutUint32(buf[:4], uint32(n))
	case 8:
		binary.LittleEndian.PutUint64(buf[:8], uint64(n))
	default:
		return ErrWrongEndianLength
	}
	return nil
}

// GetLittleEndian 大端模式获取数据值
func GetLittleEndian(buf []byte, len int) (int, error) {
	switch len {
	case 1:
		return int(buf[0]), nil
	case 2:
		return int(binary.LittleEndian.Uint16(buf[:2])), nil
	case 4:
		return int(binary.LittleEndian.Uint32(buf[:4])), nil
	case 8:
		return int(binary.LittleEndian.Uint64(buf[:8])), nil
	default:
		return 0, ErrWrongEndianLength
	}
}
