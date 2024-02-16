package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/dharmab/skyeye/internal/application"
	srs "github.com/dharmab/skyeye/pkg/simpleradio/types"
	"github.com/ggerganov/whisper.cpp/bindings/go/pkg/whisper"
	"github.com/lithammer/shortuuid/v3"
)

func main() {

	ctx, cancel := context.WithCancel(context.Background())

	interuptChan := make(chan os.Signal, 1)
	signal.Notify(interuptChan, os.Interrupt, syscall.SIGTERM)
	go func() {
		s := <-interuptChan
		slog.Info("received shutdown signal", "signal", s)
		cancel()
		time.Sleep(time.Second)
		os.Exit(0)
	}()
	LogLevel := flag.String("log-level", "info", "logging level (debug=-4, info=0, warn=4, error=8)")
	DCSGRPCAddress := flag.String("dcs-grpc-server-address", "localhost:50051", "address of the DCS-gRPC server")
	GRPCConnectionTimeout := flag.Duration("grpc-connection-timeout", 2*time.Second, "gRPC connection timeout")
	SRSAddress := flag.String("srs-server-address", "localhost:5002", "address of the SRS server")
	SRSConnectionTimeout := flag.Duration("srs-connection-timeout", 10*time.Second, "")
	SRSClientName := flag.String("srs-client-name", "SkyEye", "SRS client name. Appears in the client list and in in-game transmissions")
	SRSExternalAWACSModePassword := flag.String("srs-eam-password", "", "SRS external AWACS mode password")
	SRSFrequency := flag.Float64("srs-frequency", 133.0, "AWACS frequency")
	SRSCoalition := flag.String("srs-coalition", "blue", "SRS Coalition (either blue or red)")
	WhisperModelPath := flag.String("whisper-model", "", "Path to whisper.cpp model")

	slog.Info("parsing CLI flags")
	flag.Parse()

	var level slog.Level
	switch strings.ToLower(*LogLevel) {
	case "error":
		level = slog.LevelError
	case "warn":
		level = slog.LevelWarn
	case "info":
		level = slog.LevelInfo
	case "debug":
		level = slog.LevelDebug
	default:
		level = slog.LevelInfo
	}

	slog.SetDefault(
		slog.New(
			slog.NewJSONHandler(
				os.Stderr,
				&slog.HandlerOptions{
					Level: level,
				},
			),
		),
	)

	var coalition srs.Coalition
	if strings.EqualFold(*SRSCoalition, "blue") {
		coalition = srs.CoalitionBlue
	} else if strings.EqualFold(*SRSCoalition, "red") {
		coalition = srs.CoalitionRed
	} else {
		exitOnErr(errors.New("srs-coalition must be either blue or red"))
	}

	slog.Info("loading whisper model", "path", *WhisperModelPath)
	whisperModel, err := whisper.New(*WhisperModelPath)
	if err != nil {
		exitOnErr(fmt.Errorf("failed to load whisper model: %w", err))
	}
	defer whisperModel.Close()

	slog.Info("generating client GUID")
	clientGUID := shortuuid.New()

	config := application.Configuration{
		DCSGRPCAddress:               *DCSGRPCAddress,
		GRPCConnectionTimeout:        *GRPCConnectionTimeout,
		SRSAddress:                   *SRSAddress,
		SRSConnectionTimeout:         *SRSConnectionTimeout,
		SRSClientName:                *SRSClientName,
		SRSClientGUID:                clientGUID,
		SRSExternalAWACSModePassword: *SRSExternalAWACSModePassword,
		SRSFrequency:                 *SRSFrequency,
		SRSCoalition:                 coalition,
		WhisperModel:                 whisperModel,
	}

	slog.Info("starting application")
	app, err := application.NewApplication(ctx, config)
	exitOnErr(err)
	err = app.Run(ctx)
	exitOnErr(err)
}

func exitOnErr(err error) {
	if err != nil {
		slog.With("error", err).Error("application exiting with error")
		os.Exit(1)
	}
}
