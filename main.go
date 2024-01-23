/*
* ChromoDB
* ******************************************************************
* Originally authored by Alex Gaetano Padula
* Copyright (C) ChromoDB
*
* This program is free software: you can redistribute it and/or modify
* it under the terms of the GNU General Public License as published by
* the Free Software Foundation, either version 3 of the License, or
* (at your option) any later version.
*
* This program is distributed in the hope that it will be useful,
* but WITHOUT ANY WARRANTY; without even the implied warranty of
* MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
* GNU General Public License for more details.
*
* You should have received a copy of the GNU General Public License
* along with this program.  If not, see <http://www.gnu.org/licenses/>.
 */
package main

import (
	"bufio"
	"bytes"
	"chromodb/datastructure"
	"chromodb/system"
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"
)

// System starts here
// ./chromodb
// ./chromodb --help
// ./chromodb --memory-limit=7500 * 1024 * 1024
// ./chromodb --shell=false --user=alex --pasword=somepassword
// ./chromodb --shell=false --user=alex --pasword=somepassword --tls=true --key="key.pem" --cert="cert.pem"
func main() {
	var db system.Database // Main system variable

	db.Config.MemoryLimit = 750 * 1024 * 1024 // 750MB

	go db.MonitorMemory() // This is for the MEM/mem command.  We check every 5 seconds

	// Load database and index file
	tree, err := datastructure.OpenFractalTree("chromo.db", "chromo.idx")
	if err != nil {
		return
	}

	db.FractalTree = tree // Set tree into system variable
	db.Config.Port = 7676 // Set default port

	var shell bool  // Use shell, good for embedded stuff
	var tls bool    // Upgrade clients to TLS
	var help bool   // Help show all flags
	var user string // Database user for remote connections to use
	var pass string // Database password for remote connections to use
	// if tls enabled
	var cert string // tls cert location
	var key string  // tls key location

	flag.BoolVar(&help, "help", help, "displays flag instructions")
	flag.BoolVar(&shell, "shell", shell, "true or false to use internal shell")
	flag.BoolVar(&tls, "tls", tls, "enable tls listener.  you must provide a cert and key using --cert and --key flags.")
	flag.IntVar(&db.Config.MemoryLimit, "memory-limit", db.Config.MemoryLimit, "configure desired memory limit.  default is 750mb i.e 750 * 1024 * 1024")
	flag.StringVar(&user, "user", user, "database user username for when using network")
	flag.StringVar(&pass, "pass", pass, "database user password for when using network")
	flag.StringVar(&cert, "key", user, "tls cert location")
	flag.StringVar(&key, "cert", pass, "tls key location")
	flag.IntVar(&db.Config.Port, "port", db.Config.Port, "tcp/tls listener port default is 7676")

	flag.Parse() // parse flags

	if help { // if help display flag usages
		flag.Usage()
		os.Exit(0)
	}

	if !shell { // if not shell we will start up a networked ChromoDB
		if user == "" && pass == "" {
			fmt.Println("Database username and password is required when configuring database to be networked.")
			os.Exit(1)
		}

		db.DBUser.Username = user
		db.DBUser.Password = pass
		// User is now set for listener

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		go func() {
			sigCh := make(chan os.Signal, 1)
			signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
			<-sigCh
			cancel()
		}()

		db.StartTCPTLSListener(ctx)

		db.Stop()

	} else { // Not networked

		qChan := make(chan []byte) // Query channel

		go func() {
			for {

				reader := bufio.NewReader(os.Stdin)
				fmt.Print("db>")
				query, _ := reader.ReadBytes('\n')

				qChan <- bytes.TrimSpace(query) // Receive from standard input and relay to query channel
			}
		}()

		for {
			query := <-qChan // Receive from query channel

			// Check for exit command
			if bytes.HasSuffix(query, []byte("exit")) {
				fmt.Println("..\nbye!")
				break
			} // maybe not needed or unnecessary

			// Execute the command
			output, err := db.ExecuteCommand(query)
			if err != nil {
				fmt.Println("Error executing query:", err)
				continue
			}

			// Print the output of the command
			fmt.Println(output)
		}
	}
}
