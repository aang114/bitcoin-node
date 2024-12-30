package networking

import (
	"encoding/hex"
	"errors"
	"fmt"
	"github.com/aang114/bitcoin-node/constants"
	"github.com/aang114/bitcoin-node/message"
	"log"
	"net"
	"os"
	"sync"
	"sync/atomic"
	"time"
)

var ErrNodeHasNoPeersOrUnconnectedAddrs = errors.New("node has no peers or unconnected addresses")

type ErrSendGetAddrMsgFailed struct {
	Peer *Peer
}

func (e ErrSendGetAddrMsgFailed) Error() string {
	return fmt.Sprintf("sendGetAddrMsg() failed for peer : %s", e.Peer.conn.RemoteAddr())
}

type InvPayloadWithSender struct {
	InvPayload *message.InvPayload
	Sender     *Peer
}

type BlockPayloadWithSender struct {
	BlockPayload *message.BlockPayload
	Sender       *Peer
}

type Node struct {
	mu                  sync.RWMutex
	protocolVersion     uint32
	services            message.Services
	minimumPeers        int
	tickerDuration      time.Duration
	tcpDialTimeout      time.Duration
	getAddrWaitTime     time.Duration
	blocksFileDirectory string
	peers               *SafeMap[*Peer, struct{}]
	connectedAddrs      *SafeMap[TCPAddress, struct{}]
	unconnectedAddrs    *SafeMap[TCPAddress, struct{}]
	blocks              *SafeSlice[*message.BlockPayload]
	blockHashes         *SafeMap[message.Hash256, struct{}]
	HasQuit             bool
	QuitCh              chan struct{}
	addPeersCh          chan struct{}
	invMsgCh            chan *InvPayloadWithSender
	blockMsgCh          chan *BlockPayloadWithSender
}

func NewNode(
	protocolVersion uint32,
	services message.Services,
	minimumPeers int,
	blocksFileDirectory string,
	tickerDuration time.Duration,
	tcpDialTimeout time.Duration,
	getAddrWaitTime time.Duration,
) *Node {
	n := Node{
		protocolVersion:     protocolVersion,
		services:            services,
		minimumPeers:        minimumPeers,
		tickerDuration:      tickerDuration,
		tcpDialTimeout:      tcpDialTimeout,
		getAddrWaitTime:     getAddrWaitTime,
		blocksFileDirectory: blocksFileDirectory,
		peers:               NewSafeMap[*Peer, struct{}](),
		connectedAddrs:      NewSafeMap[TCPAddress, struct{}](),
		unconnectedAddrs:    NewSafeMap[TCPAddress, struct{}](),
		blocks:              NewSafeSlice[*message.BlockPayload](0),
		blockHashes:         NewSafeMap[message.Hash256, struct{}](),
		HasQuit:             false,
		QuitCh:              make(chan struct{}),
		addPeersCh:          make(chan struct{}, 1),
		// TODO - Decide on the channel buffer length
		invMsgCh: make(chan *InvPayloadWithSender, minimumPeers),
		// TODO - Decide on the channel buffer length
		blockMsgCh: make(chan *BlockPayloadWithSender, minimumPeers),
	}

	return &n
}

func (n *Node) Start() {
	err := n.readBlocksFromDisk()
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			log.Printf("File %s does not exist. Starting afresh...", n.blocksFileDirectory)
		} else {
			log.Printf("⚠️ Couldn't read the blocks in file %s due to error: %s. Quitting now...", n.blocksFileDirectory, err)
			n.Quit()
			return
		}
	} else {
		log.Printf("💾 Successfully read %d blocks in file %s", n.blocks.Len(), n.blocksFileDirectory)
	}

	if n.peers.Len() < n.minimumPeers {
		n.notifyThatPeersIsBelowMinPeers()
	}

	n.selectLoop()
}

func (n *Node) AddPeer(remoteAddr *net.TCPAddr, receivingServices message.Services) (*Peer, error) {
	conn, err := PerformHandshake(remoteAddr, n.tcpDialTimeout, n.services, receivingServices)
	if err != nil {
		return nil, err
	}
	onQuitting := func(peerNode *Peer) { n.removePeerFromNode(peerNode) }
	p, err := NewPeer(conn, onQuitting, n.invMsgCh, n.blockMsgCh)
	if err != nil {
		return nil, err
	}
	n.addPeerToNode(p)
	go p.Start()
	return p, nil
}

