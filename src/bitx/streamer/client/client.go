package client

import (
	"errors"
	"io"
	"log"
	"time"

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
	queue     *Queue
}

// ErrInvalidAddress indicates the provided address is not a valid server
// host:port combination.
var ErrInvalidAddress = errors.New("invalid address")

// New returns a new client.
func New(pair string) *Client {
	cl := &Client{
		pair:  pair,
		queue: NewQueue(),
	}
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

	log.Printf("bitx/streamer/client.Connect: Connected to %s.", address)

	cl.rpcClient = streamerpb.NewStreamerClient(conn)

	return nil
}

// fetchOrderBook fetches the current order book for the client's pair.
// If it fails it retries indefinitely.
func (cl *Client) fetchOrderBook() {
	var retry = 0
	for {
		if retry > 0 {
			time.Sleep(time.Second)
		}
		log.Printf("bitx/streamer/client.fetchOrderBook: Fetching order book.")
		req := &streamerpb.GetOrderBookRequest{
			Pair: cl.pair,
		}
		ob, err := cl.rpcClient.GetOrderBook(context.Background(), req)
		if err != nil {
			log.Printf("bitx/streamer/client.fetchOrderBook: Error fetching "+
				"order book: %v", err)
			retry++
			continue
		}
		orderBook, err := makeOrderBook(ob)
		if err != nil {
			log.Printf("bitx/streamer/client.fetchOrderBook: Error making "+
				"order book: %v", err)
			retry++
			continue
		}
		cl.orderBook = orderBook
		log.Printf("bitx/streamer/client.fetchOrderBook: Built order book "+
			"with %d order(s).", orderBook.Len())
		break
	}
}

// StreamForever listens for trading updates from the server.
func (cl *Client) StreamForever() {
	cl.fetchOrderBook()
	go cl.processQueueForever()

	stream, err := cl.rpcClient.StreamUpdates(context.Background(),
		&streamerpb.StreamUpdatesRequest{Pair: cl.pair})
	if err != nil {
		log.Printf("bitx/streamer/client.Stream: %v", err)
		return
	}
	for {
		update, err := stream.Recv()
		if err == io.EOF {
			log.Printf("bitx/streamer/client.Stream: EOF")
			return
		}
		if err != nil {
			log.Printf("bitx/streamer/client.Stream: %v", err)
			return
		}
		cl.queue.Enqueue(update)
		log.Printf("bitx/streamer/client.Stream: Received update: "+
			"sequence = %d, queue = %d.", update.Sequence, cl.queue.Len())
	}
}

func (cl *Client) processQueueForever() {
	for {
		func() {
			obj := cl.queue.Dequeue()
			if obj == nil {
				return
			}

			u := obj.(*streamerpb.Update)
			err := cl.orderBook.handleUpdate(u)
			if err != nil {
				log.Printf("bitx/streamer/client.processQueueForever: %v",
					err)
				cl.fetchOrderBook()
			}

			log.Printf("bitx/streamer/client.processQueueForever: Processed "+
				"update: sequence = %d, queue = %d", u.Sequence, cl.queue.Len())
		}()
	}
}
