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

    // Calculate the size of a single record in the index file
    recordSize := len(key) + binary.Size(int64(0))

    // Determine the total number of records in the index file
    indexFileInfo, err := db.indexFile.Stat()
    if err != nil {
        return err
    }
    totalRecords := indexFileInfo.Size() / int64(recordSize)

    // Perform binary search to find the position of the key to be deleted
    low, high := int64(0), totalRecords
    for low < high {
        mid := (low + high) / 2

        // Seek to the middle of the index file
        offset := mid * int64(recordSize)
        _, err := db.indexFile.Seek(offset, io.SeekStart)
        if err != nil {
            return err
        }

        // Read key from the index file
        indexKey := make([]byte, len(key))
        _, err = db.indexFile.Read(indexKey)
        if err != nil && err != io.EOF {
            return err
        }

        // Compare keys
        cmp := bytes.Compare(indexKey, key)
        if cmp == 0 {
            // Key found, delete the entry by shifting subsequent records

            // Calculate the size of the data record
            var offset int64
            if err := binary.Read(db.indexFile, binary.LittleEndian, &offset); err != nil {
                return err
            }
            dataRecordSize := int64(binary.Size(uint32(len(key))) + binary.Size(uint32(0)) + len(key) + binary.Size(int64(0)))

            // Shift subsequent records
            var shiftedData []byte
            if mid < totalRecords-1 {
                shiftedDataSize := (totalRecords - mid - 1) * recordSize
                shiftedData = make([]byte, shiftedDataSize)
                _, err = db.indexFile.Read(shiftedData)
                if err != nil && err != io.EOF {
                    return err
                }

                _, err = db.indexFile.Seek(-int64(len(shiftedData)), io.SeekCurrent)
                if err != nil {
                    return err
                }

                _, err = db.indexFile.Write(shiftedData)
                if err != nil {
                    return err
                }
            }

            // Truncate the index file to remove the deleted key
            err := db.indexFile.Truncate(indexFileInfo.Size() - recordSize)
            if err != nil {
                return err
            }

            // Update the nextOffset
            db.nextOffset -= dataRecordSize

            return nil
        } else if cmp < 0 {
            // Key may be in the upper half
            low = mid + 1
        } else {
            // Key may be in the lower half
            high = mid
        }
    }

    // Key not found
    return fmt.Errorf("key not found")
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

    // Calculate the size of a single record in the index file
    recordSize := len(key) + binary.Size(int64(0))

    // Determine the total number of records in the index file
    indexFileInfo, err := db.indexFile.Stat()
    if err != nil {
        return err
    }
    totalRecords := indexFileInfo.Size() / int64(recordSize)

    // Perform binary search to find the correct position for the new key
    low, high := int64(0), totalRecords
    for low < high {
        mid := (low + high) / 2

        // Seek to the middle of the index file
        offset := mid * int64(recordSize)
        _, err := db.indexFile.Seek(offset, io.SeekStart)
        if err != nil {
            return err
        }

        // Read key from the index file
        indexKey := make([]byte, len(key))
        _, err = db.indexFile.Read(indexKey)
        if err != nil && err != io.EOF {
            return err
        }

        // Compare keys
        cmp := bytes.Compare(indexKey, key)
        if cmp == 0 {
            // Key already exists, update the value

            // Read and discard the offset
            var existingOffset int64
            if err := binary.Read(db.indexFile, binary.LittleEndian, &existingOffset); err != nil {
                return err
            }

            // Update the value in the data file
            if err := db.writeDataRecord(db.dataFile, existingOffset, key, value); err != nil {
                return err
            }

            return nil
        } else if cmp < 0 {
            // Key may be in the upper half
            low = mid + 1
        } else {
            // Key may be in the lower half
            high = mid
        }
    }

    // At this point, 'low' indicates the position where the new key should be inserted

    // Seek to the correct position in the index file for insertion
    offset := low * int64(recordSize)
    _, err = db.indexFile.Seek(offset, io.SeekStart)
    if err != nil {
        return err
    }

    // Shift the subsequent records to make space for the new key
    var shiftedData []byte
    if low < totalRecords {
        shiftedData = make([]byte, (totalRecords-low)*recordSize)
        _, err = db.indexFile.Read(shiftedData)
        if err != nil && err != io.EOF {
            return err
        }

        _, err = db.indexFile.Seek(-int64(len(shiftedData)), io.SeekCurrent)
        if err != nil {
            return err
        }

        _, err = db.indexFile.Write(shiftedData)
        if err != nil {
            return err
        }
    }

    // Write the new key and offset to the index file
    if _, err := db.indexFile.Write(key); err != nil {
        return err
    }
    if err := binary.Write(db.indexFile, binary.LittleEndian, db.nextOffset); err != nil {
        return err
    }

    // Update the nextOffset
    db.nextOffset += int64(binary.Size(uint32(len(key))) + binary.Size(uint32(len(value))) + len(key) + len(value) + binary.Size(int64(0)))

    // Write key-value pair to the data file
    if err := db.writeDataRecord(db.dataFile, db.nextOffset-int64(len(key))-int64(len(value))-binary.Size(int64(0)), key, value); err != nil {
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

    // Determine the size of a single record in the index file
    recordSize := len(key) + binary.Size(int64(0))

    // Determine the total number of records in the index file
    indexFileInfo, err := db.indexFile.Stat()
    if err != nil {
        return nil, err
    }
    totalRecords := indexFileInfo.Size() / int64(recordSize)

    // Perform binary search
    low, high := int64(0), totalRecords-1
    for low <= high {
        mid := (low + high) / 2

        // Seek to the middle of the index file
        offset := mid * int64(recordSize)
        _, err := db.indexFile.Seek(offset, io.SeekStart)
        if err != nil {
            return nil, err
        }

        // Read key from the index file
        indexKey := make([]byte, len(key))
        _, err = db.indexFile.Read(indexKey)
        if err != nil {
            return nil, err
        }

        // Compare keys
        cmp := bytes.Compare(indexKey, key)
        if cmp == 0 {
            // Key found, read offset and retrieve value from data file
            var offset int64
            if err := binary.Read(db.indexFile, binary.LittleEndian, &offset); err != nil {
                return nil, err
            }
            _, err := db.dataFile.Seek(offset, io.SeekStart)
            if err != nil {
                return nil, err
            }

            // Read value from data file
            var valueLength uint32
            if err := binary.Read(db.dataFile, binary.LittleEndian, &valueLength); err != nil {
                return nil, err
            }
            valueData := make([]byte, valueLength)
            if _, err := db.dataFile.Read(valueData); err != nil {
                return nil, err
            }

            return valueData, nil
        } else if cmp < 0 {
            // Key may be in the upper half
            low = mid + 1
        } else {
            // Key may be in the lower half
            high = mid - 1
        }
    }

    // Key not found
    return nil, fmt.Errorf("key not found")
}
