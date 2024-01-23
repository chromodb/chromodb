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
	"bytes"
	"context"
	"encoding/base64"
	"io"
	"io/ioutil"
	"net"
	"os"
	"testing"
	"time"

	"chromodb/datastructure"
)

func TestMain_ShellMode(t *testing.T) {
	// Redirect standard output for testing
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	// Create a temporary directory for test files
	tempDir := os.TempDir()

	// Initialize a FractalTree for the Database
	db, err := datastructure.OpenFractalTree(tempDir+"/chromo.db", tempDir+"/chromo.idx")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	// Set up a context with a timeout
	_, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	// Run the main function in shell mode
	go func() {
		defer w.Close()
		main()
	}()

	// Send shell commands to the standard input
	w.Write([]byte("MEM\r\n"))
	w.Write([]byte("PUT->test_key->test_value\r\n"))
	w.Write([]byte("GET->test_key\r\n"))
	w.Write([]byte("DEL->test_key\r\n"))
	w.Write([]byte("exit\r\n"))

	// Capture the output
	var output bytes.Buffer
	_, _ = io.Copy(&output, r)

	// Restore standard output
	os.Stdout = oldStdout

	// Verify the output
	expected := "db>Current memory usage: 0 bytes\n" +
		"db>PUT SUCCESS\n" +
		"db>test_value\n" +
		"db>DEL SUCCESS\n" +
		"..\n" +
		"bye!\n"

	if output.String() != expected {
		t.Errorf("Expected output:\n%s\nGot:\n%s", expected, output.String())
	}
}

func TestMain_NetworkMode(t *testing.T) {
	// Redirect standard output for testing
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	// Create a temporary directory for test files
	tempDir, err := ioutil.TempDir("", "chromodb_test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempDir)

	// Initialize a FractalTree for the Database
	db, err := datastructure.OpenFractalTree(tempDir+"/chromo.db", tempDir+"/chromo.idx")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	// Set up a context with a timeout
	_, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	// Run the main function in network mode
	go func() {
		defer w.Close()
		os.Args = []string{"chromodb", "--shell=false", "--user=testuser", "--pass=testpassword", "--port=7676"}
		main()
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

	// Send network commands to the TCP listener
	conn.Write([]byte("MEM\r\n"))
	conn.Write([]byte("PUT->test_key->test_value\r\n"))
	conn.Write([]byte("GET->test_key\r\n"))
	conn.Write([]byte("DEL->test_key\r\n"))

	// Capture the output
	var output bytes.Buffer
	_, _ = io.Copy(&output, r)

	// Verify the output
	expected := "Current memory usage: 0 bytes\n" +
		"PUT SUCCESS\n" +
		"test_value\n" +
		"DEL SUCCESS\n"

	if output.String() != expected {
		t.Errorf("Expected output:\n%s\nGot:\n%s", expected, output.String())
	}

	// Send the exit command
	conn.Write([]byte("exit\r\n"))

	// Capture the final output
	var finalOutput bytes.Buffer
	_, _ = io.Copy(&finalOutput, r)

	// Verify the final output
	expectedFinal := "..\n" +
		"bye!\n"

	if finalOutput.String() != expectedFinal {
		t.Errorf("Expected final output:\n%s\nGot:\n%s", expectedFinal, finalOutput.String())
	}

	// Restore standard output
	os.Stdout = oldStdout
}
