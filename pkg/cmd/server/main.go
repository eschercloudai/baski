/*
Copyright 2023 EscherCloudAI.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package main

import (
	"context"
	"errors"
	"github.com/eschercloudai/baski/pkg/cmd/util/flags"
	"github.com/eschercloudai/baski/pkg/server/server"
	"github.com/spf13/cobra"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

type Options struct {
	ip   string
	port int32
	dev  bool
	flags.S3Flags
	bucket string
}

func (o *Options) AddFlags(cmd *cobra.Command) {
	cmd.Flags().StringVarP(&o.ip, "bind-address", "a", "127.0.0.1", "The ip to bind to")
	cmd.Flags().Int32VarP(&o.port, "bind-port", "p", 8080, "The port to bind to")
	cmd.Flags().BoolVarP(&o.dev, "dev", "d", false, "Set to true to allow all in cors world")
	o.S3Flags.AddFlags(cmd)
	cmd.Flags().StringVar(&o.bucket, "bucket", "baski", "The S3 bucket")
}

func requireFlag(cmd *cobra.Command, name string) {

	err := cmd.MarkFlagRequired(name)
	if err != nil {
		log.Fatalln(err)
	}
}

func start() *cobra.Command {
	o := &Options{}

	cmd := &cobra.Command{
		Use:   "run",
		Short: "Runs the api server",
		Long:  "Runs the api server to which the front end will connect",
		Run: func(cmd *cobra.Command, args []string) {

			s := &server.Server{
				Options: server.Options{
					ListenAddress: o.ip,
					ListenPort:    o.port,
					Endpoint:      o.Endpoint,
					AccessKey:     o.AccessKey,
					SecretKey:     o.SecretKey,
					Bucket:        o.bucket,
				},
			}

			server, err := s.NewServer(o.dev)
			if err != nil {
				log.Fatalln(err)
			}

			stop := make(chan os.Signal, 1)

			signal.Notify(stop, syscall.SIGTERM)

			go func() {
				<-stop

				// Shutdown the server, Kubernetes gives us 30 seconds before a SIGKILL.
				ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
				defer cancel()

				if err := server.Shutdown(ctx); err != nil {
					log.Fatalln(err, "server shutdown error")
				}
			}()

			if err := server.ListenAndServe(); err != nil {
				if errors.Is(err, http.ErrServerClosed) {
					return
				}

				log.Fatalln(err, "unexpected server error")

				return
			}
		},
	}
	o.AddFlags(cmd)

	return cmd
}

// Execute runs the execute command for the Cobra library allowing commands & flags to be utilised.
func main() {
	if err := start().Execute(); err != nil {
		os.Exit(1)
	}
}
