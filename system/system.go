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
package system

import (
	"bufio"
	"bytes"
	"chromodb/datastructure"
	"context"
	"crypto/tls"
	"encoding/base64"
	"errors"
	"fmt"
	"net"
	"net/textproto"
	"os"
	"runtime"
	"strings"
	"sync"
	"time"
)

// Database is the ChromoDB main struct
type Database struct {
	DataStructure      *datastructure.DataStructure // Database tree
	CurrentMemoryUsage int                          // Current memory usage in bytes
	TCPListener        net.Listener                 // TCPListener
	Wg                 *sync.WaitGroup              // System waitgroup
	Config             Config                       // ChromoDB configurations
	DBUser             DBUser                       // Database user
	Mu                 *sync.Mutex
	Connections        map[net.Addr]net.Conn
}

// DBUser is a database user
type DBUser struct {
	Username string // database user username
	Password string // database user password
}

// Config is the ChromoDB configurations struct
type Config struct {
	MemoryLimit int    // default is 750mb
	Port        int    // Port for listener, default is 7676
	TLS         bool   // Whether listener should listen on TLS or not
	TLSKey      string // If TLS is set where is the TLS key located?
	TLSCert     string // if TLS is set where is TLS cert located?
}

// MonitorMemory monitors memory usage for database
func (db *Database) MonitorMemory() {
	ticker := time.NewTicker(time.Second * 5) // Check memory usage every 5 seconds

	for range ticker.C {
		var m runtime.MemStats
		runtime.ReadMemStats(&m)

		db.CurrentMemoryUsage = int(m.Alloc)
		if m.Alloc > uint64(db.Config.MemoryLimit) {
			fmt.Println("Memory usage exceeds the limit. Exiting...")
			os.Exit(1)
		}
	}
}

// ExecuteCommand takes a query and executes it
func (db *Database) ExecuteCommand(query []byte) (interface{}, error) {
	res, err := db.QueryParser(query)
	if err != nil {
		return nil, err
	}

	return res, nil
}

// QueryParser parses incoming query
func (db *Database) QueryParser(query []byte) (interface{}, error) {
	switch {
	case bytes.HasPrefix(bytes.ToUpper(query), []byte("MEM")):
		return []byte(fmt.Sprintf("Current memory usage: %d bytes", db.CurrentMemoryUsage)), nil
	case bytes.HasPrefix(bytes.ToUpper(query), []byte("PUT")):
		db.StartTransaction()
		opSpl := bytes.Split(query, []byte("->"))

		if len(opSpl) != 3 {
			db.RollbackTransaction()
			return nil, errors.New("bad sequence")
		}

		err := db.DataStructure.Put(bytes.TrimSpace(opSpl[1]), bytes.TrimSpace(opSpl[2]))
		if err != nil {
			db.RollbackTransaction()
			return nil, err
		}

		db.CommitTransaction()
		return []byte("PUT SUCCESS"), nil
	case bytes.HasPrefix(bytes.ToUpper(query), []byte("GET")):
		db.StartTransaction()
		opSpl := bytes.Split(query, []byte("->"))

		if len(opSpl) < 2 {
			db.RollbackTransaction()
			return nil, errors.New("bad sequence")
		}

		res, err := db.DataStructure.Get(bytes.TrimSpace(opSpl[1]))
		if err != nil {
			db.RollbackTransaction()
			return nil, err
		}

		db.CommitTransaction()
		return res, nil

	case bytes.HasPrefix(bytes.ToUpper(query), []byte("DISK")):
		totalDiskSpace, err := getDiskSpace("chromo.db", "chromo.idx")
		if err != nil {
			return nil, err
		}

		return []byte(fmt.Sprintf("DISK USAGE: %d bytes", totalDiskSpace)), nil
	case bytes.HasPrefix(bytes.ToUpper(query), []byte("DEL")):
		db.StartTransaction()
		opSpl := bytes.Split(query, []byte("->"))

		if len(opSpl) < 2 {
			db.RollbackTransaction()
			return nil, errors.New("bad sequence")
		}

		db.DataStructure.Delete(bytes.TrimSpace(opSpl[1]))

		db.CommitTransaction()
		return []byte("DEL SUCCESS"), nil
	}

	return nil, errors.New("nonexistent command")
}

