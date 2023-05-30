package main

import "fmt"

type ReplicaID = int16

type MessageType = int8

const (
	PingMessageType MessageType = 1
	PongMessageType             = 2
)

type Message struct {
	ReplicaID int16
	Type      int8
	Payload   []byte
}

type Replica struct {
	id                   ReplicaID
	network              *Network
	sendMessageQueue     []Message
	receivedMessageQueue []Message
}

func NewReplica(id int16, network *Network) *Replica {
	return &Replica{id: id, network: network}
}

func (replica *Replica) tick() {
	fmt.Printf("replica %d: number of messages to send: %d\n", replica.id, len(replica.sendMessageQueue))
	fmt.Printf("replica %d: number of messages to process: %d\n", replica.id, len(replica.receivedMessageQueue))
	if message := replica.nextReceivedMessage(); message != nil {
		replica.processMessage(message)
	}

	if message := replica.nextMessageToSend(); message != nil {
		replica.network.send(replica.id, *message)
	}
}

func (replica *Replica) nextReceivedMessage() *Message {
	if len(replica.receivedMessageQueue) == 0 {
		return nil
	}

	message := replica.receivedMessageQueue[0]
	replica.receivedMessageQueue = replica.receivedMessageQueue[1:]

	return &message
}

func (replica *Replica) nextMessageToSend() *Message {
	if len(replica.sendMessageQueue) == 0 {
		return nil
	}

	message := replica.sendMessageQueue[0]
	replica.sendMessageQueue = replica.sendMessageQueue[1:]
	return &message
}

func (replica *Replica) ping(replicaID int16) {
	replica.sendMessageQueue = append(replica.sendMessageQueue, Message{
		ReplicaID: replicaID,
		Type:      PingMessageType,
	})
}

func (replica *Replica) processMessage(message *Message) {
	if message.Type == PingMessageType {
		replica.network.send(replica.id, Message{
			ReplicaID: message.ReplicaID,
			Type:      PongMessageType,
		})
	}

	if message.Type == PongMessageType {
		fmt.Printf("replica %d received pong message from replica %d\n", replica.id, message.ReplicaID)
	}
}

func (replica *Replica) onMessageReceived(fromReplicaID int16, message Message) {
	replica.receivedMessageQueue = append(replica.receivedMessageQueue, Message{
		ReplicaID: fromReplicaID,
		Type:      message.Type,
		Payload:   message.Payload,
	})
}
