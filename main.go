package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net"
	"os"
	"strconv"
	"strings"
	"time"
)

/*
	Store global variables
	lines 25 & 27 from this YouTube video https://www.youtube.com/watch?v=waZt518cxFI (4m45s)
*/
var (
	configData = readConfig()
	t          = time.Now()
	myTime     = t.Format(time.RFC3339) + "\n"
	// Create a map of connections and assign value "true"
	OpenConnections = make(map[net.Conn]bool)
	// Create a channel that accepts or passes a new connection
	newConnection = make(chan net.Conn)
)

// Store values of the message in a struct
type Message struct {
	messageContent     string
	senderID           string
	destinationID      string
	destinationAddress string
}

func main() {
	// Take user input for which line from the config file to use for process ID, port, and localhost
	arguments := os.Args
	if len(arguments) == 1 {
		fmt.Println("Please provide line number of config file")
		return
	}
	lineNumber := arguments[1]
	// Identify the port number to use
	port := ":" + configData[lineNumber][2]

	// Launch a concurrent TCP Server
	go createTCPServer(port)
	println("Please provide a command in the form send destination message or STOP to stop proccess")
	for {
		// Take user input
		reader := bufio.NewReader(os.Stdin)
		fmt.Print(">> ")
		text, _ := reader.ReadString('\n')
		// Parse the text from the reader
		textParsed := parseLine(text)
		if len(textParsed) < 3 {
			println("Please provide a command in the form send destination message or STOP to stop proccess")
			return
		}
		destination := textParsed[1]
		messageReceived := textParsed[2]
		if len(arguments) > 2 {
			for i := 3; i < len(arguments); i++ {
				messageReceived = messageReceived + " " + textParsed[i]
			}
		}
		destinationAdress := configData[destination][1]
		// Create a new struct to store all the variables
		m := Message{messageReceived, lineNumber, destination, destinationAdress}
		delays := configData["0"]
		delay, err := sliceAtoi(delays)
		if err != nil {
			fmt.Println(err)
			return
		}
		// Create a TCP Client
		createTCPClient(m, delay)
	}
}

// *UPDATE LINE NUMBERS BELOW*
// Lines 40-46 & 52-54 from https://stackoverflow.com/questions/8757389/reading-a-file-line-by-line-in-go
// Stores the config data into a hashmap key is the line number value is an array with the data arr[0] = ID, arr[1] = hostaddress arr[2] = port
func readConfig() map[string][]string {
	// open "config.txt" file
	file, err := os.Open("config.txt")
	if err != nil {
		log.Fatal(err)
	}
	// Delay closing of the file until other functions return
	defer file.Close()

	configData := make(map[string][]string)
	scanner := bufio.NewScanner(file)
	currentLineNum := 0
	configLine := ""
	for scanner.Scan() {
		configLine = string(scanner.Text())
		configLineParsed := parseLine(configLine)
		configData[strconv.Itoa(currentLineNum)] = configLineParsed
		currentLineNum++
	}

	if err = scanner.Err(); err != nil {
		log.Fatal(err)
	}
	return configData
}

// parseLine parses the lines and stores them in an array
func parseLine(line string) []string {
	return strings.Split(line, " ")
}

// all lines from https://stackoverflow.com/questions/24972950/go-convert-strings-in-array-to-integer
// sliceAtoi converts a string array into an int array
func sliceAtoi(str []string) ([]int, error) {
	intarr := make([]int, 0, len(str))
	for _, a := range str {
		i, err := strconv.Atoi(a)
		if err != nil {
			return intarr, err
		}
		intarr = append(intarr, i)
	}
	return intarr, nil
}

/*
	Creates TCP server using first command line argument
	most of the code is from here https://www.linode.com/docs/guides/developing-udp-and-tcp-clients-and-servers-in-go/
	lines 158 & 160 is from https://www.youtube.com/watch?v=waZt518cxFI (7m00s)
	lines 165 & 167 is from https://www.youtube.com/watch?v=8Epy35neq9M (4m55s)
*/
func createTCPServer(PORT string) {
	// Listen announces on the local network address
	l, err := net.Listen("tcp", PORT)
	if err != nil {
		fmt.Println(err)
		return
	}
	// delay closing the server until other functions return
	defer l.Close()

	// Use Go routine to run concurrently with the server
	go func() {
		for {
			// Accept connection interface from every new clients
			c, err := l.Accept()
			fmt.Println("Remote address: ", c.RemoteAddr())
			if err != nil {
				fmt.Println(err)
				return
			}
			// Add the new connection to the openConnections map
			OpenConnections[c] = true
			// Pass the new connections into the newConnection channel
			newConnection <- c
		}
	}()
	for {
		// The connection is equal to the connection that is passed to the new connection channel
		c := <-newConnection
		// Invoke unicastReceive function to receive messages from other processes
		go unicastReceive(c)
	}
}