func (n *Node) Quit() {
	n.mu.Lock()
	defer n.mu.Unlock()

	log.Printf("Quitting Node...")

	if n.HasQuit {
		return
	}
	n.HasQuit = true

	for _, peer := range n.peers.Keys() {
		peer.Quit()
	}

	close(n.QuitCh)

	err := n.saveBlocksToDisk()
	if err != nil {
		log.Printf("⚠️ Could not save blocks due to error: %s", err)
	} else {
		log.Printf("💾 Successfully saved blocks to file %s", n.blocksFileDirectory)
	}
}

func (n *Node) selectLoop() {
	ticker := time.NewTicker(n.tickerDuration)

	for {
		select {
		case <-n.QuitCh:
			log.Printf("[selectLoop] Node's QuitCh was closed")
			return
		case <-ticker.C:
			log.Printf("[selectLoop] Executing handleTickerResponse()...")
			err := n.handleTickerResponse()
			if err != nil {
				log.Printf("[selectLoop] handleTickerResponse() failed with error %s", err)
			} else {
				log.Printf("[selectLoop] handleTickerResponse() executed successfully")
			}
		case _ = <-n.addPeersCh:
			log.Printf("[selectLoop] Executing handleAddPeersChResponse()...")
			err := n.handleAddPeersChResponse()
			if err != nil {
				log.Printf("[selectLoop] handleAddPeersChResponse() failed with error %s", err)
				sendGetAddrFailed := &ErrSendGetAddrMsgFailed{}
				if errors.As(err, sendGetAddrFailed) {
					log.Printf("[selectLoop] Quitting peer %s because sending it did not reply to getaddr msg in time", sendGetAddrFailed.Peer.conn.RemoteAddr())
					sendGetAddrFailed.Peer.Quit()
				} else if errors.Is(err, ErrNodeHasNoPeersOrUnconnectedAddrs) {
					log.Printf("[selectLoop] Quitting node due to error %s", err)
					n.Quit()
				}
			} else {
				log.Printf("[selectLoop] handleAddPeersChResponse() executed successfully")
			}
		case invMsg := <-n.invMsgCh:
			log.Printf("[selectLoop] Executing handleInvMsg()...")
			err := n.handleInvMsg(invMsg)
			if err != nil {
				log.Printf("[selectLoop] Quitting peer %s due to error %s", invMsg.Sender.conn.RemoteAddr(), err)
				invMsg.Sender.Quit()
			} else {
				log.Printf("[selectLoop] handleInvMsg() executed successfully")
			}
		case blockMsg := <-n.blockMsgCh:
			log.Printf("[selectLoop] Executing handleBlockMsg()...")
			err := n.handleBlockMsg(blockMsg)
			if err != nil {
				log.Printf("[selectLoop] Quitting peer %s due to error %s", blockMsg.Sender.conn.RemoteAddr(), err)
				blockMsg.Sender.Quit()
			} else {
				log.Printf("[selectLoop] handleBlockMsg() executed successfully")
			}
		}

	}
}

func (n *Node) handleTickerResponse() error {
	missingBlocksHashes, err := n.getMissingBlocksHashes()
	if err != nil {
		return err
	}
	if len(missingBlocksHashes) > 0 {
		randomPeer, ok := n.peers.GetRandomKey()
		if !ok {
			return nil
		}
		return n.sendGetBlockDataMsg(randomPeer, missingBlocksHashes)
	}

	err = n.requestForNewBlocks()
	return err
}

func (n *Node) requestForNewBlocks() error {
	latestBlockHash := message.Hash256(constants.GenesisBlockHash)
	var err error
	if length := n.blocks.Len(); length > 0 {
		latestBlockHash, err = n.getLatestBlockHash()
		if err != nil {
			return err
		}
	}
	log.Printf("sending getblocks message with latest block 0x%s", hex.EncodeToString(latestBlockHash[:]))
	zeroBlockHash := message.Hash256{}
	randomPeer, ok := n.peers.GetRandomKey()
	if !ok {
		return nil
	}
	// hashStop set to zero to get as many blocks as possible (500)
	return n.sendGetBlocksMsg(randomPeer, []message.Hash256{latestBlockHash}, zeroBlockHash)
}

func (n *Node) handleAddPeersChResponse() error {
	return n.addPeersIfNecessary()
}

func (n *Node) handleInvMsg(i *InvPayloadWithSender) error {
	blockHashes := make([]message.Hash256, 0)

	for _, inventory := range i.InvPayload.InventoryList {
		if inventory.Type == message.MsgBlock || inventory.Type == message.MsgWitnessBlock {
			if _, ok := n.blockHashes.Get(inventory.Hash); !ok {
				blockHashes = append(blockHashes, inventory.Hash)
			}
		}
	}

	log.Printf("%d blocks found in inv message sent by peer %s", len(blockHashes), i.Sender.conn.RemoteAddr())

	if len(blockHashes) == 0 {
		return nil
	}

	return n.sendGetBlockDataMsg(i.Sender, blockHashes)
}

