package livechess

import (
	"context"
	"fmt"
	"github.com/seekerror/stdlib/pkg/util/slicex"
)

// AutoDetect returns the serial number of the likely eboard to use. It is likely if it's the only
// one altogether or the only one ACTIVE of multiple.
func AutoDetect(ctx context.Context, client Client) (EBoardSerial, error) {
	boards, err := client.EBoards(ctx)
	if err != nil {
		return "", err
	}

	switch len(boards) {
	case 0:
		return "", fmt.Errorf("no eboards")
	case 1:
		return boards[0].SerialNr, nil
	default:
		candidates := slicex.MapIf(boards, func(b EBoardResponse) (EBoardSerial, bool) {
			return b.SerialNr, b.State == Active
		})

		switch len(candidates) {
		case 0:
			return "", fmt.Errorf("multiple eboards, but none active")
		case 1:
			return candidates[0], nil
		default:
			return "", fmt.Errorf("multiple active eboards: %v", candidates)
		}
	}
}
