# mp1

## Instructions
    Download mp1 directory. 
    Once inside the mp1 directory run the command go run main.go {process id} to start a process. 
    Use the command send {destination} {message} to send a message to the server of the destination process.

Currently there are three processes specified in the config file. Therefore you can have three terminal windows open to have three processes running by executing the go run go main.go {process id} command, and have them all communicate with each other.

## Documentation
Running the main.go file followed by a process number creates a server for the process. The program then reads the user input, which should be in the form, send destination message. It launches a concurrent TCP server using Goroutines and passes new connections to the channel. Then, the program creates a client with a connection to the destination using create TCPClient. The client runs the unicastSend function to send the message to the destination server along with the process id. 

The unicastSend function has the tcp connection, a message struct, and the delay range as inputs. It uses the tcp connection to send the message to the server, the message struct to retrieve the message information (the message, host address, destination id, and sender id), and the delay range to determine how long to delay the message. It implements the delay by calling on the sleep function, multiplying time in milliseconds by the random generated value between the range of minimum and maximum delay values listed in the config.txt file. Finally, It prints out the message, destination or process ID, and time. 

The unicastReceive function has the connection created by the server as input. It uses this connection to read the message from the client, and print out that it received a message, the message it received, the sender id, and the time.

When there are additional commands with a different destination, the program updates the connection in the client so that the unicastSend function can send the message to the correct server. 

## Code Sources, Citations & Explanations

https://www.linode.com/docs/guides/developing-udp-and-tcp-clients-and-servers-in-go/ <br />
Some of the code to create the client and server comes from here. 
The client creates a connection with the server using net.Dial function which takes in the protocol type (tcp), and the host address of the server. Once a connection has been made, the client function reads the user input and sends it to the server, using the fPrintf function which takes in the connection with the server and the message as inputs. 

https://www.youtube.com/watch?v=waZt518cxFI (4m45s, 7m00s) <br />
https://www.youtube.com/watch?v=8Epy35neq9M (4m55s) <br />
Code lines 25, 27, 158, 160, 165, 167 are from this YouTube video. 

A concurrent TCP server is launched using Goroutines and channels. The program creates a channel that passes a new connection and adds it to the map of connections called openConnections, which assigns a boolean value “true.” It then passes new connections into the newConnection channel, and when the connection is equal to the one that is passed into the new connection channel, it invokes unicastReceive function to receive messages from other processes. 

https://stackoverflow.com/questions/24972950/go-convert-strings-in-array-to-integer <br />
All lines from the sliceAtoi function are from this website. 
The sliceAtoi function takes in a string array and returns an integer array by looping through the range of the string array, converting the string value to integer value, and appending the converted value to the new integer array. 

https://stackoverflow.com/questions/8757389/reading-a-file-line-by-line-in-go <br />
All lines from the sliceAtoi function are from this website. 
It opens the file using os.Open(path of file), and then reads it line by line using a scanner and for loop. 
