package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	// gx managed packages
	"gx/ipfs/QmR8BauakNcBa3RbE4nbQu76PDiJgoQgz8AJdhJuiU4TAw/go-cid"
	"gx/ipfs/QmRKbEchaYADxSCyyjhDh4cTrUby8ftXUb8MRLBTHQYupw/go-libp2p-net"
	"gx/ipfs/QmUSE3APe1pMFVsUBZUZaKQKERiPteCWvTAERtVQmtXzgE/go-ipfs-addr"
	"gx/ipfs/QmUymf8fJtideyv3z727BcZUifGBjMZMpCJqu3Gxk5aRUk/go-libp2p-peerstore"
	"gx/ipfs/QmVrjR2KMe57y4YyfHdYa3yKD278gN8W7CTiqSuYmxjA7F/go-libp2p-host"
	"gx/ipfs/QmXnpYYg2onGLXVxM4Q5PEFcx29k8zeJQkPeLAk9h9naxg/go-libp2p"
	"gx/ipfs/QmadRyQYRn64xHb5HKy2jRFp2Der643Cgo7NEjFgs4MX2k/go-libp2p-kad-dht"
	"gx/ipfs/QmcqU6QUDSXprb1518vYDGczrTJTyGwLG9eUa5iNX4xUtS/go-libp2p-peer"
	"gx/ipfs/QmerPMzPk1mJVowm8KgmoknWa4yCYvvugMPsgWmDNUvDLW/go-multihash"
)

// IPFS bootstrap nodes. Used to find other peers in the network.
var bootstrapPeers = []string{
	"/ip4/104.131.131.82/tcp/4001/ipfs/QmaCpDMGvV2BGHeYERUEnRQAwe3N8SzbUtfsmvsqQLuvuJ",
	"/ip4/104.236.179.241/tcp/4001/ipfs/QmSoLPppuBtQSGwKDZT2M73ULpjvfd3aZ6ha4oFGL1KrGM",
	"/ip4/104.236.76.40/tcp/4001/ipfs/QmSoLV4Bbm51jM9C4gDYZQ9Cy3U6aXMJDAbzgu2fzaDs64",
	"/ip4/128.199.219.111/tcp/4001/ipfs/QmSoLSafTMBsPKadTEgaXctDQVcqN88CNLHXMkTNwMKPnu",
	"/ip4/178.62.158.247/tcp/4001/ipfs/QmSoLer265NRgSp2LA3dPaeykiS1J6DifTC88f5uVQKNAd",
}

var (
	// Should be set most uniquely, i.e. ./chat -r $(cat uuid.txt)
	rendezvous string

	// Reader and writer regarding the streams and the chat, respectively
	readWriters []*bufio.ReadWriter

	// Slices to store the peers the chat is or were connected to
	writeToPeers  []peer.ID
	readFromPeers []peer.ID

	// The host of the chat
	chat host.Host

	// Some internal minor vars
	err       error
	lastError string
	cmdUsage  map[string]string
)

// Initialise the chat commands during boot
func commandUsageInit() {
	cmdUsage = make(map[string]string)

	cmdUsage["chat"] = "\\chat"
	cmdUsage["connections"] = "\\connections"
	cmdUsage["peer"] = "\\peer <peer.ID Qm*...>"
	cmdUsage["quit"] = "\\quit"
}

// executeCommand takes the first argument and executes the related function accordingly
func executeCommand(commandline string) {

	// Trim string and split result by white spaces
	commandFields := strings.Fields(strings.Trim(strings.Trim(commandline, " "), "\\"))

	// Check for empty string without prefix
	if len(commandFields) > 0 {

		// Switch according to the first word and call appropriate function with the rest as arguments
		switch commandFields[0] {

		case "chat":
			chatHost(commandFields[1:])

		case "connections":
			chatConnections(commandFields[1:])

		case "peer":
			chatPeer(commandFields[1:])

		case "quit":
			quitChat(commandFields[1:])

		default:
			usage()
		}
	} else {
		usage()
	}
}

// chatHost shows some data of this host
func chatHost(arguments []string) {

	// Get rid of warning
	_ = arguments

	fmt.Printf("<ID>: %s\n", chat.ID())
	fmt.Printf("<ID>: %s\n", chat.ID().Pretty())
	for i, cAddr := range chat.Addrs() {
		fmt.Printf("<ADDR %d>: %v\n", i, cAddr)
	}
	fmt.Print(chat.ID(), " ")
}