func (n *Node) handleBlockMsg(msg *BlockPayloadWithSender) error {
	blockHash, err := msg.BlockPayload.GetBlockHash()
	if err != nil {
		return err
	}
	log.Printf("Received Block 0x%s from peer %s", hex.EncodeToString(blockHash[:]), msg.Sender.conn.RemoteAddr())
	err = n.addBlockToNode(msg.BlockPayload)
	if err != nil {
		return err
	}

	missingBlockHashes, err := n.getMissingBlocksHashes()
	if err != nil {
		return err
	}
	log.Printf("There are %d missing blocks", len(missingBlockHashes))
	if len(missingBlockHashes) == 0 {
		return nil
	}

	//randomPeer, ok := n.peers.GetRandomKey()
	//if !ok {
	//	return nil
	//}
	//log.Printf("Requesting %d missing blocks from peer %s", len(missingBlockHashes), randomPeer.conn.RemoteAddr())
	//return n.sendGetBlockDataMsg(randomPeer, missingBlockHashes)

	// since we know msg.Sender is historically responsive to "inv" requests, let's ask it for the missing blocks rather than a random peer
	return n.sendGetBlockDataMsg(msg.Sender, missingBlockHashes)
}

func (n *Node) saveBlocksToDisk() error {
	blocks := n.blocks.GetAll()
	if len(blocks) == 0 {
		return errors.New("no blocks to write to file")
	}

	f, err := os.Create(fmt.Sprintf("/tmp/%s", n.blocksFileDirectory))
	if err != nil {
		return err
	}
	defer f.Close()

	blocksCountEncoded, err := message.VarInt(len(blocks)).Encode()
	if err != nil {
		return err
	}
	_, err = f.Write(blocksCountEncoded)
	if err != nil {
		return err
	}

	for _, block := range blocks {
		blockEncoded, err := block.Encode()
		if err != nil {
			return err
		}
		_, err = f.Write(blockEncoded)
		if err != nil {
			return err
		}
	}

	if err != nil {
		return err
	}

	return os.Rename(fmt.Sprintf("/tmp/%s", n.blocksFileDirectory), n.blocksFileDirectory)
}

func (n *Node) readBlocksFromDisk() error {
	f, err := os.Open(n.blocksFileDirectory)
	if err != nil {
		return err
	}
	defer f.Close()

	blocksCount, err := message.DecodeVarInt(f)
	if err != nil {
		return err
	}
	blocks := make([]*message.BlockPayload, blocksCount)
	for i := range blocksCount {
		block, err := message.DecodeBlockPayload(f)
		if err != nil {
			return err
		}
		blocks[i] = block
	}

	for _, block := range blocks {
		err := n.addBlockToNode(block)
		if err != nil {
			return err
		}
	}

	return nil
}

func (n *Node) addPeersIfNecessary() error {
	if n.peers.Len() == 0 && n.unconnectedAddrs.Len() == 0 {
		n.Quit()
		return ErrNodeHasNoPeersOrUnconnectedAddrs
	}

	if n.peers.Len() >= n.minimumPeers {
		return nil
	}

	log.Printf("⚠️ Warning: Node is currently below the minimum peers required (Current peers count: %d)", n.peers.Len())

	connectionsToAdd := n.minimumPeers - n.peers.Len()

	log.Printf("Requesting for %d new addresses", connectionsToAdd)

	if randomPeer, ok := n.peers.GetRandomKey(); ok && n.unconnectedAddrs.Len() < connectionsToAdd {
		getAddrResponseCh, err := n.sendGetAddrMsg(randomPeer)
		if err != nil {
			return err
		}
		var addresses []message.Address
		// times out if a response is not gotten in `n.getAddrWaitTime` seconds
		select {
		case a := <-getAddrResponseCh:
			addresses = a
		case <-time.After(n.getAddrWaitTime):
			addresses = nil
		}
		for _, address := range addresses {
			tcpAddress := TCPAddress{IpAddress: [16]byte(address.NetworkAddress.IpAddress.To16()), Port: address.NetworkAddress.Port}
			n.addUnconnectedAddrToNode(tcpAddress)
		}
	}

	log.Printf("Connecting to new peers until min peers reached (Current peers count: %d)", n.peers.Len())

	// the error rate for dialing with new peers is very high. that's why we try to connect with 10 times the minimum peers required
	maxNewPeers := n.minimumPeers * 10
	successCount := n.attemptAddingSomePeers(maxNewPeers)
	log.Printf("Successfully added %d new peers", successCount)
	if n.peers.Len() < n.minimumPeers {
		n.notifyThatPeersIsBelowMinPeers()
		log.Printf("Could not connect until min peers reached (Current peers count: %d)", n.peers.Len())
	} else {
		log.Printf("🎯 Successfully connected until min peers reached (Current peer count: %d)", n.peers.Len())
	}

	return nil
}

