package adapter

import (
	"fmt"

	qshare "github.com/ACLzz/go-qshare"
	"github.com/ACLzz/go-qshare/internal/crypt"
	pbConnections "github.com/ACLzz/go-qshare/internal/protobuf/gen/connections"
	"google.golang.org/protobuf/proto"
)

type ConnRequest struct {
	EndpointInfo []byte
}

func (a *Adapter) UnmarshalConnRequest(msg []byte) (ConnRequest, error) {
	frame, err := unmarshalOfflineFrame(msg)
	if err != nil {
		return ConnRequest{}, err
	}

	if frame.GetType() != pbConnections.V1Frame_CONNECTION_REQUEST || frame.GetConnectionRequest() == nil {
		return ConnRequest{}, ErrInvalidMessage
	}

	return ConnRequest{
		EndpointInfo: frame.GetConnectionRequest().GetEndpointInfo(),
	}, nil
}

func (a *Adapter) SendConnRequest(endpointID, hostname string, device qshare.DeviceType) error {
	endpointInfo, err := CraftEndpointInfo(hostname, device)
	if err != nil {
		return fmt.Errorf("craft endpoint info: %w", err)
	}

	return a.writeOfflineFrame(&pbConnections.V1Frame{
		Type: pbConnections.V1Frame_CONNECTION_REQUEST.Enum(),
		ConnectionRequest: &pbConnections.ConnectionRequestFrame{
			EndpointId:   proto.String(endpointID),
			EndpointName: []byte(hostname + ".local"),
			EndpointInfo: endpointInfo,
			Mediums:      []pbConnections.ConnectionRequestFrame_Medium{pbConnections.ConnectionRequestFrame_WIFI_LAN},
		},
	})
}

func CraftEndpointInfo(hostname string, device qshare.DeviceType) ([]byte, error) {
	hostnameBytes := []byte(hostname)
	buf := make([]byte, 1, len(hostnameBytes)+17)

	buf[0] = byte(device << 1)
	randomData, err := crypt.RandomBytes(16)
	if err != nil {
		return nil, fmt.Errorf("generate random data: %w", err)
	}

	buf = append(buf, randomData...)
	buf = append(buf, byte(len(hostnameBytes)))
	buf = append(buf, hostnameBytes...)

	return buf, nil
}
