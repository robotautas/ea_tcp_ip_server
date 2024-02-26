package main

import (
	"bufio"
	"fmt"
	"io"
	"math/rand"
	"net"
	"os"
	"strconv"
	"strings"
	"time"
)

var names = []string{
	"sample 1",
	"sample 2",
	"sample 3",
	"sample 4",
	"sample 5",
	"sample 6",
	"sample 7",
	"sample 8",
	"sample 9",
	"sample 10",
}

func main() {
	// Listen on TCP port 8080
	ln, err := net.Listen("tcp", ":1984")
	if err != nil {
		fmt.Println(err)
		return
	}
	// defer ln.Close()

	fmt.Println("TCP server listening on port 1984")

	for {
		// Accept a connection
		println("\nWaiting for connection...")
		conn, err := ln.Accept()
		println(conn)
		if err != nil {
			fmt.Println(err)
			continue
		}

		// Handle the connection in a new goroutine
		go handleConnection(conn)
	}
}

func handleConnection(conn net.Conn) {
	reader := bufio.NewReader(conn)
	defer conn.Close()

	for {
		message, err := reader.ReadString('\n')
		if err != nil {
			if err != io.EOF {
				fmt.Println("Error reading:", err.Error())
			} else {
				fmt.Println("Client disconnected")
				return
			}
			break
		}

		fmt.Printf("Message received: '%s'", message)
		// Add any additional handling for the received message here

		// Optionally, send a response back to the client
		respond(message, conn)

	}
}

func respond(m string, conn net.Conn) {
	switch m {
	case "?STS\r\n":
		conn.Write([]byte("STS READY"))
	case "STRT\r\n":
		go startSequence(conn)
	default:
		if strings.HasPrefix(m, "?NAM ") {
			fmt.Printf("REQUEST WAS: %s", m)
			numberStr := strings.TrimSpace(m[4:])
			number, err := strconv.Atoi(numberStr)
			if err != nil {
				// Handle conversion error
				fmt.Printf("Failed to parse number: %v\n", err)
				return
			}
			response := generateStringWithNumber(number)
			formatted_response := fmt.Sprintf("NAM %d %s", number, response)
			println(formatted_response)
			conn.Write([]byte(formatted_response))
		} else if strings.HasPrefix(m, "?WGH ") {
			numberStr := strings.TrimSpace(m[4:])
			number, err := strconv.Atoi(numberStr)
			if err != nil {
				// Handle conversion error
				fmt.Printf("Failed to parse number: %v\n", err)
				return
			}
			conn.Write([]byte(fmt.Sprintf("WGH %d 1.0%d", number, number)))
		} else if strings.HasPrefix(m, "?PCT ") {
			cleanRequest := strings.Replace(m, "\r\n", "", -1)
			splittedRequest := strings.Split(cleanRequest, " ")
			element := splittedRequest[2]
			numberStr := splittedRequest[1]

			number, err := strconv.Atoi(numberStr)
			if err != nil {
				// Handle conversion error
				fmt.Printf("Failed to parse number: %v\n", err)
				return
			}
			conn.Write([]byte(fmt.Sprintf("PCT %v %v 1.00%v", number, element, number)))
		} else {
			conn.Write([]byte(fmt.Sprintf("Response message to %s\n", m)))
		}
	}
}

func startSequence(conn net.Conn) {
	conn.Write([]byte("STARTED SEQUENCE!\n"))
	file, err := os.Open("sequence.log")
	if err != nil {
		fmt.Println("Error opening file:", err)
		return
	}
	defer file.Close()

	var lastTime time.Time
	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		line := scanner.Text()
		parts := strings.SplitN(line, "\t", 2)
		println(line)

		if len(parts) < 2 {
			fmt.Println("Invalid line format:", line)
			continue
		}

		currentTime, err := time.Parse("15:04:05", parts[0])
		if err != nil {
			fmt.Println("Error parsing timestamp:", err)
			continue
		}

		if !lastTime.IsZero() {
			waitDuration := currentTime.Sub(lastTime)
			if waitDuration > 0 {
				time.Sleep(waitDuration)
			}
		}

		lastTime = currentTime

		fmt.Println(parts[1]) // Print the line excluding the timestamp
		conn.Write([]byte(fmt.Sprintf("%s\r\n", parts[1])))
	}

	if err := scanner.Err(); err != nil {
		fmt.Println("Error reading file:", err)
	}
}

func generateStringWithNumber(n int) string {
	// Seed the random number generator to ensure different results on each run
	rand.Seed(time.Now().UnixNano())

	// Convert the input integer to string
	numberStr := strconv.Itoa(n)

	// Generate a random number (for example's sake, let's make it between 100 and 999)
	randomNumber := rand.Intn(900) + 100

	// Concatenate the input number's string representation with the random number, prefixed with "random "
	return "random " + numberStr + "-" + strconv.Itoa(randomNumber)
}
