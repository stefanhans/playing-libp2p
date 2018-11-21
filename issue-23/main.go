package main

import (
	"context"
	"flag"
	"fmt"
	"time"

	"github.com/ipfs/go-cid"
	"github.com/ipfs/go-ipfs-addr"
	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p-kad-dht"
	"github.com/libp2p/go-libp2p-peerstore"
	"github.com/multiformats/go-multihash"
)

func main() {
	rendezvousString := flag.String("r", "", "")
	flag.Parse()

	ctx := context.Background()

	host, err := libp2p.New(ctx)
	if err != nil {
		panic(err)
	}

	// Created host data
	fmt.Println("Created host.ID(): ", host.ID())
	fmt.Println("Created host.Addrs(): ", host.Addrs())

	// This makes the difference !!!
	host.Peerstore().SetAddrs(host.ID(), host.Addrs(), peerstore.ConnectedAddrTTL)

	kadDht, err := dht.New(ctx, host)
	if err != nil {
		panic(err)
	}

	addr, _ := ipfsaddr.ParseString("/ip4/104.131.131.82/tcp/4001/ipfs/QmaCpDMGvV2BGHeYERUEnRQAwe3N8SzbUtfsmvsqQLuvuJ")
	peerinfo, _ := peerstore.InfoFromP2pAddr(addr.Multiaddr())

	if err := host.Connect(ctx, *peerinfo); err != nil {
		fmt.Println(err)
	}

	v1b := cid.V1Builder{Codec: cid.Raw, MhType: multihash.SHA2_256}
	rendezvousPoint, _ := v1b.Sum([]byte(*rendezvousString))

	tctx, cancel := context.WithTimeout(ctx, time.Second*10)
	defer cancel()
	if err := kadDht.Provide(tctx, rendezvousPoint, true); err != nil {
		panic(err)
	}

	tctx, cancel = context.WithTimeout(ctx, time.Second*10)
	defer cancel()
	peers, err := kadDht.FindProviders(tctx, rendezvousPoint)
	if err != nil {
		panic(err)
	}

	for _, peer := range peers {

		// Matched peer data
		if peer.ID.String() == host.ID().String() {
			fmt.Println("Found peer.ID: ", peer.ID)
			fmt.Println("Found peer.Addrs: ", peer.Addrs)
		}
	}
}
