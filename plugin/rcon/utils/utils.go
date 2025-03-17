// utils 来自 github.com/bearbin/mcgorcon
package utils

import (
	"bytes"
	"encoding/binary"
	"errors"
	"io"
	"net"
	"time"

	"github.com/Kittengarten/KittenCore/kitten/core"
)

const (
	BadAuth        = -1
	PayloadMaxSize = 1460
)

const (
	PacketResponse = iota
	_
	PacketCommand
	PacketLogin
)

type MCConn struct {
	conn     net.Conn
	password string
}

type packetType int32

type Header struct {
	Size      int32
	RequestID int32
	Type      packetType
}

func (c *MCConn) Open(addr, password string) error {
	conn, err := net.DialTimeout(`tcp`, addr, core.TimeOutSeconds*time.Second)
	if err != nil {
		return err
	}
	*c = MCConn{
		conn:     conn,
		password: password,
	}
	return nil
}

func (c *MCConn) Close() error {
	return c.conn.Close()
}

// SendCommand 向服务器发送命令并返回结果
func (c *MCConn) SendCommand(command string) (string, error) {
	// 发送包
	if len(command) > PayloadMaxSize {
		return ``, errors.New(`命令过长喵！`)
	}
	head, payload, err := c.sendPacket(PacketCommand, []byte(command))
	if err != nil {
		return ``, err
	}
	// 验证失败，返回错误
	if head.RequestID == BadAuth {
		return ``, errors.New(`验证失败，不能发送命令喵！`)
	}
	return string(payload), nil
}

// Authenticate 验证用户身份
func (c *MCConn) Authenticate() error {
	// 发送包
	head, _, err := c.sendPacket(PacketLogin, []byte(c.password))
	if err != nil {
		return err
	}
	// 验证失败，返回错误
	if head.RequestID == BadAuth {
		return errors.New(`验证失败喵！`)
	}
	return nil
}

// sendPacket 发送二进制包并返回响应
func (c *MCConn) sendPacket(t packetType, p []byte) (Header, []byte, error) {
	// 生成二进制包
	packet, err := packetise(t, p)
	if err != nil {
		return Header{}, nil, err
	}
	// 发送二进制包
	_, err = c.conn.Write(packet)
	if err != nil {
		return Header{}, nil, err
	}
	// 接收并解码响应
	return depacketise(c.conn)
}

// packetise 编码数据包并转换为二进制表达
func packetise(t packetType, p []byte) ([]byte, error) {
	if len(p) > PayloadMaxSize {
		return nil, errors.New(`数据包太大了喵！`)
	}
	var buf bytes.Buffer
	err := errors.Join(
		binary.Write(&buf, binary.LittleEndian, int32(len(p)+10)),
		binary.Write(&buf, binary.LittleEndian, int32(0)),
		binary.Write(&buf, binary.LittleEndian, t),
		binary.Write(&buf, binary.LittleEndian, p),
		binary.Write(&buf, binary.LittleEndian, [2]byte{}),
	)
	if err != nil {
		return nil, err
	}
	// 数据包太大，无法处理
	if buf.Len() >= PayloadMaxSize {
		return nil, errors.New(`数据包太大了喵！`)
	}
	// 返回数据包的字节切片
	return buf.Bytes(), nil
}

// depacketise 解码数据包
func depacketise(r io.Reader) (Header, []byte, error) {
	head := Header{}
	if err := binary.Read(r, binary.LittleEndian, &head); err != nil {
		return Header{}, nil, err
	}
	payload := make([]byte, head.Size-8)
	if _, err := io.ReadFull(r, payload); err != nil {
		return Header{}, nil, err
	}
	// 检查
	switch head.Type {
	case PacketResponse, PacketCommand:
		return head, payload[:len(payload)-2], nil
	default:
		return Header{}, nil, errors.New(`数据包类型错误喵！`)
	}
}
