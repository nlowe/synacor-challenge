package cmd

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"time"

	"github.com/nsf/termbox-go"

	"github.com/nlowe/synacor-challenge/synacor"
	"github.com/spf13/cobra"
)

func NewRootCmd() *cobra.Command {
	var debugLog string
	challengeFile := "challenge.bin"
	watchdog := 5 * time.Minute

	result := &cobra.Command{
		Use:   "synacor",
		Short: "Go implementation of the synacor challenge VM spec",
		RunE: func(_ *cobra.Command, _ []string) error {
			f, err := os.Open(challengeFile)
			if err != nil {
				return err
			}

			defer func() {
				_ = f.Close()
			}()

			ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
			defer cancel()

			in := make(chan rune)

			c, out, err := synacor.NewCPUFrom(ctx, f, in)
			if err != nil {
				return err
			}

			if debugLog != "" {
				if _, err := os.Stat(debugLog); os.IsExist(err) {
					if err := os.Remove(debugLog); err != nil {
						return err
					}
				}

				debug, err := os.Open(debugLog)
				if err != nil {
					return err
				}

				defer func() {
					_ = debug.Close()
				}()

				c.Debug = debug
			}

			go func() {
				for r := range out {
					fmt.Print(string(r))
				}
			}()

			go func() {
				<-ctx.Done()
				termbox.Interrupt()
			}()

			go func() {
				for {
					switch ev := termbox.PollEvent(); ev.Type {
					case termbox.EventKey:
						if ev.Key == termbox.KeyCtrlC {
							cancel()
							return
						}

						if ev.Key == termbox.KeyEnter {
							fmt.Println()
							in <- '\n'
						}

						if ev.Key == termbox.KeySpace {
							fmt.Print(" ")
							in <- ' '
						}

						if ev.Ch != 0 {
							fmt.Print(string(ev.Ch))
							in <- ev.Ch
						}
					case termbox.EventError:
						panic(ev.Err)
					case termbox.EventInterrupt:
						return
					}
				}
			}()

			if err := termbox.Init(); err != nil {
				return err
			}
			defer termbox.Close()

			c.WatchdogTimeout = watchdog
			c.Run()

			return nil
		},
	}

	flags := result.PersistentFlags()

	flags.StringVarP(&challengeFile, "challenge-file", "c", challengeFile, "Challenge File to execute")
	flags.StringVarP(&debugLog, "debug-log", "d", debugLog, "Record instructions to debug log")
	flags.DurationVar(&watchdog, "io-watchdog", watchdog, "Timeout for I/O operations")

	return result
}