/*
	Creates TCP client using second command line argument most of the general code such as reader, net.Dial
	is from here: https://www.linode.com/docs/guides/developing-udp-and-tcp-clients-and-servers-in-go/
*/
func createTCPClient(inputMessage Message, delay []int) {
	CONNECT := inputMessage.destinationAddress
	messageContent := inputMessage.messageContent
	userID := inputMessage.senderID
	destinationID := inputMessage.destinationID
	destinationAddress := inputMessage.destinationAddress

	// Dial connects to the address on the named network
	c, err := net.Dial("tcp", CONNECT)
	if err != nil {
		fmt.Println(err)
		return
	}
	// Read user input
	reader := bufio.NewReader(os.Stdin)
	for {
		newMessage := Message{messageContent: messageContent, destinationID: destinationID, destinationAddress: destinationAddress, senderID: userID}
		// Invoke unicastSend function to send the message to other processes with the delay
		unicastSend(c, newMessage, delay)
		readMessage, _ := reader.ReadString('\n')
		messageParsed := parseLine(readMessage)
		if len(messageParsed) < 3 {
			println("Please provide a command in the form send destination message or STOP to stop proccess")
			return
		}
		// Case for handling if the message is more than one word
		messageReceived := messageParsed[2]
		if len(messageParsed) > 2 {
			for i := 3; i < len(messageParsed); i++ {
				messageReceived = messageReceived + " " + messageParsed[i]
			}
		}
		messageContent = messageReceived
		//
		if messageParsed[1] != destinationID {
			messageContent = messageReceived
			destinationID = messageParsed[1]
			destinationAddress = configData[messageParsed[1]][1]
			c, err = net.Dial("tcp", destinationAddress)
			if err != nil {
				fmt.Println(err)
				return
			}
		}
	}
}

// unicastSend sends (struct) message to the other processes with the delay and prints upon completion
func unicastSend(c net.Conn, inputMessage Message, delay []int) {
	message := inputMessage.messageContent
	id := inputMessage.destinationID
	sender := inputMessage.senderID

	// "STOP" to exit the program
	if strings.TrimSpace(string(message)) == "STOP" {
		fmt.Println("TCP client exiting...")
		return
	}

	// Go routine for implementing the delay, sending the message, and printing values
	go func() {
		// Set minimum and maximum delay values
		minDelay := delay[0]
		maxDelay := delay[1]
		// Generate a random integer between min and max delay value
		delayed := rand.Intn(maxDelay-minDelay+1) + minDelay
		// Implement delay with sleep function by multiplying delay value by time
		time.Sleep(time.Duration(delayed) * time.Millisecond)
		message = strings.TrimSpace(message)
		newMessage := sender + " " + message
		sendMessage := message
		_, err := fmt.Fprintf(c, newMessage+"\n")
		if err != nil {
			fmt.Println("Error with client", err)
			return
		}
		// Write data to the connection in bytes
		c.Write([]byte(sendMessage))
		fmt.Printf("Sent '%s' to process %s, system time is %s\n", sendMessage, id, myTime)
		fmt.Print(">> ")
	}()
}

/*
	unicastReceive receives messages from other processes, parses them and stores them into variables,
	identifies which process sent the message, and prints accordingly
*/
func unicastReceive(c net.Conn) {
	for {
		reader := bufio.NewReader(c)
		message, err := reader.ReadString('\n')
		parsedMessage := parseLine(message)
		processID := parsedMessage[0]
		receivedMessage := parsedMessage[1]
		// Case for handling if the message is more than one word
		if len(parsedMessage) > 2 {
			for i := 2; i < len(parsedMessage); i++ {
				receivedMessage = receivedMessage + " " + parsedMessage[i]
			}
		}
		receivedMessage = strings.TrimSpace(receivedMessage)
		if err == io.EOF {
			c.Close()
			fmt.Println("Connection closed. ")
			os.Exit(0)
		}
		message = strings.TrimSpace(message)
		fmt.Printf("Received '%s' from process %s, system time is %s\n", receivedMessage, processID, myTime)
		fmt.Print(">> ")
	}
}
