package livechess

import (
	"encoding/json"
	"sync/atomic"
)

var (
	counter atomic.Int32
)

type ID int

func NewID() ID {
	return ID(counter.Add(1))
}

type Call struct {
	Call  string          `json:"call"`
	ID    ID              `json:"id"`
	Param json.RawMessage `json:"param,omitempty"`
}

type MethodParam struct {
	Method string          `json:"method,omitempty"` // "setup", "unsubscribe
	ID     ID              `json:"id,omitempty"`
	Param  json.RawMessage `json:"param,omitempty"`
}

type FenParam struct {
	FEN string `json:"fen,omitempty"`
}

type FlipParam struct {
	Flip bool `json:"flip"`
}

/*
{
    "call": "subscribe",
    "id": 42,
    "param": {
        "feed": "eboardevent",
        "id": 7,
        "param": {
            "serialnr": "7425"
        }
    }
}
*/

func NewSubscribeCall(id, feed ID, serial EBoardSerial) Call {
	v, _ := json.Marshal(SerialParam{
		SerialNr: serial,
	})
	param, _ := json.Marshal(FeedParam{
		Feed:  "eboardevent",
		ID:    feed,
		Param: v,
	})
	return Call{
		Call:  "subscribe",
		ID:    id,
		Param: param,
	}
}

/*
{
    "id":2,
    "call":"call",
    "param":{
        "id":7,
        "method":"setup",
        "param":{
            "fen":"r1bqkbnr/pppp1ppp/2n5/4p3/4P3/5N2/PPPP1PPP/RNBQKB1R w QKqk - 2 3"
        }
    }

*/

func NewSetupCall(id, feed ID, fen string) Call {
	v, _ := json.Marshal(FenParam{
		FEN: fen,
	})
	param, _ := json.Marshal(MethodParam{
		Method: "setup",
		ID:     feed,
		Param:  v,
	})
	return Call{
		Call:  "call",
		ID:    id,
		Param: param,
	}
}

/*
{
    "id":4,
    "call":"call",
    "param":{
        "id":7,
        "method":"flip",
        "param":{
            "flip":true
        }
    }
}
*/

func NewFlipCall(id, feed ID, flipped bool) Call {
	v, _ := json.Marshal(FlipParam{
		Flip: flipped,
	})
	param, _ := json.Marshal(MethodParam{
		Method: "setup",
		ID:     feed,
		Param:  v,
	})
	return Call{
		Call:  "call",
		ID:    id,
		Param: param,
	}
}

type ResponseType string

const (
	CallResponse  ResponseType = "call"
	FeedResponse  ResponseType = "feed"
	ErrorResponse ResponseType = "error"
)

type Response struct {
	Response ResponseType    `json:"response"` // "call", "error", "feed"
	ID       ID              `json:"id"`       // id of call or feed
	Param    json.RawMessage `json:"param,omitempty"`
	Time     int             `json:"time"`
}

type FeedParam struct {
	Feed  string          `json:"feed,omitempty"` // "eboardevent"
	ID    ID              `json:"id,omitempty"`
	Param json.RawMessage `json:"param,omitempty"`
}

type SerialParam struct {
	SerialNr EBoardSerial `json:"serialnr,omitempty"`
}

type ErrorParam struct {
	Message string `json:"message,omitempty"`
}

/*
export interface EBoardEventResponse {
    serialnr: string;
    flipped?: boolean;
    board?: string | undefined;
    clock?: ClockResponse | null | undefined;
    start?: string;
    san?: string[];
    match?: boolean;
}
*/

type EBoardEventResponse struct {
	// serialnr: string with the serialnr of the board generating an event
	SerialNr EBoardSerial `json:"serialnr"`
	// flipped: indicator if the board is flipped or not
	Flipped bool `json:"flipped"`
	// board: current position on the board, may be omitted if there was no board change
	Board string `json:"board,omitempty"`
	// clock: value of the clock, may be omitted if there was no clock change
	Clock *ClockResponse `json:"clock,omitempty"`
	// start: FEN string with the position from where reconstruction of moves is done
	Start string `json:"start,omitempty"`
	// san: array of SAN values with the detected moves
	San []string `json:"san"`
	// match: boolean indicating if the current board exactly matches the reconstructed move
	Match bool `json:"match"`
}
