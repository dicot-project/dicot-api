/*
 * This file is part of the Dicot project
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
 *
 * Copyright 2017 Red Hat, Inc.
 *
 */

package main

import (
	"bytes"
	"encoding/base64"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"syscall"

	"github.com/spf13/pflag"
	"golang.org/x/crypto/ssh/terminal"

	"github.com/dicot-project/dicot-api/pkg/crypto"
)

func hashIt(pw string) {
	hash, err := crypto.HashPassword(pw)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable hash password: %s\n", err)
		os.Exit(1)
	}

	hash64 := base64.StdEncoding.EncodeToString([]byte(hash))
	fmt.Printf("%s\n", hash64)
	os.Exit(0)
}

func main() {
	pflag.CommandLine.AddGoFlagSet(flag.CommandLine)

	var pwfile string
	pflag.StringVar(&pwfile, "password-file", "", "Path to file containing password")

	pflag.Parse()

	if pwfile != "" {
		pw, err := ioutil.ReadFile(pwfile)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Unable to read password file '%s': %s\n", pwfile, err)
			os.Exit(1)
		}
		pwstr := strings.TrimSuffix(string(pw), "\n")

		hashIt(pwstr)
	} else {
		for {
			fmt.Print("Enter password: ")

			pw1, err := terminal.ReadPassword(int(syscall.Stdin))
			if err != nil {
				fmt.Fprintf(os.Stderr, "Unable read password: %s\n", err)
				os.Exit(1)
			}
			fmt.Println("")

			fmt.Print("Repeat password: ")
			pw2, err := terminal.ReadPassword(int(syscall.Stdin))
			if err != nil {
				fmt.Fprintf(os.Stderr, "Unable read password: %s\n", err)
				os.Exit(1)
			}
			fmt.Println("")

			if bytes.Compare(pw1, pw2) != 0 {
				fmt.Fprint(os.Stderr, "Passwords do not match\n")
				continue
			}

			hashIt(string(pw1))
		}
	}
}
