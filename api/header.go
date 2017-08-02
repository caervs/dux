package api

import "io"

type Header struct {
	Session int64
	Size    int
}

func ReadHeader(in io.Reader) (*Header, error) {
	buff := make([]byte, 10)
	_, err := io.ReadFull(in, buff)
	if err != nil {
		return nil, err
	}
	var session int64
	for i := 0; i < 8; i++ {
		session += int64(buff[i]) << uint((7-i)*8)
	}
	size := int(buff[9]) + (int(buff[8]) << 8)
	return &Header{
		Session: session,
		Size:    size,
	}, nil
}

func (h *Header) ToBytes() []byte {
	buff := make([]byte, 10)
	for i := 0; i < 8; i++ {
		buff[i] = byte((h.Session >> uint((7-i)*8)) % 256)
	}
	buff[8] = byte(h.Size >> 8)
	buff[9] = byte(h.Size % 256)
	return buff
}
