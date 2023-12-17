// livechess is a command line tool for DGT LiveChess for developement.
package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/herohde/livechess-go/pkg/livechess"
	"github.com/seekerror/build"
	"github.com/seekerror/logw"
	"github.com/seekerror/stdlib/pkg/util/contextx"
	"github.com/seekerror/stdlib/pkg/util/signalx"
	"os"
)

var version = build.NewVersion(0, 1, 0)

var (
	serial = flag.String("serial", "auto", "Board selection by serial number (default: auto)")
	start  = flag.String("start", "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1", "Board start FEN position for setup or \"none\" (default: initial)")
)

func init() {
	flag.Usage = func() {
		fmt.Fprint(os.Stderr, fmt.Sprintf(`usage: livechess <command> [options]

DGT LiveChess command line tool, v%v.
commands:
 - eboards: list eboards
 - watch:   watch selected eboard
options:
`, version))
		flag.PrintDefaults()
	}
}

func main() {
	ctx := context.Background()

	if len(os.Args) < 2 {
		flag.Usage()
		logw.Exitf(ctx, "No command specified")
	}
	flag.CommandLine.Parse(os.Args[2:])

	command := os.Args[1]
	switch command {
	case "eboards":
		boards, err := livechess.DefaultClient.EBoards(ctx)
		if err != nil {
			logw.Exitf(ctx, ": %v", command)
		}
		for _, board := range boards {
			buf, _ := json.Marshal(board)
			fmt.Println(string(buf))
		}

	case "watch":
		id := livechess.EBoardSerial(*serial)
		if id == "auto" {
			auto, err := livechess.AutoDetect(ctx, livechess.DefaultClient)
			if err != nil {
				logw.Exitf(ctx, "Watch failed to autodetect board: %v", err)
			}
			id = auto
		}

		quit := signalx.TrapInterrupt()

		wctx, cancel := contextx.WithQuitCancel(ctx, quit)
		defer cancel()

		client, events, err := livechess.NewFeed(wctx, id)
		if err != nil {
			logw.Exitf(ctx, "Watch %v failed: %v", id, err)
		}

		if *start != "none" {
			if err := client.Setup(ctx, *start); err != nil {
				logw.Exitf(ctx, "Failed to setup %v: %v", *start, err)
			}
		}

		for {
			select {
			case event, ok := <-events:
				if !ok {
					return
				}
				buf, _ := json.Marshal(event)
				fmt.Println(string(buf))

			case <-quit:
				return
			}
		}

	default:
		flag.Usage()
		logw.Exitf(ctx, "Invalid command: %v", command)
	}
}
