/*
 * Copyright 2018 Yuji Ito <llamerada.jp@gmail.com>
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */
package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/fsnotify/fsnotify"
)

const (
	rootPath = "."
)

func enableDebugMode() error {
	log.Println("start debug mode")

	dfw, err := newDebugFileWatcher(rootPath)
	if err != nil {
		return err
	}

	go func() {
		for {
			err := dfw.wait()
			if err != nil {
				log.Fatal(err)
			}

			time.Sleep(time.Second * 1)

			err = build(rootPath)
			if err != nil {
				log.Print("failed to build", err)

			} else {
				os.Exit(0)
			}
		}
	}()

	return nil
}

func build(rootPath string) error {
	outs, err := execHelper(".", "make", []string{"build", "-C", rootPath})
	for _, out := range outs {
		fmt.Println(out)
	}
	return err
}

type debugFileWatcher struct {
	rootPath string
	watched  []string
	watcher  *fsnotify.Watcher
}

func newDebugFileWatcher(rootPath string) (*debugFileWatcher, error) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}

	return &debugFileWatcher{
		rootPath: rootPath,
		watched:  make([]string, 0),
		watcher:  watcher,
	}, nil
}

func (w *debugFileWatcher) wait() error {
	isFirst := true

	for {
		work1, err := execHelper(w.rootPath, "git", []string{"ls-files"})
		if err != nil {
			return err
		}
		work2, err := execHelper(w.rootPath, "git", []string{"ls-files", "--exclude-standard", "-o"})
		if err != nil {
			return err
		}

		updated := false
		for _, file := range append(work1, work2...) {
			ignored := false
			for _, watched := range w.watched {
				if file == watched {
					ignored = true
					break
				}
			}
			if !ignored {
				w.watched = append(w.watched, file)
				w.watcher.Add(filepath.Join(w.rootPath, file))
				updated = true
			}
		}

		if !isFirst && updated {
			break
		}
		isFirst = false

		event, ok := <-w.watcher.Events
		if ok && event.Op&fsnotify.Write != 0 {
			return nil
		}
	}

	return nil
}

func execHelper(path, command string, args []string) ([]string, error) {
	cmd := exec.Command(command, args...)
	cmd.Dir = path
	out, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("error on execHelper %v", map[string]interface{}{
			"path":    path,
			"command": command,
			"args":    args,
		})
	}
	return strings.Split(string(out), "\n"), nil
}
