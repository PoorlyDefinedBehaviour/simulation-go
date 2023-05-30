package main

import "fmt"

type NetworkConfig struct {
	PathClogProbability      float64
	MessageReplayProbability float64
	DropMessageProbability   float64
	MaxNetworkPathClogTicks  uint64
	MaxMessageDelayTicks     uint64
}

type Network struct {
	config NetworkConfig

	rand Rand

	ticks uint64

	replicas []*Replica

	// The network path from replica A to replica B.
	networkPaths []NetworkPath

	// Messages that need to be sent to a replica.
	sendMessageQueue PriorityQueue
}

type NetworkPath struct {
	FromReplicaID          ReplicaID
	ToReplicaID            ReplicaID
	MakeReachableAfterTick uint64
}

func newNetworkPath(fromReplicaID, toReplicaID ReplicaID) NetworkPath {
	return NetworkPath{
		FromReplicaID:          fromReplicaID,
		ToReplicaID:            toReplicaID,
		MakeReachableAfterTick: 0,
	}
}

type MessageToSend struct {
	AfterTick     uint64
	FromReplicaID ReplicaID
	Message       Message
	Index         int
}

func NewNetwork(config NetworkConfig, replicas []*Replica, rand Rand) *Network {
	return &Network{
		config:       config,
		replicas:     replicas,
		rand:         rand,
		networkPaths: buildNetworkPaths(replicas),
	}
}

func buildNetworkPaths(replicas []*Replica) []NetworkPath {
	paths := make([]NetworkPath, 0)

	for _, fromReplica := range replicas {
		for _, toReplica := range replicas {
			if fromReplica == toReplica {
				continue
			}

			paths = append(paths, newNetworkPath(fromReplica.id, toReplica.id))
		}
	}

	return paths
}

func (network *Network) send(fromReplicaID ReplicaID, message Message) {
	messageToSend := &MessageToSend{
		AfterTick:     network.randomDelay(),
		FromReplicaID: fromReplicaID,
		Message:       message,
	}
	fmt.Printf("network: will send message from replica %d to replica %d at tick %d\n",
		fromReplicaID,
		message.ReplicaID,
		messageToSend.AfterTick,
	)
	network.sendMessageQueue.Push(messageToSend)
}

func (network *Network) randomDelay() uint64 {
	return network.ticks + network.rand.genBetween(0, network.config.MaxMessageDelayTicks)
}

func (network *Network) tick() {
	network.ticks++
	fmt.Printf("network is at tick %d\n", network.ticks)

	for i := range network.networkPaths {
		shouldMakeUnreachable := network.rand.genBool(network.config.PathClogProbability)
		if shouldMakeUnreachable {
			network.networkPaths[i].MakeReachableAfterTick = network.rand.genBetween(0, network.config.MaxNetworkPathClogTicks)
		}
	}

	fmt.Printf("\n\naaaaaaa network.networkPaths %+v\n\n", network.networkPaths)

	for len(network.sendMessageQueue) > 0 {
		oldestMessage := network.sendMessageQueue.Pop().(*MessageToSend)
		if oldestMessage.AfterTick > network.ticks {
			network.sendMessageQueue.Push(oldestMessage)
			return
		}

		networkPath := network.findPath(oldestMessage.FromReplicaID, oldestMessage.Message.ReplicaID)
		if networkPath.MakeReachableAfterTick > network.ticks {
			network.sendMessageQueue.Push(oldestMessage)
			return
		}

		shouldDrop := network.rand.genBool(network.config.DropMessageProbability)
		if shouldDrop {
			continue
		}

		if oldestMessage.AfterTick < network.ticks {
			network.replicas[oldestMessage.Message.ReplicaID].onMessageReceived(oldestMessage.FromReplicaID, oldestMessage.Message)
		}

		shouldReplay := network.rand.genBool(network.config.MessageReplayProbability)
		if shouldReplay {
			network.sendMessageQueue.Push(oldestMessage)
		}
	}
}

func (network *Network) findPath(fromReplicaID, toReplicaID ReplicaID) NetworkPath {
	for _, path := range network.networkPaths {
		if path.FromReplicaID == fromReplicaID && path.ToReplicaID == toReplicaID {
			return path
		}
	}
	panic("unreachable")
}
