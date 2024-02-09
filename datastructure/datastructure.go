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
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"os"
)

// DataStructure represents the ChromoDB database structure
type DataStructure struct {
	dataFile   *os.File
	indexFile  *os.File
	nextOffset int64
}

// Delete takes a provided key and deletes the entry
func (db *DataStructure) Delete(key []byte) error {
	// Reset the offset of the index file to the beginning
	_, err := db.indexFile.Seek(0, io.SeekStart)
	if err != nil {
		return err
	}

	// Initialize a buffer to store the updated index data
	var updatedIndexBuffer bytes.Buffer

	for {
		// Read key from the index file
		indexKey := make([]byte, len(key))
		_, err := db.indexFile.Read(indexKey)
		if err == io.EOF {
			break
		} else if err != nil {
			return err
		}

		// Compare keys
		if bytes.Equal(indexKey, key) {

			// Read and discard the offset
			var offset int64
			if err := binary.Read(db.indexFile, binary.LittleEndian, &offset); err != nil {
				return err
			}

			// Seek to the corresponding offset in the data file
			_, err := db.dataFile.Seek(offset, io.SeekStart)
			if err != nil {
				return err
			}

			// Read the key length
			var keyLength uint32
			if err := binary.Read(db.dataFile, binary.LittleEndian, &keyLength); err != nil {
				return err
			}

			// Read the value length
			var valueLength uint32
			if err := binary.Read(db.dataFile, binary.LittleEndian, &valueLength); err != nil {
				return err
			}

			// Calculate the size of the data record
			dataRecordSize := int64(binary.Size(keyLength)) + int64(binary.Size(valueLength)) + int64(keyLength) + int64(valueLength) + int64(binary.Size(int64(0)))

			// Skip the data record in the data file
			if _, err := db.dataFile.Seek(dataRecordSize, io.SeekCurrent); err != nil {
				return err
			}

		} else {
			// Write the key to the updated index buffer
			if _, err := updatedIndexBuffer.Write(indexKey); err != nil {
				return err
			}

			// Read and write the offset to the updated index buffer
			var offset int64
			if err := binary.Read(db.indexFile, binary.LittleEndian, &offset); err != nil {
				return err
			}
			if err := binary.Write(&updatedIndexBuffer, binary.LittleEndian, offset); err != nil {
				return err
			}
		}
	}

	// Truncate the index file to remove the deleted key
	if err := db.indexFile.Truncate(0); err != nil {
		return err
	}

	// Write the updated index data to the index file
	if _, err := db.indexFile.Write(updatedIndexBuffer.Bytes()); err != nil {
		return err
	}

	return nil
}

// OpenDB opens or creates a DataStructure bassed DB
func OpenDB(dataFilename, indexFilename string) (*DataStructure, error) {
	dataFile, err := os.OpenFile(dataFilename, os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		return nil, err
	}

	indexFile, err := os.OpenFile(indexFilename, os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		return nil, err
	}

	// Calculate the next offset
	dataFileInfo, err := dataFile.Stat()
	if err != nil {
		return nil, err
	}
	nextOffset := dataFileInfo.Size()

	return &DataStructure{
		dataFile:   dataFile,
		indexFile:  indexFile,
		nextOffset: nextOffset,
	}, nil
}

// Close closes the DB
func (db *DataStructure) Close() error {
	if err := db.dataFile.Close(); err != nil {
		return err
	}
	return db.indexFile.Close()
}

