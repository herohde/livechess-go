package livechess

// EBoardSerial represents the eboard serial number.
type EBoardSerial string

// EBoardState represents the eboard state.
type EBoardState string

const (
	// NotActivated represents null: board has not yet been activated, typically source will be null to
	NotActivated EBoardState = ""
	// Active represents ‘ACTIVE’: board is alive and well
	Active EBoardState = "ACTIVE"
	// Inactive represents ‘INACTIVE’: board is not active because of operator action
	Inactive EBoardState = "INACTIVE"
	// NotResponding represents ‘NOTRESPONDING’: board is not responding to requests
	NotResponding EBoardState = "NOTRESPONDING"
	// Delayed represents ‘DELAYED’: board is present but board information is behind the real board state.
	// This kind of state only occurs in wireless setups of Caissa
	Delayed EBoardState = "DELAYED"
)

/*
export interface EBoardResponse {
  serialnr: string;
  source: string | null;
  state: string | null;
  battery: string | null;
  comment: string | null;
  board: string | null;
  flipped: boolean;
  clock: ClockResponse | null;
}
*/

type EBoardResponse struct {
	// SerialNr is the serial number of the board. This is a unique key.
	SerialNr EBoardSerial `json:"serialnr"`
	// Source is the name of the source for this eboard. This corresponds to the device field in sources.
	// This value will be null if the board is not present.
	Source string `json:"source,omitempty"`
	// Status information
	State EBoardState `json:"state,omitempty"`
	// Battery is string with battery information if a battery is present. This will be a floating
	// point number followed by % (bluetooth boards) or V (when connected to a caissa module).
	Battery string `json:"battery,omitempty"`
	// Comments is a string or null with an operator comment on the board.
	Comment string `json:"comment,omitempty"`
	// Board is a string containing the board part of a FEN string with the pieces detected on the board.
	Board string `json:"board,omitempty"`
	// Flipped is true if the board is flipped
	Flipped bool `json:"flipped"`
	// Clock include clock information, when a clock is present.
	Clock *ClockResponse `json:"clock,omitempty"`
}

/*
export interface ClockResponse {
  white: number;
  black: number;
  run: boolean | null;
  time: number;
}
*/

type ClockResponse struct {
	// White represents white: number of seconds for white on the clock.
	White int `json:"white,omitempty"`
	// Black represents black: number of seconds for black on the clock.
	Black int `json:"black,omitempty"`
	// Run represents run: null if the clock is not running, otherwise true for running clock for white,  false when clock is running for black.
	Run *bool `json:"run,omitempty"`
	// Time represents time: timestamp in milliseconds when this information was retrieved from the clock. This allows for the calculation of the running clock value.
	Time int `json:"time,omitempty"`
}
