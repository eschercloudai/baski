/*
Copyright 2023 EscherCloud.

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

package gitRepo

import (
	"context"
	"log"
	"os"
	"os/signal"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
)

// GitClone will clone a designated repo.
func GitClone(repo, cloneLocation string, reference plumbing.ReferenceName) (*git.Repository, error) {
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		<-stop
		log.Printf("\nSignal detected, canceling operation...\n")
		cancel()
	}()

	log.Printf("downloading code from repo: %s\n", repo)
	gitRepo, err := git.PlainCloneContext(ctx, cloneLocation, false, &git.CloneOptions{
		URL:           repo,
		ReferenceName: reference,
	})
	if err != nil {
		return nil, err
	}

	return gitRepo, nil
}
