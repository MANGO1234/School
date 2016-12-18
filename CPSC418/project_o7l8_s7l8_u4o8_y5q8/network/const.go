package network

// TODO: this is used for the connection logic which is spread in network.go and nodethread.go
// Will move the stuff in this file in somewhere that makes more sense
// SORRY IN ADVANCE

// Note on the design of how connection between nodes are established:
//
// The node with smaller listening addr pokes    the node with a larger  one.
// The node with larger  listening addr connects the node with a smaller one.
// We distinguish between client-initiated vs. system-initiated dials because
// the client-initiated ones doesn't know the uuid of the other node. For the
// system-initiated dials, if the uuid of the node with the listening address
// changes, it means that the previous session ended either caused by a user
// initiated disconnect command or by stopping the node and restarting with
// same listening address. In this case, we need to consider the old session
// associated with the the uuid we have as left.
// this scheme avoids deadlock and infinite connection replacement where we
// maintain only one connection between any two nodes
//
// For example, when A and B connect to each other simultaneously,
// with a deterministic algorithm, it's possible that the two nodes
// performs symmetric actions, causing connections to be constantly
// replaced and failure to agree on one single connection between
// the nodes. If we try to use locks or channels to forbid this
// unnecessary reconnection loop, it's easy to get deadlock. Having
// two connections between the nodes solves this problem in a way
// but handling two connections when we only need to handle one is
// costly and leads to other problems associated with handling more
// connections

// the purpose of dialing to a node
const (
	// client-initiated poke to a remote node
	dialingTypeRegister = "register"
	// poke a known node so it has information to connect
	dialingTypePoke = "poke"
	// establish persistent connection between the nodes
	dialingTypeConnect = "connect"
)
