package client

import (
	"errors"
	"io"
	"log"

	"golang.org/x/net/context"
	"google.golang.org/grpc"

	"bitx/streamer/streamerpb"
)

// Client is a streamer gRPC client. It connects to a service which exposes
// the service described by streamerpb.proto.
type Client struct {
	pair      string
	rpcClient streamerpb.StreamerClient
	orderBook *OrderBook
	Quit      chan bool
}

// ErrInvalidAddress indicates the provided address is not a valid server
// host:port combination.
var ErrInvalidAddress = errors.New("invalid address")

// New returns a new client.
func New(pair string) *Client {
	cl := &Client{pair: pair, Quit: make(chan bool, 1)}
	return cl
}

// Connect connects to the gRPC server.
func (cl *Client) Connect(address string) error {
	if address == "" {
		return ErrInvalidAddress
	}

	var creds []grpc.DialOption
	// TODO(neil): TLS?
	creds = append(creds, grpc.WithInsecure())

	conn, err := grpc.Dial(address, creds...)
	if err != nil {
		return err
	}

	log.Printf("bitx/streamer/client.Connect(): Connected to %s.", address)

	cl.rpcClient = streamerpb.NewStreamerClient(conn)

	return nil
}

// FetchOrderBook fetches the current order book for the client's pair.
func (cl *Client) FetchOrderBook() error {
	log.Printf("bitx/streamer/client.FetchOrderBook(): Fetching order book.")
	req := &streamerpb.GetOrderBookRequest{
		Pair: cl.pair,
	}
	ob, err := cl.rpcClient.GetOrderBook(context.Background(), req)
	if err != nil {
		return err
	}
	orderBook, err := makeOrderBook(ob)
	if err != nil {
		return err
	}
	cl.orderBook = orderBook
	log.Printf("bitx/streamer/client.FetchOrderBook(): Built order book with "+
		"%d order(s).", orderBook.Len())
	return nil
}

// Stream listens for trading updates from the server.
func (cl *Client) Stream() {
	defer func() {
		cl.Quit <- true
	}()
	stream, err := cl.rpcClient.StreamUpdates(context.Background(),
		&streamerpb.StreamUpdatesRequest{Pair: cl.pair})
	if err != nil {
		log.Printf("bitx/streamer/client.Stream(): %v", err)
		return
	}
	for {
		update, err := stream.Recv()
		if err == io.EOF {
			log.Printf("bitx/streamer/client.Stream(): EOF")
			return
		}
		if err != nil {
			log.Printf("bitx/streamer/client.Stream(): %v", err)
			return
		}
		log.Printf("bitx/streamer/client.Stream(): Received update "+
			"(sequence = %d).", update.Sequence)
		err = cl.orderBook.handleUpdate(update)
		if err == ErrOutOfSequence {
			if err := cl.FetchOrderBook(); err != nil {
				log.Printf("bitx/streamer/client.Stream(): Error refetching "+
					"order book: %v", err)
				return
			}
		}
		if err != nil {
			log.Printf("bitx/streamer/client.Stream(): %v", err)
			return
		}
	}
}
