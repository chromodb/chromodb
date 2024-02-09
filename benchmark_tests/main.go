/*
* Some benchmark tests and consistency tests
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
	"fmt"
	"log"
	"net"
	"os"
	"sync"
	"time"
)

// Inserts 5000 keys using 5000 connections
func insertParallel() {
	wg := &sync.WaitGroup{}
	start := time.Now()

	for i := 0; i < 5000; i++ {
		wg.Add(1)
		go func(j int) {
			defer wg.Done()
			// Resolve the string address to a TCP address
			tcpAddr, err := net.ResolveTCPAddr("tcp", "localhost:7676")
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}

			// Connect to the address with tcp
			conn, err := net.DialTCP("tcp", nil, tcpAddr)
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}

			// Send a message to the ChromoDB running instance
			_, err = conn.Write([]byte("YWxleFwwc29tZXBhc3N3b3Jk\n")) // we are using a user of alex and password of somepassword
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}

			_, err = bufio.NewReader(conn).ReadString('\n')
			if err != nil {
				fmt.Println(err)
				return
			}

			_, err = conn.Write([]byte(fmt.Sprintf("PUT->key%d->value%d\n", j, j))) // we are using a user of alex and password of somepassword
			if err != nil {
				fmt.Println(err)
				return
			}

			// Read from the connection untill a new line is send
			_, err = bufio.NewReader(conn).ReadString('\n')
			if err != nil {
				fmt.Println(err)
				return
			}

			conn.Close()
		}(i)
	}

	wg.Wait()

	elapsed := time.Since(start)
	fmt.Println("Fin")

	log.Printf("ChromoDB took to insert 5000 keys with 5000 connections: %s", elapsed)
}

// Insert 5000 keys using 100 connections
func insertParallel2() {
	wg := &sync.WaitGroup{}
	start := time.Now()

	e := 0 // current entry
	eMu := &sync.Mutex{}

	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(j int) {
			defer wg.Done()
			// Resolve the string address to a TCP address
			tcpAddr, err := net.ResolveTCPAddr("tcp", "localhost:7676")
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}

			// Connect to the address with tcp
			conn, err := net.DialTCP("tcp", nil, tcpAddr)
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}

			// Send a message to the ChromoDB running instance
			_, err = conn.Write([]byte("YWxleFwwc29tZXBhc3N3b3Jk\n")) // we are using a user of alex and password of somepassword
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}

			_, err = bufio.NewReader(conn).ReadString('\n')
			if err != nil {
				fmt.Println(err)
				return
			}

			for z := 0; z < 50; z++ {
				eMu.Lock()
				e += 1
				eMu.Unlock()
				_, err = conn.Write([]byte(fmt.Sprintf("PUT->key%d->value%d\n", e, e))) // we are using a user of alex and password of somepassword
				if err != nil {
					fmt.Println(err)
					return
				}
			}

			// Read from the connection untill a new line is send
			_, err = bufio.NewReader(conn).ReadString('\n')
			if err != nil {
				fmt.Println(err)
				return
			}

			conn.Close()
		}(i)
	}

	wg.Wait()

	elapsed := time.Since(start)
	fmt.Println("Fin")

	log.Printf("ChromoDB took to insert 5000 keys with 100 connections: %s", elapsed)
}

// Inserts 5000 keys linearly
func insertSingleConnection() {
	start := time.Now()

	tcpAddr, err := net.ResolveTCPAddr("tcp", "localhost:7676")
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	// Connect to the address with tcp
	conn, err := net.DialTCP("tcp", nil, tcpAddr)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	// Send a message to the ChromoDB running instance
	_, err = conn.Write([]byte("YWxleFwwc29tZXBhc3N3b3Jk\n")) // we are using a user of alex and password of somepassword
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	_, err = bufio.NewReader(conn).ReadString('\n')
	if err != nil {
		fmt.Println(err)
		return
	}

	for i := 0; i < 5000; i++ {
		_, err = conn.Write([]byte(fmt.Sprintf("PUT->key%d->value%d\n", i, i))) // we are using a user of alex and password of somepassword
		if err != nil {
			fmt.Println(err)
			return
		}

		// Read from the connection untill a new line is send
		_, err = bufio.NewReader(conn).ReadString('\n')
		if err != nil {
			fmt.Println(err)
			return
		}
	}

	conn.Close()
	elapsed := time.Since(start)
	log.Printf("ChromoDB took to insert 5000 keys with a single connection: %s", elapsed)
}

// Updates key1 with 500 connections
func testConsistency() {

	// Resolve the string address to a TCP address
	tcpAddr, err := net.ResolveTCPAddr("tcp", "localhost:7676")
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	// Connect to the address with tcp
	conn, err := net.DialTCP("tcp", nil, tcpAddr)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	// Send a message to the ChromoDB running instance
	_, err = conn.Write([]byte("YWxleFwwc29tZXBhc3N3b3Jk\n")) // we are using a user of alex and password of somepassword
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	_, err = bufio.NewReader(conn).ReadString('\n')
	if err != nil {
		fmt.Println(err)
		return
	}
	for i := 0; i < 500; i++ {

		_, err = conn.Write([]byte(fmt.Sprintf("PUT->key%d->value%d\n", 1, i))) // we are using a user of alex and password of somepassword
		if err != nil {
			fmt.Println(err)
			return
		}

	}

	conn.Close()
}

// Checks value of key1 after parallel connection check for consistency
func testConsistencyAfter() {
	// Single connection
	tcpAddr, err := net.ResolveTCPAddr("tcp", "localhost:7676")
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	// Connect to the address with tcp
	conn, err := net.DialTCP("tcp", nil, tcpAddr)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	// Send a message to the ChromoDB running instance
	_, err = conn.Write([]byte("YWxleFwwc29tZXBhc3N3b3Jk\n")) // we are using a user of alex and password of somepassword
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	_, err = bufio.NewReader(conn).ReadString('\n')
	if err != nil {
		fmt.Println(err)
		return
	}

	_, err = conn.Write([]byte("get->key1\n")) // we are using a user of alex and password of somepassword
	if err != nil {
		fmt.Println(err)
		return
	}

	res, err := bufio.NewReader(conn).ReadString('\n')
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println(res)

	conn.Close()
}

// Inserts large key value
func insertLargeKeyValue() {
	// Single connection
	tcpAddr, err := net.ResolveTCPAddr("tcp", "localhost:7676")
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	// Connect to the address with tcp
	conn, err := net.DialTCP("tcp", nil, tcpAddr)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	// Send a message to the ChromoDB running instance
	_, err = conn.Write([]byte("YWxleFwwc29tZXBhc3N3b3Jk\n")) // we are using a user of alex and password of somepassword
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	_, err = bufio.NewReader(conn).ReadString('\n')
	if err != nil {
		fmt.Println(err)
		return
	}

	var testVal []byte

	for i := 0; i < 1000; i++ {
		testVal = append(testVal, byte(i))
	}

	_, err = conn.Write([]byte(fmt.Sprintf("put->long_key_name_test->%v\n", testVal)))
	if err != nil {
		fmt.Println(err)
		return
	}

	_, err = bufio.NewReader(conn).ReadString('\n')
	if err != nil {
		fmt.Println(err)
		return
	}

	_, err = conn.Write([]byte("get->long_key_name_test\n")) // we are using a user of alex and password of somepassword
	if err != nil {
		fmt.Println(err)
		return
	}

	res, err := bufio.NewReader(conn).ReadString('\n')
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println(res)

	conn.Close()
}

// Make sure you have a local database running
func main() {

	insertParallel()
	insertParallel2()
	insertSingleConnection()
	testConsistency()
	testConsistencyAfter() // should be 499
	insertLargeKeyValue()

}