// Put is like insert & update.  Will create a key-value but will replace an existing
// if key already exists
func (db *DataStructure) Put(key, value []byte) error {
	// Reset the offset of the index file to the beginning
	_, err := db.indexFile.Seek(0, io.SeekStart)
	if err != nil {
		return err
	}

	// Check if the key already exists
	for {
		// Read key from the index file
		indexKey := make([]byte, len(key))
		if _, err := db.indexFile.Read(indexKey); err == io.EOF {
			break
		} else if err != nil {
			return err
		}

		// Compare keys
		if bytes.Equal(indexKey, key) {
			// Key already exists, update the value

			// Read offset from the index file
			var offset int64
			if err := binary.Read(db.indexFile, binary.LittleEndian, &offset); err != nil {
				return err
			}

			// Seek to the corresponding offset in the data file
			_, err := db.dataFile.Seek(offset, io.SeekStart)
			if err != nil {
				return err
			}

			// Update the value in the data file
			if err := db.writeDataRecord(db.dataFile, offset, key, value); err != nil {
				return err
			}

			return nil
		}

		// Read and discard the offset (we don't need it for lookup)
		if _, err := db.indexFile.Seek(int64(binary.Size(int64(0))), io.SeekCurrent); err != nil {
			return err
		}
	}

	// Key does not exist, proceed with adding the new key-value pair

	// Write key to the index file
	if _, err := db.indexFile.Write(key); err != nil {
		return err
	}

	// Get the current offset in the data file
	offset, err := db.dataFile.Seek(0, io.SeekEnd)
	if err != nil {
		return err
	}

	// Write offset to the index file
	if err := binary.Write(db.indexFile, binary.LittleEndian, offset); err != nil {
		return err
	}

	// Write key-value pair to the data file
	// (using the offset obtained before writing to the index file)
	if err := db.writeDataRecord(db.dataFile, offset, key, value); err != nil {
		return err
	}

	return nil
}

// writeDataRecord writes a key-value record to the specified data file at the specified offset
func (db *DataStructure) writeDataRecord(dataFile io.Writer, offset int64, key, value []byte) error {
	// Write key length
	if err := binary.Write(dataFile, binary.LittleEndian, uint32(len(key))); err != nil {
		return err
	}

	// Write value length
	if err := binary.Write(dataFile, binary.LittleEndian, uint32(len(value))); err != nil {
		return err
	}

	// Write key
	if _, err := dataFile.Write(key); err != nil {
		return err
	}

	// Write value
	if _, err := dataFile.Write(value); err != nil {
		return err
	}

	// Write offset of the next record in the data file
	if err := binary.Write(dataFile, binary.LittleEndian, offset); err != nil {
		return err
	}

	return nil
}

// Get retrieves the value associated with a key
func (db *DataStructure) Get(key []byte) ([]byte, error) {
	// Reset the offset of the index file to the beginning
	_, err := db.indexFile.Seek(0, io.SeekStart)
	if err != nil {
		return nil, err
	}

	for {
		// Read key from the index file
		indexKey := make([]byte, len(key))
		if _, err := db.indexFile.Read(indexKey); err == io.EOF {
			break
		} else if err != nil {
			return nil, err
		}

		// Compare keys
		if bytes.Equal(indexKey, key) {
			// Read offset from the index file
			var offset int64
			if err := binary.Read(db.indexFile, binary.LittleEndian, &offset); err != nil {
				return nil, err
			}

			// Seek to the corresponding offset in the data file
			_, err := db.dataFile.Seek(offset, io.SeekStart)
			if err != nil {
				return nil, err
			}

			// Read the key length
			var keyLength uint32
			if err := binary.Read(db.dataFile, binary.LittleEndian, &keyLength); err != nil {
				return nil, err
			}

			// Read the value length
			var valueLength uint32
			if err := binary.Read(db.dataFile, binary.LittleEndian, &valueLength); err != nil {
				return nil, err
			}

			// Read the key
			keyData := make([]byte, keyLength)
			if _, err := db.dataFile.Read(keyData); err != nil {
				return nil, err
			}

			// Read the value
			valueData := make([]byte, valueLength)
			if _, err := db.dataFile.Read(valueData); err != nil {
				return nil, err
			}

			return valueData, nil
		}

		// Read and discard the offset (we don't need it for lookup)
		if _, err := db.indexFile.Seek(int64(binary.Size(int64(0))), io.SeekCurrent); err != nil {
			return nil, err
		}
	}

	// Key not found
	return nil, fmt.Errorf("key not found")
}
