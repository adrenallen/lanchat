//Garrett Allen
package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"net"
	"os"
	"regexp"
	"strconv"
	"strings" // only needed below for sample processing
	"time"
	"unicode/utf8"
)

var username string
var chatPort int
var reader *bufio.Reader = bufio.NewReader(os.Stdin)
var networks []string
var chatPeers []string
var chatHistory []string
var myIPs []string
var debugFlag *bool

type MessageObj struct {
	Ident string `json:"Ident"`
	Data  string `json:"Data"`
}

func main() {
	// fmt.Printf("%v", getMyIPs())
	fmt.Printf("### Welcome to lanchat! ###\n")
	username = getUsername()
	chatPort = getChatPort()

	debugFlag = flag.Bool("debug", false, "Output debug info")
	flag.Parse()

	networks = getMyIPs()

	//set my ip list for avoiding loopbacks
	r, _ := regexp.Compile("[0-9]*\\.[0-9]*\\.[0-9]*\\.[0-9]*")
	for _, myip := range networks {
		newAddr := r.FindString(myip)
		myIPs = append(myIPs, newAddr)
	}

	go findPeers(networks)

	fmt.Println("\nJoining the chat room\n")

	go server()
	go client()

	for {
	}

}

func addPeerToList(addr string) {
	for _, myip := range myIPs {
		if myip == addr {
			return
		}
	}
	chatPeers = append(chatPeers, addr)

}

//higher level func to loop the IPs we find on our interface network
func findPeers(networks []string) {
	for _, netInt := range networks {
		// fmt.Printf("\nSearching for peers on network of %v port %v as %v", netInt, chatPort, username)

		netAddresses := getIPAddressFromNetwork(netInt)
		for _, netAd := range netAddresses {
			if ok := pingAddressForListen(netAd); ok {
				addPeerToList(netAd)
			}
		}

	}

}

//Attempt comms to this IP to see if they wanna talk!
func pingAddressForListen(netAddr string) bool {
	conn, err := net.DialTimeout("tcp", netAddr+":"+strconv.Itoa(chatPort), time.Millisecond*100)
	if *debugFlag {
		if err != nil {
			fmt.Println(err)
		} else {
			fmt.Printf("Found listener on %v\n", netAddr)
		}
	}

	if err != nil {
		return false
	}

	msgA := &MessageObj{"join", username}
	dataEnc, err := json.Marshal(msgA)
	if err != nil {
		panic(err)
	}

	fmt.Fprintf(conn, string(dataEnc))

	conn.Close()

	return true

}

//Get the user's computer IP's
func getMyIPs() []string {
	var retList []string
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		os.Stderr.WriteString("Oops: " + err.Error() + "\n")
		os.Exit(1)
	}

	for _, a := range addrs {
		if ipnet, ok := a.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			retList = append(retList, a.String())
		}
	}

	if *debugFlag {
		fmt.Printf("Found interface IPs of %v \n", retList)
	}

	return retList
}

//Ask user for preferred port for chat
func getChatPort() int {
	un := 9002
	fmt.Printf("Enter chat port (Default %v): ", un)
	if userIn, _ := reader.ReadString('\n'); userIn != "" {
		userIn = strings.TrimRight(userIn, "\n")
		userInParse, err := strconv.ParseInt(userIn, 10, 64)
		if err == nil {
			return int(userInParse)
		}
	}
	return un
}

//Get the list of IPs based on the IP on the PC interface
func getIPAddressFromNetwork(netString string) []string {

	var retList []string
	if strings.Contains(netString, "127.0.0.1") {

		// retList = append(retList, "127.0.0.1")
		return retList

	}

	ip, ipnet, err := net.ParseCIDR(netString)
	if err != nil {
		fmt.Println(err)
	}
	for ip := ip.Mask(ipnet.Mask); ipnet.Contains(ip); inc(ip) {
		retList = append(retList, ip.String())
	}
	return retList
}

//IP incrementer
func inc(ip net.IP) {
	for j := len(ip) - 1; j >= 0; j-- {
		ip[j]++
		if ip[j] > 0 {
			break
		}
	}
}

//ask user for a username
func getUsername() string {

	fmt.Print("Enter your username: ")
	un, _ := reader.ReadString('\n')
	un = strings.TrimRight(un, "\n")
	strLen := utf8.RuneCountInString(un)
	for strLen < 3 {
		fmt.Printf("\nUsername must be 3 characters! You entered %v.\nEnter your username: ", strconv.Itoa(strLen))
		un, _ = reader.ReadString('\n')
		un = strings.TrimRight(un, "\n")
		strLen = utf8.RuneCountInString(un)
	}
	return un
}

//the server or listener portion of the app
func server() {
	// fmt.Println("Launching server...")

	// listen on all interfaces
	ln, err := net.Listen("tcp", ":"+strconv.Itoa(chatPort))

	if err != nil {
		panic(err)
	}

	// accept connection on port

	// run loop forever (or until ctrl-c)
	for {
		conn, _ := ln.Accept()
		checkForNewAddress(conn.LocalAddr().String())
		message, _ := bufio.NewReader(conn).ReadString('\n')

		// fmt.Printf("\n\nGOT\n %v \n\n", message)

		var dat map[string]string
		message = strings.TrimRight(message, "\n")

		if err := json.Unmarshal([]byte(message), &dat); err != nil {
			panic(err)
		}
		switch dat["Ident"] {
		case "message":
			chatHistory = append(chatHistory, string(message))
			fmt.Printf("%v", dat["Data"])
		case "join":
			chatHistory = append(chatHistory, string(message))
			fmt.Printf("%v", dat["Data"]+" has joined the chat\n")
		default:
			panic("Received unknown message type")
		}

		conn.Close()

	}
}

//check the IP we just received from to make sure we have it in our active list
func checkForNewAddress(addr string) bool {

	r, _ := regexp.Compile("[0-9]*\\.[0-9]*\\.[0-9]*\\.[0-9]*")
	newAddr := r.FindString(addr)
	for _, oldAddr := range chatPeers {
		if oldAddr == newAddr {
			return false
		}
	}

	addPeerToList(newAddr)

	return true
}

//the talker part of the app
func client() {
	// connect to this socket

	for {

		// read in input from stdin
		reader := bufio.NewReader(os.Stdin)
		text, _ := reader.ReadString('\n')

		go sendMessage(text)

	}
}

//function to send a message to all of our active list
func sendMessage(text string) {

	if utf8.RuneCountInString(text) < 1 {
		return
	}

	msgA := &MessageObj{"message", text}
	dataEnc, err := json.Marshal(msgA)
	if err != nil {
		panic(err)
	}

	for _, addr := range chatPeers {
		conn, err := net.Dial("tcp", addr+":"+strconv.Itoa(chatPort))
		if err != nil {
			if *debugFlag {
				fmt.Print(err)
			}
		} else {
			// msg := username + " - " + text + "\n"
			fmt.Fprintf(conn, string(dataEnc))
			chatHistory = append(chatHistory, string(dataEnc))
			conn.Close()
		}

	}

}
