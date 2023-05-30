package main

import (
	cryptorand "crypto/rand"
	"fmt"
	"math"
	"math/big"
)

const numReplicas = 3

func main() {
	bigint, err := cryptorand.Int(cryptorand.Reader, big.NewInt(math.MaxInt64))
	if err != nil {
		panic(fmt.Errorf("generating seed: %w", err))
	}
	seed := bigint.Uint64()

	rand := NewRand(seed)

	networkConfig := NetworkConfig{
		PathClogProbability:      0.05,
		MessageReplayProbability: 0.05,
		DropMessageProbability:   0.05,
		MaxNetworkPathClogTicks:  100,
		MaxMessageDelayTicks:     100,
	}

	replicas := make([]*Replica, 0, numReplicas)

	for id := 0; id < numReplicas; id++ {
		replicas = append(replicas, NewReplica(int16(id), nil))
	}

	network := NewNetwork(networkConfig, replicas, rand)
	for _, replica := range replicas {
		replica.network = network
	}

	replicas[0].ping(1)

	for i := 0; i < 3; i++ {
		fmt.Printf("- tick %d\n", i)

		network.tick()

		for _, replica := range replicas {
			fmt.Printf("-- replica %d\n", replica.id)
			replica.tick()
		}
	}
}
