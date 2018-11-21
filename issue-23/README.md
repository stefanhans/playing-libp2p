`go run main.go -r $(uuidgen)`

`host.Peerstore().SetAddrs(host.ID(), host.Addrs(), peerstore.ConnectedAddrTTL)` makes the difference

 
 