func (n *Node) sendGetAddrMsg(peer *Peer) (<-chan []message.Address, error) {
	getAddrResponseCh, err := peer.sendGetAddrMsg()
	if err != nil {
		return nil, err
	}

	return getAddrResponseCh, nil
}

func (n *Node) sendGetBlocksMsg(peer *Peer, blockLocatorHashes []message.Hash256, hashStop message.Hash256) error {
	return peer.sendGetBlocksMsg(n.protocolVersion, blockLocatorHashes, hashStop)
}

func (n *Node) sendGetBlockDataMsg(peer *Peer, blockHashes []message.Hash256) error {
	blockInventories := make([]message.Inventory, len(blockHashes))
	for i, blockHash := range blockHashes {
		blockInventories[i] = message.Inventory{Type: message.MsgBlock, Hash: blockHash}
	}

	return peer.sendGetBlockDataMsg(blockInventories)
}

func (n *Node) attemptAddingSomePeers(maxNewPeers int) uint64 {
	var successCount atomic.Uint64

	var wg sync.WaitGroup
	for _ = range maxNewPeers {
		unconnectedAddr, ok := n.unconnectedAddrs.Pop()
		if !ok {
			break
		}

		wg.Add(1)
		go func() {
			defer wg.Done()
			_, err := n.AddPeer(&net.TCPAddr{IP: unconnectedAddr.IpAddress[:], Port: int(unconnectedAddr.Port)}, message.NodeNetwork)
			if err != nil {
				log.Printf("❌ Could not add peer %s due to error: %s (Current peer count: %d)", unconnectedAddr.String(), err, n.peers.Len())
			} else {
				successCount.Add(1)
			}
		}()
	}
	wg.Wait()

	return successCount.Load()
}

func (n *Node) addPeerToNode(peerNode *Peer) {
	n.peers.Set(peerNode, struct{}{})
	n.connectedAddrs.Set(peerNode.tcpAddress, struct{}{})
	n.unconnectedAddrs.Delete(peerNode.tcpAddress)
}

func (n *Node) removePeerFromNode(peerNode *Peer) {
	n.peers.Delete(peerNode)
	n.connectedAddrs.Delete(peerNode.tcpAddress)

	log.Printf("⬇️ Removing peer %s from node (Current peers count: %d)", peerNode.conn.RemoteAddr(), n.peers.Len())

	if n.peers.Len() < n.minimumPeers {
		n.notifyThatPeersIsBelowMinPeers()
	}
}

func (n *Node) addUnconnectedAddrToNode(unconnectedAddr TCPAddress) {
	if _, ok := n.connectedAddrs.Get(unconnectedAddr); !ok {
		n.unconnectedAddrs.Set(unconnectedAddr, struct{}{})
	}
}

func (n *Node) notifyThatPeersIsBelowMinPeers() {
	select {
	case n.addPeersCh <- struct{}{}:
	default:
		log.Println("addPeersCh has already been notified")
	}
}

func (n *Node) addBlockToNode(block *message.BlockPayload) error {
	blockHash, err := block.GetBlockHash()
	if err != nil {
		return err
	}
	if _, ok := n.blockHashes.Get(blockHash); ok {
		return nil
	}

	n.blockHashes.Set(blockHash, struct{}{})
	n.blocks.Append(block)

	log.Printf("️➕ Added block 0x%s to node", hex.EncodeToString(blockHash[:]))

	return nil
}

func (n *Node) getMissingBlocksHashes() ([]message.Hash256, error) {
	missingBlocks := make([]message.Hash256, 0)
	// genesis block's previous block
	zeroBlockHash := message.Hash256{}

	for _, block := range n.blocks.GetAll() {
		if _, ok := n.blockHashes.Get(block.PrevBlock); !ok && block.PrevBlock != zeroBlockHash {
			missingBlocks = append(missingBlocks, block.PrevBlock)
		}
	}

	return missingBlocks, nil
}

// TODO - Improve (this is very inefficient in the long term since it iterates over every block)
func (n *Node) getLatestBlockHash() (message.Hash256, error) {
	var latestBlock *message.BlockPayload
	latestTimestamp := uint32(0)

	for _, block := range n.blocks.GetAll() {
		if block.Timestamp > latestTimestamp {
			latestTimestamp = block.Timestamp
			latestBlock = block
		}
	}

	if latestBlock == nil {
		return message.Hash256{}, errors.New("No blocks exist")
	}

	return latestBlock.GetBlockHash()
}
