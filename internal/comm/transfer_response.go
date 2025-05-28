package comm

import pbSharing "github.com/ACLzz/go-qshare/internal/protobuf/gen/sharing"

func (a *Adapter) SendTransferResponse(isAccepted bool) error {
	status := pbSharing.ConnectionResponseFrame_REJECT
	if isAccepted {
		status = pbSharing.ConnectionResponseFrame_ACCEPT
	}

	return a.WriteSecureFrame(&pbSharing.V1Frame{
		Type: pbSharing.V1Frame_RESPONSE.Enum(),
		ConnectionResponse: &pbSharing.ConnectionResponseFrame{
			Status: status.Enum(),
		},
	})
}
