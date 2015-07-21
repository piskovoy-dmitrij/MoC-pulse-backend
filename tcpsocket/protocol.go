package tcpsocket

/*
	Tcp Server uses next protocol.
	Every packet should have header.
	Also packet can have content data.

	Header structure:
		2 bytes - opcode number. Opcodes are described below.
		4 bytes - content size.
	Content structure:
		It has json structure, every opcode has its own keys.
*/

// Opcodes
// CS_* - Client to Server
// SC_* - Server to Client
const (
	CS_AUTH = iota
	SC_AUTH
	CS_GET_VOTES
	SC_GET_VOTES_RESULT
	CS_GET_VOTE
	SC_GET_VOTE_RESULT
	CS_VOTE_FOR
	SC_UPDATE_VOTE
	CS_CREATE_VOTE
	SC_NEW_VOTE
)

// CS_AUTH
type CSAuthRequest struct {
	Token string `json:"token"`
}

// CS_GET_VOTES
type CSGetVotesRequest struct {
	// empty
}

// CS_GET_VOTE
type CSGetVoteRequest struct {
	Id string `json:"id"`
}

// CS_VOTE_FOR
type CSVoteForRequest struct {
	Id      string `json:"id"`
	ColorId int    `json:"color"`
}

// CS_CREATE_VOTE
type CSCreateVoteRequest struct {
	Name string `json:"name"`
}