// StartTCPTLSListener starts TCP/TLS listener
func (db *Database) StartTCPTLSListener(ctx context.Context) error {
	db.Wg = &sync.WaitGroup{}
	addr := "0.0.0.0:7676" // is the default for ChromoDB
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}

	db.TCPListener = listener

	db.Connections = make(map[net.Addr]net.Conn)

	go func() {
		for {
			conn, err := listener.Accept()
			if err != nil {
				// Handle errors (e.g., when the server is shutting down)
				select {
				case <-ctx.Done():
					return
				default:
					fmt.Println("Error accepting connection:", err)
				}
				continue
			}

			// Handle the connection in a separate goroutine
			db.Wg.Add(1)
			go func(c net.Conn) {
				defer db.Wg.Done()
				defer c.Close()
				defer delete(db.Connections, conn.RemoteAddr())

				if db.Config.TLS {
					cert, err := tls.LoadX509KeyPair(db.Config.TLSCert, db.Config.TLSKey)
					if err != nil {
						fmt.Println("Error loading certificate and private key:", err)
						return
					}

					// Perform TLS handshake to upgrade the connection
					tlsConn := tls.Server(conn, &tls.Config{
						Certificates:       []tls.Certificate{cert},
						InsecureSkipVerify: false,
					})

					err = tlsConn.Handshake()
					if err != nil {
						fmt.Println("TLS handshake error:", err)
						return
					}
				}

				// HANDLE AUTH
				// We expect username\0password encoded in base64
				auth := textproto.NewConn(conn)
				defer auth.Close() // close textproto conn

				toDecode, err := auth.ReadLine()
				if err != nil {
					fmt.Println("Auth textproto setup failure:", err)
					return
				}

				decodeString, err := base64.StdEncoding.DecodeString(toDecode)
				if err != nil {
					conn.Write([]byte("Invalid authentication. Bye!\r\n"))
					return
				}

				authSpl := strings.Split(string(decodeString), "\\0")

				if db.DBUser.Username != authSpl[0] {
					conn.Write([]byte("Invalid authentication. Bye!\r\n"))
					return
				}

				if db.DBUser.Password != authSpl[1] {
					conn.Write([]byte("Invalid authentication. Bye!\r\n"))
					return
				}

				conn.Write([]byte("AUTH OK\r\n"))

				db.Connections[conn.RemoteAddr()] = conn

				reader := bufio.NewReader(conn)

				for {

					// Read a line (until CRLF)
					line, err := reader.ReadBytes('\n')
					if err != nil {
						return
					}

					// Check for trailing CRLF
					if len(line) >= 2 && line[len(line)-2] == '\r' && line[len(line)-1] == '\n' {
						// Trailing CRLF found
						res, err := db.QueryParser(line)
						if err != nil {
							conn.Write(append([]byte(err.Error()), []byte("\r\n")...))
						} else {
							conn.Write(append(res.([]byte), []byte("\r\n")...))
						}
					} else if line[len(line)-1] == '\n' {
						// Trailing LF found
						res, err := db.QueryParser(line)
						if err != nil {
							conn.Write(append([]byte(err.Error()), []byte("\r\n")...))
						} else {
							conn.Write(append(res.([]byte), []byte("\r\n")...))
						}
					} else {
						fmt.Println("Invalid format, missing trailing CRLF")
					}
				}

			}(conn)
		}
	}()

	fmt.Println("TCP/TLS listener is listening on", addr)

	// Wait for the shutdown signal
	select {
	case <-ctx.Done():
		return nil
	}
}

// getDiskSpace gets combined disk space of provided files
func getDiskSpace(filePaths ...string) (int64, error) {
	var totalDiskSpace int64

	for _, filePath := range filePaths {
		// Get file information
		fileInfo, err := os.Stat(filePath)
		if err != nil {
			return 0, err
		}

		// Add file size to total disk space
		totalDiskSpace += fileInfo.Size()
	}

	return totalDiskSpace, nil
}

// Stop stops the TCP server
func (db *Database) Stop() {
	fmt.Println("TCP/TLS listener is shutting down...")
	if db.TCPListener != nil {
		_ = db.TCPListener.Close()
	}
	// Wait for all active connections to finish
	db.Wg.Wait()

	for _, c := range db.Connections {
		c.Close()
		delete(db.Connections, c.RemoteAddr())
	}

	fmt.Println("TCP/TLS listener stopped")
}

// StartTransaction locks to start a transaction
func (db *Database) StartTransaction() {
	db.Mu.Lock()
}

// RollbackTransaction unlocks on error
func (db *Database) RollbackTransaction() {
	db.Mu.Unlock()
}

// CommitTransaction similar to unlock but for commit
func (db *Database) CommitTransaction() {
	db.Mu.Unlock()
}
