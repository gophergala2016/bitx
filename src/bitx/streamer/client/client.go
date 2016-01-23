package client

import (
	"errors"
	"io"
	"log"

	"golang.org/x/net/context"
	"google.golang.org/grpc"

	"bitx/streamer/streamerpb"
)

type Client struct {
	pair      string
	rpcClient streamerpb.StreamerClient
	orderBook *OrderBook
	Quit      chan bool
}

var ErrInvalidAddress = errors.New("invalid address")

// New returns a new client.
func New(pair string) *Client {
	cl := &Client{pair: pair, Quit: make(chan bool, 1)}
	return cl
}

// Connect connects to the gRPC server.
func (cl *Client) Connect(address string) error {
	log.Printf("bitx/streamer/client.Connect(): Connecting to %s", address)

	if address == "" {
		return ErrInvalidAddress
	}

	creds := make([]grpc.DialOption, 0)
	// TODO(neil): TLS?
	creds = append(creds, grpc.WithInsecure())

	conn, err := grpc.Dial(address, creds...)
	if err != nil {
		return err
	}

	cl.rpcClient = streamerpb.NewStreamerClient(conn)

	return nil
}

// FetchOrderBook fetches the current order book for the client's pair.
func (cl *Client) FetchOrderBook() error {
	log.Printf("bitx/streamer/client.FetchOrderBook(): Fetching...")
	req := &streamerpb.GetOrderBookRequest{
		Pair: cl.pair,
	}
	ob, err := cl.rpcClient.GetOrderBook(context.Background(), req)
	if err != nil {
		return err
	}
	log.Printf("bitx/streamer/client.FetchOrderBook(): Received: %#v", ob)
	orderBook, err := makeOrderBook(ob)
	if err != nil {
		return err
	}
	cl.orderBook = orderBook
	log.Printf("bitx/streamer/client.FetchOrderBook(): Built order book: %#v",
		orderBook)
	log.Printf("%s", cl.orderBook)
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
		log.Printf("bitx/streamer/client.Stream(): Received %#v", update)
		err = cl.orderBook.handleUpdate(update)
		if err == ErrOutOfSequence {
			if err := cl.FetchOrderBook(); err != nil {
				log.Printf("bitx/streamer/client.Stream(): Error refetching "+
					"order book: %v", err)
				return
			}
		}
		log.Printf("%s", cl.orderBook)
		if err != nil {
			log.Printf("bitx/streamer/client.Stream(): %v", err)
			return
		}
	}
}