// chatConnections shows all connected peers of both directions
func chatConnections(arguments []string) {

	// Get rid of warning
	_ = arguments

	for i, wPeer := range writeToPeers {
		fmt.Printf("<WRITE_CONNECTIONS>: %d: %s\n", i, wPeer)
	}
	for i, rPeer := range readFromPeers {
		fmt.Printf("<READ_CONNECTIONS>: %d: %s\n", i, rPeer)
	}
	fmt.Print(chat.ID(), " ")
}

// chatPeer shows data of a specified peer
func chatPeer(arguments []string) {

	// Check at least two words exists
	if len(arguments) < 2 {
		fmt.Printf("ERROR: wrong format: e.g. %q\n", "<peer.ID Qm*YDJjDm>")
		return
	}

	// Join the two words of peer ID
	pIn := strings.Join(arguments[:2], " ")

	// Loop over all peers from the store of the chat
	for _, p := range chat.Peerstore().Peers() {

		// Search the given ID and print accordingly
		if p.String() == pIn {
			fmt.Printf("<ID>: %s\n", p)
			fmt.Printf("<ID>: %s\n", p.Pretty())
			for i, pAddr := range chat.Peerstore().Addrs(p) {
				fmt.Printf("<ADDR %d>: %v\n", i, pAddr)
			}
			isWriteConnected, isReadConnected := false, false
			for _, wc := range writeToPeers {
				if wc.String() == pIn {
					isWriteConnected = true
				}
			}
			for _, rc := range readFromPeers {
				if rc.String() == pIn {
					isReadConnected = true
				}
			}
			fmt.Printf("<WRITE CONNECTED>: %v\n", isWriteConnected)
			fmt.Printf("<READ CONNECTED>: %v\n", isReadConnected)
			fmt.Print(chat.ID(), " ")
		}
	}
}

// quitChat does the expected
func quitChat(arguments []string) {

	// Get rid of warning
	_ = arguments

	os.Exit(0)
}

// usage displays all available chat commands
func usage() {
	for _, cmd := range cmdUsage {
		fmt.Printf("<CMD USAGE>: %s\n", cmd)
	}
	fmt.Print(chat.ID(), " ")
}

// handleStream manages new incoming streams
func handleStream(stream net.Stream) {

	// Add new remote peer as peer to read from
	readFromPeers = append(readFromPeers, stream.Conn().RemotePeer())

	// Create a buffer stream for non blocking read and write
	rw := bufio.NewReadWriter(bufio.NewReader(stream), bufio.NewWriter(stream))

	// Go routine to process stream lines
	go readData(rw)

	// TODO: Unclear, if we need this Go routine and how many
	go writeData()

	// Shows the number of saved peers and connections, respectively
	fmt.Printf("\rInbound Connections: %d Outbound Connections: %d ",
		len(readFromPeers), len(writeToPeers))

	// 'stream' will stay open until you close it (or the other side closes it).
}

// TODO: Still we miss read connections in tests!

// readData reads the message from other peers and prints it currently with a green prompt
func readData(rw *bufio.ReadWriter) {

	// Continuously waiting for incoming lines
	for {
		str, err := rw.ReadString('\n')
		if err != nil {
			if err.Error() != lastError {
				log.Println("ERROR: ", err)
				lastError = err.Error()
			}
			continue
		}

		if str == "" {
			continue
		}

		if str != "\n" {
			// Green console colour: 	\x1b[32m
			// Reset console colour: 	\x1b[0m

			// By default the sender's peer id was sent at the beginning of every line
			fmt.Printf("\n\x1b[32m%s\x1b[0m", str)
			fmt.Printf("%s ", chat.ID())
		}
	}
}

// writeData takes lines from standard input and process it as message to be send or command to be executed
func writeData() {

	// Buffer reading from chat
	stdReader := bufio.NewReader(os.Stdin)

	// Keep reading
	for {

		// Wait and read last line
		line, err := stdReader.ReadString('\n')
		if err != nil {
			panic(err)
		}

		// Set default prompt
		fmt.Print(chat.ID(), " ")

		// An empty line writes a prompt locally but does not send anything
		if strings.Trim(line, "\n") == "" {
			continue
		}

		// Remove linebreak and check leading backslash for a chat command
		if strings.HasPrefix(strings.Trim(line, "\n"), "\\") {
			executeCommand(line)
			continue
		}

		// Loop over all connected writers
		for _, rx := range readWriters {

			// Write sender's ID and the last line written
			rx.WriteString(fmt.Sprintf("%v %s", chat.ID(), line))
			rx.Flush()
		}
	}
}

