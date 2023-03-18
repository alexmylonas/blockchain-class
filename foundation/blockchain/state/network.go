package state

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"net/http"

	"github.com/ardanlabs/blockchain/foundation/blockchain/peer"
)

func (s *State) NetRequestPeerStatus(pr peer.Peer) (peer.PeerStatus, error) {
	s.evHandler("state: NetRequestPeerStatus: started for peer %s", pr.Host)
	defer s.evHandler("state: NetRequestPeerStatus: completed for peer %s", pr.Host)

	statusUrl := pr.StatusUrl()
	var ps peer.PeerStatus
	if err := send(http.MethodGet, statusUrl, nil, &ps); err != nil {
		return peer.PeerStatus{}, err
	}

	s.evHandler("state: NetRequestPeerStatus: peer-node[%s]: latestBlkNum [%s]: knowPeers [%s]", pr.Host, ps.LatestBlockNum, ps.KnownPeers)

	return ps, nil
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
