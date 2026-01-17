package main

import (
	"errors"
	"fmt"
	"log/slog"
	"os"

	"immich-photo-frame/internal/app"
)

func main() {
	programLevel := new(slog.LevelVar)
	logger := slog.New(slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{Level: programLevel}))
	slog.SetDefault(logger)

	if ok, err := debugArgSet(os.Args[1:]); err != nil {
		slog.Error("argparse failed", "error", err)
		os.Exit(1)
	} else if ok {
		programLevel.Set(slog.LevelDebug)
	}

	if err := app.Run(); err != nil {
		slog.Error("app failed", "error", err)
		os.Exit(1)
	}
}

// debugArgSet checks the list of program arguments (without the program name)
// for the --debug flag. If an unexpected number of args is passed, or the flag
// is not "--debug", an error is returned.
func debugArgSet(args []string) (bool, error) {
	if len(args) == 0 {
		return false, nil
	} else if len(args) != 1 {
		return false, errors.New("too many arguments")
	}

	if args[0] != "--debug" {
		return false, fmt.Errorf(`unrecognized arg %q, expected "--debug"`, args[0])
	}
	return true, nil
}
