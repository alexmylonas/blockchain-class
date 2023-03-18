package state

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"

	"github.com/ardanlabs/blockchain/foundation/blockchain/database"
	"github.com/ardanlabs/blockchain/foundation/blockchain/peer"
)

func (s *State) NetRequestPeerStatus(pr peer.Peer) (peer.PeerStatus, error) {
	s.evHandler("state: NetRequestPeerStatus: started for peer %s", pr.Host)
	defer s.evHandler("state: NetRequestPeerStatus: completed for peer %s", pr.Host)

	statusUrl := pr.Url() + peer.StatusUri
	var ps peer.PeerStatus
	if err := send(http.MethodGet, statusUrl, nil, &ps); err != nil {
		return peer.PeerStatus{}, err
	}

	s.evHandler("state: NetRequestPeerStatus: peer-node[%s]: latestBlkNum [%s]: knowPeers [%s]", pr.Host, ps.LatestBlockNum, ps.KnownPeers)

	return ps, nil
}

func (s *State) NetRequestMempool(pr peer.Peer) ([]database.BlockTx, error) {
	s.evHandler("stae: NetRequestMempool : started for peer %s", pr.Host)
	defer s.evHandler("state: NetRequestMempool: completed for peer %s", pr.Host)

	mempoolUrl := pr.Url() + peer.MempoolUri

	var mempool []database.BlockTx
	if err := send(http.MethodGet, mempoolUrl, nil, &mempool); err != nil {
		return nil, err
	}

	s.evHandler("state: NetRequestMempool: peer-node[%s]: txs [%d]", pr.Host, len(mempool))

	return mempool, nil
}

func (s *State) NetRequestPeerBlocks(pr peer.Peer) error {
	s.evHandler("stae: NetRequestPeerBlocks : started for peer %s", pr.Host)
	defer s.evHandler("state: NetRequestPeerBlocks: completed for peer %s", pr.Host)

	// CORE NOTE: Ideally, you want to start by pulling just block headers and performing the crpytographic
	// aduit so you can verify the integrity of the blocks and to validate you are not being attacked.
	// After that you can start pulling the full block data for each block header if you are a full node.
	// And maybe only the last 100 blocks should be calculated If you are a light node.
	// i.e. getBlockHeaders
	// Currently, the Ardan blockchain is a full node only and needs the transactions to have a comple account database.
	// The cryptographic audit takes place at each full block is downloead from peers.

	from := strconv.FormatUint(s.LatestBlock().Header.Number+1, 10)
	blocksUrl := fmt.Sprintf(pr.Url()+peer.BlocksUri, from, "latest")

	// Getting all blocks from current block until peer's latest block.
	var blocksData []database.BlockData
	if err := send(http.MethodGet, blocksUrl, nil, &blocksData); err != nil {
		return err
	}

	s.evHandler("state: NetRequestPeerBlocks: peer-node[%s]: blocks [%d]", pr.Host, len(blocksData))

	for _, blockData := range blocksData {
		block, err := database.ToBlock(blockData)
		if err != nil {
			return err
		}
		if err := s.ProcessProposedBlock(block); err != nil {
			return err
		}
	}
	return nil
}

func (s *State) NetSendNodeAvailableToPeers() {
	s.evHandler("state: NetSendNodeAvailableToPeers: started")
	defer s.evHandler("state: NetSendNodeAvailableToPeers: completed")

	host := peer.New(s.Host())

	for _, pr := range s.KnowExternalPeers() {
		s.evHandler("state: NetSendNodeAvailableToPeers: sending to peer %s", pr.Host)
		peerUrl := pr.Url() + peer.PeerUri
		if err := send(http.MethodPost, peerUrl, host, nil); err != nil {
			s.evHandler("state: NetSendNodeAvailableToPeers: error sending to peer %s: %s", pr.Host, err)
		}
	}
}

func send(method string, url string, dataSend any, dataRecv any) error {
	var req *http.Request

	switch {
	case dataSend != nil:
		data, err := json.Marshal(dataSend)
		if err != nil {
			return err
		}
		req, err = http.NewRequest(method, url, bytes.NewReader(data))
		if err != nil {
			return err
		}

	default:
		var err error
		req, err = http.NewRequest(method, url, nil)
		if err != nil {
			return err
		}
	}

	var client http.Client
	resp, err := client.Do(req)
	if err != nil {
		return err
	}

	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil
	}

	if resp.StatusCode != http.StatusOK {
		msg, err := io.ReadAll(resp.Body)
		if err != nil {
			return err
		}
		return errors.New(string(msg))
	}

	if dataRecv != nil {
		return json.NewDecoder(resp.Body).Decode(dataRecv)
	}

	return nil
}
