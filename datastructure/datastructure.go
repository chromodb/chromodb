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
)

// FractalTree represents a Fractal Tree-like data structure for the ChromoDB
type FractalTree struct {
	dataFile   *os.File
	indexFile  *os.File
	nextOffset int64
}
