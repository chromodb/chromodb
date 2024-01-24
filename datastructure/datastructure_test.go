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
package datastructure

import (
	"os"
	"testing"
)

func TestFractalTree_PutAndGet(t *testing.T) {
	// Create a temporary directory for test files
	tempDir := os.TempDir()

	// Initialize FractalTree
	db, err := OpenFractalTree(tempDir+"/chromo.db", tempDir+"/chromo.idx")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	// Test Put and Get operations
	key := []byte("test_key")
	value := []byte("test_value")

	// Put the key-value pair
	err = db.Put(key, value)
	if err != nil {
		t.Fatalf("Error putting key-value pair: %v", err)
	}

	// Get the value using the key
	result, err := db.Get(key)
	if err != nil {
		t.Fatalf("Error getting value for key: %v", err)
	}

	// Verify the retrieved value matches the expected value
	if string(result) != string(value) {
		t.Errorf("Expected value %s, got %s", string(value), string(result))
	}
}

func TestFractalTree_PutAndDeleteAndGet(t *testing.T) {
	// Create a temporary directory for test files
	tempDir := os.TempDir()

	// Initialize FractalTree
	db, err := OpenFractalTree(tempDir+"/chromo.db", tempDir+"/chromo.idx")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	// Test Put, Delete, and Get operations
	key := []byte("test_key")
	value := []byte("test_value")

	// Put the key-value pair
	err = db.Put(key, value)
	if err != nil {
		t.Fatalf("Error putting key-value pair: %v", err)
	}

	// Delete the key
	err = db.Delete(key)
	if err != nil {
		t.Fatalf("Error deleting key: %v", err)
	}

	// Attempt to Get the value using the deleted key
	result, err := db.Get(key)
	if err == nil {
		t.Error("Expected error for Get after deletion, but got nil")
	}

	// Verify that the result is nil after deletion
	if result != nil {
		t.Errorf("Expected result to be nil after deletion, got %s", string(result))
	}
}
