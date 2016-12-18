README

Dependencies:
The program requires the following libraries in the GOPATH
Termbox: https://github.com/nsf/termbox-go, use 'go get -u github.com/nsf/termbox-go'
UUID package for Go: https://github.com/satori/go.uuid, use 'go get github.com/satori/go.uuid'
GoVector: https://github.com/arcaneiceman/GoVector, use 'go get github.com/arcaneiceman/GoVector'
Codec: https://github.com/hashicorp/go-msgpack, use 'go get github.com/hashicorp/go-msgpack/codec'

Usage:
go run editor.go [listening port] [public listening port]

Starts a colloborative editing peer that listens at listening port.

The public listening port is optional. It specifies the address through which other nodes can connect to the current node.
This is helpful when the program is being run behind a NAT. If it's not provided, we assume other nodes can connect
via the listening port. To test the program behind a NAT, ngrok can help to expose the local port. (See https://ngrok.com/)

Currently in master, using the public address provided by ngrok and connect to nodes on remote network is not working.
It was tested to work in commit 66f117c8183 where the main program is called guit.go, when the GoVector wasn't integrated.
In master, the GoVector library panics with the ngrok tunnel.

The program starts in the menu screen. Type 1 to start a new document. Press Esc to switch between the menu and editing the document.
Esc and now there will be options to connect to other peers and receive their document and collaboratively edit with said peers.
There's an options to disconnect from the network and prevent any peers to connect to you. You can reconnect using the above options.
Closing the document will disconnect you from the network completely and close the document. You will have to restart a new document
and connect to the original network again.

Example (on localhost)
1. Run 'go run editor.go localhost:1000' in a command line prompt
2. Enter 1 for New Document
3. Run 'go run editor.go localhost:2000' in another command line prompt
4. Enter 1 for New Document
5. Press Esc
6. Enter 1 for Connect
7. Enter localhost:1000 as the ip to connect to
8. Press Esc to edit document, now the two peers can edit collaboratively