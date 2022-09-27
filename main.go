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

var (
	configData = readConfig()
	t          = time.Now()
	myTime     = t.Format(time.RFC3339) + "\n"
	//
	OpenConnections = make(map[net.Conn]bool)
	newConnection   = make(chan net.Conn)
)

type Message struct {
	messageContent     string
	senderID           string
	destinationID      string
	destinationAddress string
}

func main() {
	arguments := os.Args
	if len(arguments) == 1 {
		fmt.Println("Please provide line number of config file")
		return
	}
	lineNumber := arguments[1]
	port := ":" + configData[lineNumber][2]

	go createTCPServer(port)
	println("Please provide a command in the form send destination message or STOP to stop proccess")
	for {
		reader := bufio.NewReader(os.Stdin)
		fmt.Print(">> ")
		text, _ := reader.ReadString('\n')
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
		m := Message{messageReceived, lineNumber, destination, destinationAdress}
		delays := configData["0"]
		delay, err := sliceAtoi(delays)
		if err != nil {
			fmt.Println(err)
			return
		}
		createTCPClient(m, delay)
	}
}

// Lines 40-46 & 52-54 from https://stackoverflow.com/questions/8757389/reading-a-file-line-by-line-in-go
// Stores the config data into a hashmap key is the line number value is an array with the data arr[0] = ID, arr[1] = hostaddress arr[2] = port
func readConfig() map[string][]string {
	file, err := os.Open("config.txt")
	if err != nil {
		log.Fatal(err)
	}
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

// Parses line and stores it in an array
func parseLine(line string) []string {
	return strings.Split(line, " ")
}

//
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
	Creates TCP server using first command line argument most of the code is from
here https://www.linode.com/docs/guides/developing-udp-and-tcp-clients-and-servers-in-go/
*/
func createTCPServer(PORT string) {
	l, err := net.Listen("tcp", PORT)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer l.Close()
	go func() {
		for {
			c, err := l.Accept()
			fmt.Println("Remote address: ", c.RemoteAddr())
			if err != nil {
				fmt.Println(err)
				return
			}
			//
			OpenConnections[c] = true
			newConnection <- c
		}
	}()
	for {
		//
		c := <-newConnection
		go unicastReceive(c)
	}
}

/*
	Creates TCP client using second command line argument most of the code is from here
https://www.linode.com/docs/guides/developing-udp-and-tcp-clients-and-servers-in-go/
*/
func createTCPClient(inputMessage Message, delay []int) {
	CONNECT := inputMessage.destinationAddress
	messageContent := inputMessage.messageContent
	userID := inputMessage.senderID
	destinationID := inputMessage.destinationID
	destinationAddress := inputMessage.destinationAddress

	c, err := net.Dial("tcp", CONNECT)
	if err != nil {
		fmt.Println(err)
		return
	}
	reader := bufio.NewReader(os.Stdin)
	for {
		newMessage := Message{messageContent: messageContent, destinationID: destinationID, destinationAddress: destinationAddress, senderID: userID}
		//
		unicastSend(c, newMessage, delay)
		readMessage, _ := reader.ReadString('\n')
		messageParsed := parseLine(readMessage)
		if len(messageParsed) < 3 {
			println("Please provide a command in the form send destination message or STOP to stop proccess")
			return
		}
		messageReceived := messageParsed[2]
		if len(messageParsed) > 2 {
			for i := 3; i < len(messageParsed); i++ {
				messageReceived = messageReceived + " " + messageParsed[i]
			}
		}
		messageContent = messageReceived
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

//
func unicastSend(c net.Conn, inputMessage Message, delay []int) {
	message := inputMessage.messageContent
	id := inputMessage.destinationID
	sender := inputMessage.senderID

	if strings.TrimSpace(string(message)) == "STOP" {
		fmt.Println("TCP client exiting...")
		return
	}
	//
	go func() {
		minDelay := delay[0]
		maxDelay := delay[1]
		delayed := rand.Intn(maxDelay-minDelay+1) + minDelay
		time.Sleep(time.Duration(delayed) * time.Millisecond)
		message = strings.TrimSpace(message)
		newMessage := sender + " " + message
		sendMessage := message
		_, err := fmt.Fprintf(c, newMessage+"\n")
		if err != nil {
			fmt.Println("Error with client", err)
			return
		}
		c.Write([]byte(sendMessage))
		fmt.Printf("Sent '%s' to process %s, system time is %s\n", sendMessage, id, myTime)
		fmt.Print(">> ")
	}()
}

//
func unicastReceive(c net.Conn) {
	for {
		reader := bufio.NewReader(c)
		message, err := reader.ReadString('\n')
		parsedMessage := parseLine(message)
		processID := parsedMessage[0]
		receivedMessage := parsedMessage[1]
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
