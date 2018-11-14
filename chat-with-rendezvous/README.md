
After two days I want to show something...

![A chat draft](/images/libp2pNineChats.png)
![with chat commands](/images/libp2pChatCommands.png)

# Initially
go build chat.go && uuidgen > uuid.txt && ./chat -r $(cat uuid.txt)

# Next chats
./chat -r $(cat uuid.txt)

