package comm

type phase uint8

const (
	initConnPhase phase = iota
	pairingPhase
	transferPhase
	// ...
)

type messageType uint8

const (
	unknown_message_type messageType = iota
	// init connection phase
	conn_request
	client_init
	client_finish
	conn_response
	// pairing phase
	paired_key_encryption
	paired_key_result
	// transfer phase
	introduction
)
