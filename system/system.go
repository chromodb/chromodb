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
	"chromodb/datastructure"
	"net"
	"sync"
)

// Database is the ChromoDB main struct
type Database struct {
	FractalTree        *datastructure.FractalTree // Database tree
	CurrentMemoryUsage int                        // Current memory usage in bytes
	TCPListener        net.Listener               // TCPListener
	Wg                 *sync.WaitGroup            // System waitgroup
	Config             Config                     // ChromoDB configurations
}

// Config is the ChromoDB configurations struct
type Config struct {
	MemoryLimit int    // default is 750mb
	Port        int    // Port for listener, default is 7676
	TLS         bool   // Whether listener should listen on TLS or not
	TLSKey      string // If TLS is set where is the TLS key located?
	TLSCert     string // if TLS is set where is TLS cert located?
}
