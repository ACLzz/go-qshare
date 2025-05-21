package comm

type phase uint8

const (
	initConnPhase phase = iota
	pairingPhase
	transferPhase
	// ...
)

type expectedMessage uint8

const (
	unknown_message_type expectedMessage = iota
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
	accept_reject
)

type transferProgress uint8

const (
	transfer_not_started transferProgress = iota
	transfer_in_progress
	transfer_finished
	transfer_error
)
