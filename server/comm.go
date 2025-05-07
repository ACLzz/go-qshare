package server

import (
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"math/big"
	"net"
	"sync"

	pbConnections "github.com/ACLzz/go-qshare/protobuf/gen/connections"
	pbSecuregcm "github.com/ACLzz/go-qshare/protobuf/gen/securegcm"
	pbSecureMessage "github.com/ACLzz/go-qshare/protobuf/gen/securemessage"
	"google.golang.org/protobuf/proto"
)

const max_message_length = 32 * 1024 // 32 Kb

type commServer struct {
	sock   net.Listener
	wg     *sync.WaitGroup
	stopCh chan struct{}
	// TODO: use kind of logger instead of prints
}

func newCommServer(port int) (commServer, error) {
	sock, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return commServer{}, fmt.Errorf("start socket: %w", err)
	}

	return commServer{
		sock:   sock,
		wg:     &sync.WaitGroup{},
		stopCh: make(chan struct{}, 1),
	}, nil
}

func (cs commServer) Listen() {
	var (
		conn net.Conn
		err  error
	)
	cs.wg.Add(1)
	defer cs.wg.Done()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	for {
		conn, err = cs.sock.Accept()
		if err != nil {
			select {
			case <-cs.stopCh:
				return
			default:
				fmt.Printf("ERR: accept conn: %s", err.Error())
			}
			continue
		}

		cs.wg.Add(1)
		c := newCommConn(conn)
		go c.read(ctx, cs.wg) // TODO: maybe limit amount of gorutines here
	}
}

type (
	commConn struct {
		conn  net.Conn
		phase phase

		// TODO: leave only necessary info from messages in conn, delete full messages
		connRequest        *pbConnections.ConnectionRequestFrame
		ukeyClientInit     *pbSecuregcm.Ukey2ClientInit
		ukeyClientFinished *pbSecuregcm.Ukey2ClientFinished
	}

	phase uint8
)

const (
	connPhase phase = iota
	pairPhase
	// ...
)

func newCommConn(conn net.Conn) commConn {
	cConn := commConn{
		conn:  conn,
		phase: connPhase,
	}

	return cConn
}

func (cc commConn) read(ctx context.Context, wg *sync.WaitGroup) {
	var (
		msgLen uint32
		err    error
		lenBuf = make([]byte, 4)
	)
	defer wg.Done()

	for {
		select {
		case <-ctx.Done():
			return
		default:
			// fetch length of the upcoming message
			if _, err = cc.conn.Read(lenBuf); err != nil {
				if errors.Is(err, io.EOF) {
					return
				}

				fmt.Printf("ERR: read msg length: %s", err.Error())
				continue
			}

			msgLen = binary.BigEndian.Uint32(lenBuf[:])
			if msgLen > max_message_length {
				fmt.Printf("ERR: message is too long (%d)\n", msgLen)
				return
			}

			fmt.Printf("message length: %d\n", msgLen)
			msgBuf := make([]byte, msgLen)

			// fetch message in bytes
			if _, err = cc.conn.Read(msgBuf); err != nil {
				fmt.Printf("ERR: fetch message: %s", err.Error())
				continue
			}

			if err = cc.processMessage(msgBuf); err != nil {
				fmt.Printf("ERR: process message: %s", err.Error())
				continue
			}
		}
	}
}

func (cc *commConn) processMessage(msg []byte) error {
	if cc.phase == connPhase {
		// first seek for the connection request
		if cc.connRequest == nil {
			var frame pbConnections.V1Frame
			if err := proto.Unmarshal(msg, &frame); err != nil {
				return fmt.Errorf("unmarshal frame: %w", err)
			}
			if frame.GetType() != pbConnections.V1Frame_CONNECTION_REQUEST {
				return ErrInvalidMessage
			}

			// for some reason it is marshaled second time in EndpointName :)
			var connRequest pbConnections.ConnectionRequestFrame
			if err := proto.Unmarshal([]byte(frame.ConnectionRequest.GetEndpointName()), &connRequest); err != nil {
				return fmt.Errorf("unmarshal Connection request: %w", err)
			}

			return cc.processConnRequest(&connRequest)

		}

		var ukeyMessage pbSecuregcm.Ukey2Message
		if err := proto.Unmarshal(msg, &ukeyMessage); err != nil {
			return fmt.Errorf("unmarshal UKEY message: %w", err)
		}

		if isAlertUKEY2MessageType(ukeyMessage.GetMessageType()) {
			return fmt.Errorf("got an error from client: '%s'", string(ukeyMessage.GetMessageData()))
		}

		// second seek for the ukey client init message
		if cc.ukeyClientInit == nil {
			if ukeyMessage.GetMessageType() != pbSecuregcm.Ukey2Message_CLIENT_INIT {
				return ErrInvalidMessage
			}

			var clientInitMsg pbSecuregcm.Ukey2ClientInit
			if err := proto.Unmarshal(ukeyMessage.GetMessageData(), &clientInitMsg); err != nil {
				return fmt.Errorf("unmarshal UKEY Client Init message: %w", err)
			}

			return cc.processUKEYInitMessage(&clientInitMsg)
		}

		// third seek for the ukey client finished message
		if cc.ukeyClientFinished == nil {
			if ukeyMessage.GetMessageType() != pbSecuregcm.Ukey2Message_CLIENT_FINISH {
				return ErrInvalidMessage
			}

			var clientFinishedMsg pbSecuregcm.Ukey2ClientFinished
			if err := proto.Unmarshal(ukeyMessage.GetMessageData(), &clientFinishedMsg); err != nil {
				return fmt.Errorf("unmarshal UKEY Client Finish message: %w", err)
			}

			return cc.processUKEYFinishedMessage(&clientFinishedMsg)
		}
	}

	return ErrInvalidMessage
}

