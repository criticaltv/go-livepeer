package core

import (
	"context"

	"github.com/golang/glog"
	"github.com/livepeer/libp2p-livepeer/eth"
	"github.com/livepeer/libp2p-livepeer/net"
	"github.com/livepeer/lpms/stream"
)

//NodeID can be converted from libp2p PeerID.
type NodeID string

type LivepeerNode struct {
	Identity     NodeID
	VideoNetwork net.VideoNetwork
	StreamDB     *StreamDB
	Eth          eth.Client
	IsTranscoder bool
}

func NewLivepeerNode() *LivepeerNode {
	return &LivepeerNode{StreamDB: NewStreamDB()}
}

func (n *LivepeerNode) Start() {
	//Connect to bootstrap node, ask for more peers

	//Connect to peers

	//Kick off process to periodically monitor peer connection by pinging them
}

//CreateTranscodeJob creates the onchain transcode job
//This can only be done by a broadcaster
func (n *LivepeerNode) CreateTranscodeJob( /*stream information + transcode config*/ ) {
	//Verify the stream exists(assume it's a local stream)

	//Call eth client to create the job
}

//Monitor the smart contract for job creation (as a transcoder)
func (n *LivepeerNode) monitorEth() {

}

//StartTranscodeJob starts a transcode job, and sends the transcoded streamIDs to the broadcaster
func (n *LivepeerNode) StartTranscodeJob() {
	//Start transcode jobs using the given config (async)

	//Send the streamIDs to the broadcaster

	//Subscribes to the original stream
}

//BroadcastToNetwork is called when a new broadcast stream is available.  It lets the network decide how
//to deal with the stream.
func (n *LivepeerNode) BroadcastToNetwork(ctx context.Context, strm *stream.VideoStream) error {
	b := n.VideoNetwork.NewBroadcaster(strm.GetStreamID())

	//Prepare the broadcast.  May have to send the MasterPlaylist as part of the handshake.

	//Kick off a go routine to broadcast the stream
	go func() {
		for {
			seg, err := strm.ReadHLSSegment()
			if err != nil {
				glog.Errorf("Error reading hls stream while broadcasting to network: %v", err)
				return //TODO: Should better handle error here
			}

			//Encode seg into []byte, then send it via b.Broadcast
		}
	}()

	select {
	case <-ctx.Done():
		glog.Errorf("Done Broadcasting")
		return nil
	}
}

//SubscribeFromNetwork subscribes to a stream on the network.  Returns the stream as a reference.
func (n *LivepeerNode) SubscribeFromNetwork(ctx context.Context, strmID StreamID) (*stream.VideoStream, error) {
	s := n.VideoNetwork.GetSubscriber(strmID.String())
	if s == nil {
		s = n.VideoNetwork.NewSubscriber(strmID.String())
	}

	//Create a new video stream
	strm := stream.NewVideoStream(strmID.String(), stream.HLS)
	err := s.Subscribe(ctx, func(seqNo uint64, data []byte) {
		//Check for segNo, decode data into HLSSegment, then write it to the stream.
		// strm.WriteHLSSegmentToStream()
	})
	if err != nil {
		glog.Errorf("Error subscribing from network: %v", err)
		return nil, err
	}
	return strm, nil
}

//UnsubscribeFromNetwork unsubscribes to a stream on the network.
func (n *LivepeerNode) UnsubscribeFromNetwork(strmID StreamID) error {
	s := n.VideoNetwork.GetSubscriber(strmID.String())
	if s == nil {
		glog.Error("Error unsubscribing from network - cannot find subscriber")
		return ErrNotFound
	}

	err := s.Unsubscribe()
	if err != nil {
		glog.Errorf("Error unsubscribing from network: %v", err)
		return err
	}

	return nil
}
