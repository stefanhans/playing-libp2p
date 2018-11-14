I started to play a bit with the [chat-with-rendezvous](https://github.com/libp2p/go-libp2p-examples/tree/master/chat-with-rendezvous) :satisfied:

After two days I want to show something...

I have tested my prototype with nine chats in parallel. It works, but is not always getting all peers connected :weary:.

![A chat draft](/images/libp2pNineChats.png)


I added some commands executable from within the chat.

![with chat commands](/images/libp2pChatCommands.png)

# Usage

First chat:

```bash
go build chat.go && uuidgen > uuid.txt && ./chat -r $(cat uuid.txt)
```

The other chats:

```bash
# Next chats
./chat -r $(cat uuid.txt)
```