func (cc *commConn) processConnRequest(req *pbConnections.ConnectionRequestFrame) error {
	if req == nil {
		return ErrInvalidMessage
	}

	if len(req.EndpointInfo) == 0 {
		return ErrInvalidMessage
	}

	cc.connRequest = req
	return nil
}

func (cc *commConn) processUKEYInitMessage(msg *pbSecuregcm.Ukey2ClientInit) error {
	cc.ukeyClientInit = msg

	// TODO: maybe add another type of cipher
	targetCipher := pbSecuregcm.Ukey2HandshakeCipher_P256_SHA512
	isSHA512Found := false
	for _, comm := range msg.CipherCommitments {
		if comm.GetHandshakeCipher() == targetCipher {
			isSHA512Found = true
			break
		}
	}
	if !isSHA512Found {
		return ErrInvalidHandshakeCipher
	}

	// TODO: do we need private key?
	privateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return fmt.Errorf("generate ecdsa private key: %w", err)
	}

	pbPublicKey, err := proto.Marshal(&pbSecureMessage.GenericPublicKey{
		Type: pbSecureMessage.PublicKeyType_EC_P256.Enum(),
		EcP256PublicKey: &pbSecureMessage.EcP256PublicKey{
			X: privateKey.PublicKey.X.FillBytes(make([]byte, 33)), // TODO: what does it do?
			Y: privateKey.PublicKey.Y.FillBytes(make([]byte, 33)),
		},
	})
	if err != nil {
		return fmt.Errorf("marshal Generic Public Key: %w", err)
	}

	random, err := randomBytes(32)
	if err != nil {
		return fmt.Errorf("generate random field for UKEY2 Server Init message: %w", err)
	}

	serverUKEYInit, err := proto.Marshal(&pbSecuregcm.Ukey2ServerInit{
		Version:         msg.Version,
		Random:          random,
		HandshakeCipher: targetCipher.Enum(),
		PublicKey:       pbPublicKey,
	})
	if err != nil {
		return fmt.Errorf("marshal UKEY2 Server Init response: %w", err)
	}

	if err = cc.writeUKEY2Message(pbSecuregcm.Ukey2Message_SERVER_INIT, serverUKEYInit); err != nil {
		return fmt.Errorf("response to client with UKEY2 Server Init: %w", err)
	}

	return nil
}

func (cc *commConn) processUKEYFinishedMessage(req *pbSecuregcm.Ukey2ClientFinished) error {
	if req == nil {
		return ErrInvalidMessage
	}

	fmt.Println("got finished message")
	return nil
}

func (cc *commConn) writeUKEY2Message(t pbSecuregcm.Ukey2Message_Type, msgData []byte) error {
	msg, err := proto.Marshal(&pbSecuregcm.Ukey2Message{
		MessageType: t.Enum(),
		MessageData: msgData,
	})
	if err != nil {
		return fmt.Errorf("marshal UKEY2 Message: %w", err)
	}

	// TODO: is there more efficient way to do this?
	msgLength := len(msg)
	msgWithPrefix := make([]byte, 4, msgLength+4)
	msgWithPrefix[0] = byte(uint8(msgLength >> 24))
	msgWithPrefix[1] = byte(uint8(msgLength >> 16))
	msgWithPrefix[2] = byte(uint8(msgLength >> 8))
	msgWithPrefix[3] = byte(uint8(msgLength))
	msgWithPrefix = append(msgWithPrefix, msg...)

	if _, err = cc.conn.Write(msgWithPrefix); err != nil {
		return fmt.Errorf("response to client with UKEY2 Message: %w", err)
	}

	return nil
}

func isAlertUKEY2MessageType(t pbSecuregcm.Ukey2Message_Type) bool {
	// TODO: fixme
	// for k := range pbSecuregcm.Ukey2Alert_AlertType_name {
	// 	if int32(t.Number()) == k {
	// 		return true
	// 	}
	// }

	return t >= 100
}

func (cs commServer) Stop() error {
	if cs.sock == nil {
		return nil
	}

	cs.stopCh <- struct{}{}
	if err := cs.sock.Close(); err != nil {
		return fmt.Errorf("close socket: %w", err)
	}
	cs.wg.Wait()

	return nil
}

func randomBytes(size uint8) ([]byte, error) {
	buf := make([]byte, size)

	for i := range size {
		n, err := rand.Int(rand.Reader, big.NewInt(256))
		if err != nil {
			return nil, fmt.Errorf("generate random int: %w", err)
		}

		buf[i] = byte(n.Int64())
	}

	return buf, nil
}
