package livechess

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	ws "github.com/gorilla/websocket"
	"github.com/seekerror/stdlib/pkg/util/iox"
	"log"
)

/*
export interface EBoardEventFeed {
    subscribe(serialnr: string): void;
    setup(fen: string): void;
    flip(flip: boolean): void;
}
*/

type FeedClient interface {
	Setup(ctx context.Context, fen string) error
	Flip(ctx context.Context, flipped bool) error
}

func NewFeed(ctx context.Context, serial EBoardSerial) (FeedClient, <-chan EBoardEventResponse, error) {
	conn, _, err := ws.DefaultDialer.Dial(fmt.Sprintf("ws://%v/%v", endpoint, path), nil)
	if err != nil {
		return nil, nil, err
	}

	// Subscribe to feedid 7 using call 42 and subscribing to eboardevent for board 7425:
	//
	// {
	//    "call": "subscribe",
	//    "id": 42,
	//    "param": {
	//        "feed": "eboardevent",
	//        "id": 7,
	//        "param": {
	//            "serialnr": "7425"
	//        }
	//    }
	// }
	//
	// LiveChess will respond with a confirmation of the subscription call and an initial feed message:
	//
	//{
	//    "id": 42,
	//    "response": "call",
	//    "param": null,
	//    "time": 1503849460753
	//}
	//
	//{
	//    "response":"feed",
	//    "id":7,
	//    "param":{
	//        "serialnr":"7425",
	//        "flipped":false,
	//        "board":"r1bqkbnr/pppp1ppp/2n5/4p3/4P3/5N2/PPPP1PPP/RNBQKB1R",
	//        "clock":{
	//            "white":5676,
	//            "black":7190,
	//            "run":true,
	//            "time":1503845105357
	//        }
	//    },
	//    "time":1503849460756
	// }

	feedID := NewID()

	subscribe := NewSubscribeCall(NewID(), feedID, serial)
	if err := conn.WriteJSON(subscribe); err != nil {
		_ = conn.Close()
		return nil, nil, fmt.Errorf("failed to subscribe: %v", err)
	}
	if _, err := readResponse(conn, true); err != nil {
		_ = conn.Close()
		return nil, nil, fmt.Errorf("subscribe rejected: %v", err)
	}

	quit := iox.NewAsyncCloserWithCancel(ctx)

	out := make(chan EBoardEventResponse, 20)
	requests := make(chan request)
	responses := make(chan Response)

	// (1) Read responses. Short-circuit event forwarding.

	go func() {
		defer close(out)
		defer quit.Close()

		for !quit.IsClosed() {
			resp, err := readResponse(conn, true)
			if err != nil {
				if !quit.IsClosed() {
					log.Printf("Failed to read feed response: %v", err)
				}
				return
			}

			switch resp.Response {
			case FeedResponse:
				event, err := unmarshal[EBoardEventResponse](resp.Param)
				if err != nil {
					log.Printf("Failed to parse feed response %v: %v", resp, err)
					return
				}

				select {
				case out <- event:
					// ok
				case <-quit.Closed():
					return
				}

			case CallResponse:
				select {
				case responses <- resp:
					// ok
				case <-quit.Closed():
					return
				}

			default:
				log.Printf("Unexpected response type %v: %v", resp.Response, resp)
			}
		}
	}()

	// (2) Write and process requests, using a single writer to websocket connection.

	go func() {
		defer conn.Close()
		defer quit.Close()

		active := map[ID]request{}

		for {
			select {
			case req := <-requests:
				if err := conn.WriteJSON(req.req); err != nil {
					req.resp <- response{err: err}
					return
				}
				active[req.req.ID] = req

			case resp := <-responses:
				if live, ok := active[resp.ID]; ok {
					live.resp <- response{resp: resp}
					delete(active, resp.ID)
				}

			case <-quit.Closed():
				return
			}
		}
	}()

	client := newFeedClient(quit.Closed(), feedID, requests)

	return client, out, nil
}

type response struct {
	resp Response
	err  error
}

func (r response) Err() error {
	if r.resp.Response == ErrorResponse {
		r, _ := unmarshal[ErrorParam](r.resp.Param)
		return errors.New(r.Message)
	}
	return r.err
}

type request struct {
	req  Call
	resp chan<- response
}

type feedClient struct {
	quit <-chan struct{}
	feed ID
	ch   chan<- request
}

func newFeedClient(quit <-chan struct{}, feed ID, ch chan<- request) *feedClient {
	return &feedClient{quit: quit, feed: feed, ch: ch}
}

func (f *feedClient) Setup(ctx context.Context, fen string) error {
	return f.call(ctx, NewSetupCall(NewID(), f.feed, fen)).Err()
}

func (f *feedClient) Flip(ctx context.Context, flipped bool) error {
	return f.call(ctx, NewFlipCall(NewID(), f.feed, flipped)).Err()
}

func (f *feedClient) call(ctx context.Context, call Call) response {
	ch := make(chan response, 1)

	select {
	case f.ch <- request{req: call, resp: ch}:
		select {
		case resp := <-ch:
			return resp

		case <-f.quit:
			return response{err: fmt.Errorf("feed closed")}
		case <-ctx.Done():
			return response{err: ctx.Err()}
		}

	case <-f.quit:
		return response{err: fmt.Errorf("feed closed")}
	case <-ctx.Done():
		return response{err: ctx.Err()}
	}
}
func readResponse(conn *ws.Conn, convertErr bool) (Response, error) {
	_, buf, err := conn.ReadMessage()
	if err != nil {
		return Response{}, err
	}
	// log.Printf("<< %v", string(buf))

	resp, err := unmarshal[Response](buf)
	if err != nil {
		return Response{}, fmt.Errorf("invalid response: %v", err)
	}
	if convertErr && resp.Response == ErrorResponse {
		r, _ := unmarshal[ErrorParam](resp.Param)
		return Response{}, errors.New(r.Message)
	}
	return resp, nil
}

func unmarshal[T any](buf []byte) (T, error) {
	var ret T
	if err := json.Unmarshal(buf, &ret); err != nil {
		return ret, err
	}
	return ret, nil
}
