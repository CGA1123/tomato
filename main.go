package main

import (
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/CGA1123/tomato/client"
	"github.com/CGA1123/tomato/pb"
	"github.com/CGA1123/tomato/server"
	"github.com/soellman/pidfile"
	"github.com/spf13/cobra"
	"google.golang.org/grpc"
)

var (
	Socket        = "/tmp/tomato.sock"
	LogFile       = "/tmp/tomato.log"
	PidFile       = "/tmp/tomato.pid"
	LogPrefix     = "🍅 "
	Quiet         = false
	ErrNotRunning = errors.New("not running")
)

func main() {
	cmd := Cmd()
	err := cmd.ExecuteContext(context.Background())
	if err != nil {
		fmt.Printf("Error: %v", err)
		if err == ErrNotRunning {
			os.Exit(33)
		}

		os.Exit(1)
	}
}

func Cmd() *cobra.Command {
	rootCmd := &cobra.Command{
		Use:           "tomato",
		Short:         "tomato is a tomato timer for your terminal!",
		SilenceUsage:  true,
		SilenceErrors: true,
	}

	rootCmd.AddCommand(
		serve(),
		kill(),
		up(),
		start(),
		stop(),
		running(),
		remaining(),
	)

	return rootCmd
}

func WithClient(f func(*client.Client) error) error {
	log.SetFlags(0)
	log.SetPrefix(LogPrefix)

	if !serverRunning() {
		return errors.New("tomato server is not running")
	}

	c, err := client.New(Socket)
	if err != nil {
		log.Printf("is the server running? start it with tomato server")
		return fmt.Errorf("error creating client: %w", err)
	}

	return f(c)
}

func serverRunning() bool {
	pid, err := pidfileContents(PidFile)
	if err != nil {
		return false
	}

	return pidIsRunning(pid)
}

func pidfileContents(filename string) (int, error) {
	contents, err := ioutil.ReadFile(filename)
	if err != nil {
		return 0, err
	}

	pid, err := strconv.Atoi(strings.TrimSpace(string(contents)))
	if err != nil {
		return 0, pidfile.ErrFileInvalid
	}

	return pid, nil
}

func pidIsRunning(pid int) bool {
	process, err := os.FindProcess(pid)
	if err != nil {
		return false
	}

	err = process.Signal(syscall.Signal(0))

	if err != nil && err.Error() == "no such process" {
		return false
	}

	if err != nil && err.Error() == "os: process already finished" {
		return false
	}

	return true
}

func remaining() *cobra.Command {
	return &cobra.Command{
		Use:   "remaining",
		Short: "Returns how long is left on the current tomato.",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return WithClient(func(c *client.Client) error {
				left, err := c.Remaining()
				if err != nil {
					return err
				}

				if left == time.Duration(0) {
					if Quiet {
						fmt.Println("0")
					} else {
						log.Printf("the clock was not running!")
						log.Printf("use `tomato start` to get going.")
					}
					return nil
				}

				minutes := left.Round(time.Minute).Minutes()

				if Quiet {
					fmt.Printf("%.0f\n", minutes)
				} else {
					log.Printf("there are %.0f minutes left on the clock!", minutes)
				}
				return nil
			})
		},
	}
}

func running() *cobra.Command {
	return &cobra.Command{
		Use:   "runnning",
		Short: "Checks whether there is a current tomato running.",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return WithClient(func(c *client.Client) error {
				running, err := c.Running()
				if err != nil {
					return err
				}

				if !running {
					return ErrNotRunning
				}

				return nil
			})
		},
	}
}

func stop() *cobra.Command {
	return &cobra.Command{
		Use:   "stop",
		Short: "Stop the currently running tomato.",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return WithClient(func(c *client.Client) error {
				left, err := c.Stop()
				if err != nil {
					return err
				}

				minutes := left.Round(time.Minute).Minutes()

				if Quiet {
					fmt.Printf("%.0f\n", minutes)
				} else {
					log.Printf("there was %.0f minute(s) left on the clock!", minutes)
				}

				return nil
			})
		},
	}
}

func start() *cobra.Command {
	return &cobra.Command{
		Use:   "start",
		Short: "Starts a tomato timer",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return WithClient(func(c *client.Client) error {
				finish, err := c.Start()
				if err != nil {
					return err
				}

				fmtd := finish.Format("15:04")
				if Quiet {
					fmt.Println(fmtd)
				} else {
					log.Printf("timer will finish at %v", fmtd)
				}

				return nil

			})
		},
	}
}

func serve() *cobra.Command {
	return &cobra.Command{
		Use:   "server",
		Short: "Starts the tomato server.",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := pidfile.WriteControl(PidFile, os.Getpid(), true); err != nil {
				return err
			}

			f, err := os.Create(LogFile)
			if err != nil {
				log.Printf("error creating logfile: %v", err)
			}

			defer func() {
				log.Printf("Shutting down...")
				log.Printf("Removing pidfile...")
				pidfile.Remove(PidFile)
				log.Printf("Removing socket...")
				os.RemoveAll(Socket)
				log.Printf("👋")
				f.Close()
			}()

			log.Printf("Will write logs to: %v", LogFile)
			log.SetOutput(f)
			log.SetPrefix(LogPrefix)
			log.Printf("Starting server at [%s]...", Socket)

			listener, err := net.Listen("unix", Socket)
			if err != nil {
				return fmt.Errorf("error opening socket: %w", err)
			}

			srv := grpc.NewServer()
			pb.RegisterTomatoServiceServer(srv, server.New())

			errorC := make(chan error, 1)
			shutdownC := make(chan os.Signal, 1)

			go func(errC chan<- error) {
				errorC <- srv.Serve(listener)
			}(errorC)

			signal.Notify(shutdownC, syscall.SIGINT, syscall.SIGTERM)

			select {
			case err := <-errorC:
				if err != nil && err != http.ErrServerClosed {
					return err
				}

				return err
			case <-shutdownC:
				return nil
			}
		},
	}
}

func up() *cobra.Command {
	return &cobra.Command{
		Use:   "up",
		Short: "Check whether the tomato server is up",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if !serverRunning() {
				return ErrNotRunning
			}

			return nil
		},
	}
}

func kill() *cobra.Command {
	return &cobra.Command{
		Use:   "kill",
		Short: "Kill the tomato server.",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if !serverRunning() {
				return errors.New("server is not running")
			}

			pid, err := pidfileContents(PidFile)
			if err != nil {
				return nil
			}

			process, err := os.FindProcess(pid)
			if err != nil {
				return nil
			}

			return process.Signal(syscall.SIGTERM)
		},
	}
}
