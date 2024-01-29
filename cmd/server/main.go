/*
Copyright 2024 Drewbernetes.

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
	"github.com/drewbernetes/baski/pkg/server/server"
	"github.com/drewbernetes/baski/pkg/util/flags"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"
)

type Options struct {
	ip   string
	port int32
	dev  bool
	flags.S3Flags
}

func (o *Options) AddFlags(cmd *cobra.Command) {
	cmd.Flags().StringVarP(&o.ip, "bind-address", "a", "0.0.0.0", "The ip to bind to")
	cmd.Flags().Int32VarP(&o.port, "bind-port", "p", 8080, "The port to bind to")
	cmd.Flags().BoolVarP(&o.dev, "dev", "d", false, "Set to true to allow all in cors world")
}

func start() *cobra.Command {
	o := &Options{}

	cmd := &cobra.Command{
		Use:   "",
		Short: "Runs the api server",
		Long: `Runs the api server to which a front end will connect to pull results.

An Openstack cloud file is required to be read and the BASKI_OS_CLOUD var should be passed for parsing said file.
There is support for pulling DogKat results too. It is expected that the endpoint, access key and secret key have permissions to read from the DogKat bucket.
The following environment variables are required to ensure as flags are not supported wrt s3 credentials.
This is because it's expected this will be run in containers/kubernetes and as such env vars are easier to pass in via secrets and the flags for setting up the server are good defaults.
  * BASKI_S3_ENDPOINT
  * BASKI_S3_ACCESSKEY
  * BASKI_S3_SECRETKEY
  * BASKI_S3_BUCKET
  * BASKI_ENABLE_DOGKAT
  * BASKI_DOGKAT_BUCKET
  * BASKI_OS_CLOUD

The server runs on 0.0.0.0:8080 by default and this can be overridden via the flags.
`,
		RunE: func(cmd *cobra.Command, args []string) error {
			viper.SetEnvPrefix("BASKI")
			viper.AutomaticEnv()

			enableDogKat := false

			enableDogKat, err := strconv.ParseBool(viper.Get("ENABLE_DOGKAT").(string))
			if err != nil {
				return err
			}

			s := &server.Server{
				Options: server.Options{
					ListenAddress: o.ip,
					ListenPort:    o.port,
					Endpoint:      viper.Get("S3_ENDPOINT").(string),
					AccessKey:     viper.Get("S3_ACCESSKEY").(string),
					SecretKey:     viper.Get("S3_SECRETKEY").(string),
					Bucket:        viper.Get("S3_BUCKET").(string),
					EnableDogKat:  enableDogKat,
					DogKatBucket:  viper.Get("DOGKAT_BUCKET").(string),
					CloudName:     viper.Get("OS_CLOUD").(string),
				},
			}

			serv, err := s.NewServer(o.dev)
			if err != nil {
				return err
			}

			stop := make(chan os.Signal, 1)

			signal.Notify(stop, syscall.SIGTERM)

			go func() {
				<-stop

				// Shutdown the server, Kubernetes gives us 30 seconds before a SIGKILL.
				ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
				defer cancel()

				if err = serv.Shutdown(ctx); err != nil {
					log.Fatalln(err, "server shutdown error")
				}
			}()

			if err = serv.ListenAndServe(); err != nil {
				if errors.Is(err, http.ErrServerClosed) {
					return err
				}

				log.Fatalln(err, "unexpected server error")

				return err
			}

			return nil
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