func main() {
	help := flag.Bool("h", false, "Display Help")
	rendezvousString := flag.String("r", rendezvous, "Unique string to identify group of nodes. Share this with your friends to let them connect with you")
	flag.Parse()

	if *help {
		fmt.Printf("This program demonstrates a simple p2p chat application using libp2p\n\n")
		fmt.Printf("Usage: Run './chat in two different terminals. Let them connect to the bootstrap nodes, announce themselves and connect to the peers\n")

		os.Exit(0)
	}

	// Set flags for logging
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)

	// Initialise the chat commands
	commandUsageInit()

	// libp2p.New constructs a new libp2p Host.
	// Other options can be added here.
	ctx := context.Background()
	chat, err = libp2p.New(ctx, libp2p.DisableRelay())
	if err != nil {
		panic(err)
	}
	fmt.Printf("Starting chat for peer %q ...\n", chat.ID())

	// Set a function as stream handler.
	// This function is called when a peer initiate a connection and starts a stream with this peer.
	chat.SetStreamHandler("/chat/1.1.0", handleStream)

	// Create new distributed hash table
	kadDht, err := dht.New(ctx, chat)
	if err != nil {
		panic(err)
	}

	// Let's connect to the bootstrap nodes first. They will tell us about the other nodes in the network.
	for _, peerAddr := range bootstrapPeers {
		pAddr, _ := ipfsaddr.ParseString(peerAddr)
		peerinfo, _ := peerstore.InfoFromP2pAddr(pAddr.Multiaddr())

		if err := chat.Connect(ctx, *peerinfo); err != nil {
			log.Println("ERROR: ", err)
		}
	}

	// We use a rendezvous point to announce our location.
	v1b := cid.V1Builder{Codec: cid.Raw, MhType: multihash.SHA2_256}
	rendezvousPoint, _ := v1b.Sum([]byte(*rendezvousString))

	// We provide the rendezvous point to the distributed hash table
	tctx, cancel := context.WithTimeout(ctx, time.Second*10)
	defer cancel()
	if err := kadDht.Provide(tctx, rendezvousPoint, true); err != nil {
		panic(err)
	}

	// Search in the background permanently for peers at the rendezvous point
	go func() {
		for {

			// 'FindProviders' will return 'PeerInfo' of all the peers which
			// have 'Provide' or announced themselves previously.
			tctx, cancel = context.WithTimeout(ctx, time.Second*10)
			peers, err := kadDht.FindProviders(tctx, rendezvousPoint)
			if err != nil {
				panic(err)
			}

			// Check all returned peers at the rendezvous point
			for _, p := range peers {

				// TODO: Find out why out host has never an address provided or returned

				// Ignore our host and peers without address
				if p.ID == chat.ID() || len(p.Addrs) == 0 {
					// No sense connecting to ourselves or if addrs are not available
					continue
				}

				// Check, if the peer already is known for writing to
				exists := false
				for _, writeConnection := range writeToPeers {
					if writeConnection.Pretty() == p.ID.Pretty() {
						exists = true
					}
				}

				// Create a stream for a new peer
				if !exists {

					stream, err := chat.NewStream(ctx, p.ID, "/chat/1.1.0")
					if err != nil {
						log.Printf("ERROR (%s): %v\n", p.ID, err)
					} else {

						// Add new remote peer as peer to write to
						writeToPeers = append(writeToPeers, stream.Conn().RemotePeer())

						// Create a buffer stream for non blocking read and write
						rw := bufio.NewReadWriter(bufio.NewReader(stream), bufio.NewWriter(stream))
						// Add new buffer to write to
						readWriters = append(readWriters, rw)

						// Go routine to process stream lines
						go readData(rw)

						// TODO: Unclear, if we need this Go routine and how many
						go writeData()

						// Shows the number of saved peers and connections, respectively
						fmt.Printf("\rInbound Connections: %d Outbound Connections: %d ",
							len(readFromPeers), len(writeToPeers))
					}
				}
			}
			// Instead of calling "defer cancel()" inside the for loop
			cancel()
		}
	}()

	// Keep the chat running
	select {}
}
