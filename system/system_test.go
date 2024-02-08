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
	"bytes"
	"chromodb/datastructure"
	"context"
	"encoding/base64"
	"io"
	"net"
	"os"
	"testing"
	"time"
)

func TestDatabase_ExecuteCommand(t *testing.T) {
	// Create a temporary directory for test files
	tempDir := os.TempDir()

	// Initialize a DS for the Database
	db, err := datastructure.OpenDB(tempDir+"/chromo.db", tempDir+"/chromo.idx")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	// Create a Database instance
	database := &Database{
		DataStructure: db,
		Config: Config{
			MemoryLimit: 750,
			Port:        7676,
			TLS:         false,
		},
		DBUser: DBUser{
			Username: "testuser",
			Password: "testpassword",
		},
	}

	// Put a key-value pair in the database
	err = database.DataStructure.Put([]byte("test_key"), []byte("test_value"))
	if err != nil {
		t.Fatalf("Error putting key-value pair: %v", err)
	}

	// Execute a GET command
	result, err := database.ExecuteCommand([]byte("GET->test_key"))
	if err != nil {
		t.Fatalf("Error executing GET command: %v", err)
	}

	// Verify the result matches the expected value
	expected := []byte("test_value")
	if !bytes.Equal(result.([]byte), expected) {
		t.Errorf("Expected result %s, got %s", string(expected), string(result.([]byte)))
	}
}

func TestDatabase_StartTCPTLSListener(t *testing.T) {
	// Create a temporary directory for test files
	tempDir := os.TempDir()

	// Initialize a DS for the Database
	db, err := datastructure.OpenDB(tempDir+"/chromo.db", tempDir+"/chromo.idx")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	// Create a Database instance
	database := &Database{
		DataStructure: db,
		Config: Config{
			MemoryLimit: 750,
			Port:        7676,
			TLS:         false,
		},
		DBUser: DBUser{
			Username: "testuser",
			Password: "testpassword",
		},
	}

	// Create a context with a timeout
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	// Start the TCP listener in a separate goroutine
	go func() {
		err := database.StartTCPTLSListener(ctx)
		if err != nil && err != context.Canceled {
			t.Errorf("Error starting TCP/TLS listener: %v", err)
		}
	}()

	// Wait for a short time to allow the listener to start
	time.Sleep(500 * time.Millisecond)

	// Connect to the TCP listener
	conn, err := net.Dial("tcp", "localhost:7676")
	if err != nil {
		t.Fatalf("Error connecting to TCP listener: %v", err)
	}
	defer conn.Close()

	// Send authentication credentials
	authString := base64.StdEncoding.EncodeToString([]byte("testuser\\0testpassword")) + "\r\n"
	conn.Write([]byte(authString))

	// Send a GET command
	conn.Write([]byte("GET->test_key\r\n"))

	// Read the response
	response, err := io.ReadAll(conn)
	if err != nil {
		t.Fatalf("Error reading response from TCP listener: %v", err)
	}

	// Verify the response contains the expected result
	expected := "key not found"
	if !bytes.Contains(response, []byte(expected)) {
		t.Errorf("Expected response to contain %s, got %s", expected, string(response))
	}

	// Stop the TCP listener
	database.Stop()
}
