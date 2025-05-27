package comm

type phase uint8

const (
	init_phase phase = iota
	pairing_phase
	transfer_phase
	disconnect_phase
)

type expectedMessage uint8

const (
	// init connection phase
	conn_request expectedMessage = iota
	client_init
	client_finish
	conn_response
	// pairing phase
	paired_key_encryption
	paired_key_result
	// transfer phase
	introduction
	accept_reject
	transfer_complete
)

type transferProgress uint8

const (
	transfer_not_started transferProgress = iota
	transfer_in_progress
	transfer_finished
	transfer_error
)
