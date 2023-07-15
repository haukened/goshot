package main

import (
	"context"
	"errors"
	"fmt"
	"image/png"
	"os"
	"os/signal"
	"strings"
	"time"

	"github.com/kbinani/screenshot"
	"github.com/urfave/cli/v2"
)

func main() {
	app := &cli.App{
		Name:     "goshot",
		Usage:    "A lightweight screenshot tool, written in Go",
		Version:  "v0.1",
		Compiled: time.Now(),
		Authors: []*cli.Author{
			{
				Name:  "David Haukeness",
				Email: "david@hauken.us",
			},
		},
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "config",
				Aliases: []string{"c"},
				Usage:   "Load configuration from `FILE`",
				EnvVars: []string{"GOSHOT_CONFIG"},
			},
		},
		Action: run,
	}

	// create a means to terminate gracefully
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt)
	defer func() {
		signal.Stop(signalChan)
		cancel()
	}()

	// listen for hard and soft exits
	go func() {
		select {
		case <-signalChan: // first signal, cancel context
			cancel()
		case <-ctx.Done():
			return
		}
		<-signalChan // second signal, hard exit
		os.Exit(1)
	}()

	if err := app.RunContext(ctx, os.Args); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func run(c *cli.Context) error {
	// get the config
	configFile := c.String("config")
	conf, err := ReadConfig(configFile)
	if err != nil {
		return err
	}

	// make sure the path exists
	if _, err = os.Stat(conf.Path); errors.Is(err, os.ErrNotExist) {
		err = os.Mkdir(conf.Path, os.ModePerm)
		if err != nil {
			return err
		}
	}

	// take a screenshot for each display
	numDisplay := screenshot.NumActiveDisplays()
	t := getFormattedTime()
	for i := 0; i < numDisplay; i++ {
		select {
		case <-c.Context.Done():
			return c.Context.Err()
		default:
			bounds := screenshot.GetDisplayBounds(i)
			img, err := screenshot.CaptureRect(bounds)
			if err != nil {
				return err
			}
			filePath := fmt.Sprintf("%s/%s_%d.png", conf.Path, t, i+1)
			file, err := os.Create(filePath)
			if err != nil {
				return err
			}
			defer file.Close()
			png.Encode(file, img)
			fmt.Printf("Saved %s\n", filePath)
		}
	}
	return nil
}

func getFormattedTime() string {
	t := time.Now().UTC().Format(time.RFC3339)
	return strings.ReplaceAll(t, ":", "")
}
