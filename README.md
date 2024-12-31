# Bitcoin Node

## Solution

This repository contains a proof-of-concept Go implementation of a Bitcoin Node, where the first 3 [milestones](#Milestone) of the task are completed. This node is also able to gracefully shutdown. Furthemore, it stores all its received blocks to disk while shutting down and also reads from disk for blocks while starting up.

### Running the Node

To run the node, run the following command:

```shell
go build main.go && ./main
```

#### Optional Flags

```shell
Usage of ./main:
  -minPeers int
        Minimum Number of Peers that the Node must be connected with at all times (default 5)
  -peer string
        First Peer to Connect with (default "46.166.142.2:8333")
```

### Implementation

At runtime, an instance of a `Node` struct (which represents the bitcoin node) is created which maintains a list of active peers (where each peer is represented as a `Peer` struct).

The `Node` struct instance runs a loop (`Node.selectLoop()`)  where it listens to different communication channels using a `select` statement.

The different communication channels it listens to are as follows:

- `Node.addPeersCh` channel: This channel is used to notify the node that its current list of active peers has fallen below the minimum number of active peers required.
- `Node.invMsgCh` channel: This channel is used by the node's active peers to send ["inv" messages](https://en.bitcoin.it/wiki/Protocol_documentation#inv) to the node.
- `Node.blockMsgCh` channel: This channel is used by the node's active peers to send ["block" messages](https://en.bitcoin.it/wiki/Protocol_documentation#block) to the node.
- Ticker Channel: This channel is triggered every `Node.tickerDuration` seconds and is used by the node to request for new blocks from its active peer(s).
- `Node.QuitCh`: This channel notifies the node that it had been quit.



## Task

### Requirements

- The implementation should be written in Go and compile at least on linux
- The solution cannot use existing P2P libraries

### Milestones

#### Milestone 1:

- The solution has to perform a full protocol-level (post-TCP) handshake with the target node you can choose an available BTC Nodes for connection: <https://bitnodes.io/>
- You can follow the specification here: <https://en.bitcoin.it/>
- Can not use the node implementation as a dependency
- You can ignore any post-handshake traffic from the target node, and it doesn't have to keep the connection alive.
- Make sure to sync from main chain

#### Milestone 2:

- Manage more than 1 connected node
- Need to perform a full protocol-level handshake with the new nodes
- You can find the specification here: <https://en.bitcoin.it/wiki/Protocol_documentation#addr> and <https://en.bitcoin.it/wiki/Protocol_documentation#getaddr>

#### Millestone 3:

- Request block informations from other peers, can be the full block data or just the headers
- Your node should be resilient to network failures (dial error, protocol not supported, incompatible version)
- Your node should check the response contents and ignore if the response doesn't contains what was requested, as well as to guarantee the chain consistency, the current should be father of the next one and so on and so forth
- No need for block or header validation, just retrieve and store should be enough

#### Milestone 4:

- Starting from the genesis block you must retrieve few blocks and must verify their transactions.
- You can find the specification here: <https://en.bitcoin.it/wiki/Protocol_documentation#Transaction_Verification>
- Bonus points if you implement your own Script (stack-based scripting system for transactions) validation, you can find the spec here <https://en.bitcoin.it/wiki/Script>
- The program can exit after validating the blocks, no need to keep syncing.

#### Milestone 5:

- This last milestone is a continuation of the Milestone 4
- The node should be able to keep syncing (retrieving, validating and importing) blocks until the tip of the chain
- The implemented node should be able to gracefully shutdown (preserving the current sync state) as well as able to resume from the latest point.

### Evaluation

- Quality: the solution should be idiomatic and adhere to Go coding conventions.
- Performance: the solution should be as fast as the handshake protocol allows, and it shouldn't block resources
- Security: the network is an inherently untrusted environment, and it should be taken into account.