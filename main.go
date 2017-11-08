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
var chatPeers []ChatPeer
var chatHistory []string
var myIPs []string
var debugFlag *bool
var selfConnect *bool

const defaultName string = "anon"

type MessageObj struct {
	Ident string `json:"Ident"`
	Data  string `json:"Data"`
}

type ChatPeer struct {
	Username string
	Address  string
}

func main() {
	fmt.Printf("### Welcome to lanchat! ###\n")
	username = getUsername()
	chatPort = getChatPort()

	debugFlag = flag.Bool("debug", false, "Output debug info")
	selfConnect = flag.Bool("selfconnect", false, "Connect only to itself for testing")
	useServer := flag.Bool("server", true, "Enabling listening server")
	flag.Parse()

	networks = getMyIPs()

	//set my ip list for avoiding loopbacks
	r, _ := regexp.Compile("[0-9]*\\.[0-9]*\\.[0-9]*\\.[0-9]*")
	for _, myip := range networks {
		newAddr := r.FindString(myip)
		myIPs = append(myIPs, newAddr)

	}

	fmt.Println("\nJoining the chat room\n")

	if *useServer {
		go server()
	}

	go client()

	if !*selfConnect {
		// go findPeers(networks)
	}

	pingAddressForListen("10.1.1.193")
	if *selfConnect {
		for _, myip := range myIPs {
			pingAddressForListen(myip)
		}
	}

	for {
	}

}

func addPeerToList(user string, addr string) {
	if !*selfConnect {
		for _, myip := range myIPs {
			if myip == addr {
				return
			}
		}
	}

	user = strings.TrimSpace(user)
	addr = strings.TrimSpace(addr)

	for idx, cp := range chatPeers {
		if cp.Address == addr {
			if cp.Username != user {
				chatPeers[idx].Username = user
				fmt.Printf("[%v] %v identified as %v\n", cp.Address, cp.Username, user)
			}
			return
		}
	}

	chatPeers = append(chatPeers, ChatPeer{user, addr})
	fmt.Printf("[%v] %v has joined the chat\n", addr, user)

}

//higher level func to loop the IPs we find on our interface network
func findPeers(networks []string) {
	for _, netInt := range networks {
		if *debugFlag {
			fmt.Printf("\n Searching for peers on network of %v port %v as %v \n", netInt, chatPort, username)
		}

		netAddresses := getIPAddressFromNetwork(netInt)
		for _, netAd := range netAddresses {
			if ok := pingAddressForListen(netAd); ok {
				addPeerToList(defaultName, netAd)
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

	fmt.Fprintf(conn, "%v\n", string(dataEnc))
	checkForNewAddress(getIPFromString(conn.RemoteAddr().String()))
	message, _ := bufio.NewReader(conn).ReadString('\n')

	// fmt.Printf("\n\nGOT\n %v \n\n", message)

	var dat map[string]string
	message = strings.TrimSpace(message)

	if err := json.Unmarshal([]byte(message), &dat); err != nil {
		panic(err)
	}

	addPeerToList(strings.TrimSpace(dat["Data"]), getIPFromString(conn.RemoteAddr().String()))

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

			if ipnet.IP.To4() != nil {
				retList = append(retList, a.String())
			}

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
		userIn = strings.TrimSpace(userIn)
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
	un = strings.TrimSpace(un)
	strLen := utf8.RuneCountInString(un)
	for strLen < 3 {
		fmt.Printf("\nUsername must be 3 characters! You entered %v.\nEnter your username: ", strconv.Itoa(strLen))
		un, _ = reader.ReadString('\n')
		un = strings.TrimSpace(un)
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
		checkForNewAddress(getIPFromString(conn.RemoteAddr().String()))
		message, _ := bufio.NewReader(conn).ReadString('\n')

		// fmt.Printf("\n\nGOT\n %v \n\n", message)

		var dat map[string]string
		message = strings.TrimSpace(message)

		if err := json.Unmarshal([]byte(message), &dat); err != nil {
			panic(err)
		}
		switch dat["Ident"] {
		case "message":
			fmt.Printf("%v - %v", getIPUsername(getIPFromString(conn.RemoteAddr().String())), dat["Data"])
		case "join":
			addPeerToList(strings.TrimSpace(dat["Data"]), getIPFromString(conn.RemoteAddr().String()))

			msgA := &MessageObj{"join", username}
			dataEnc, err := json.Marshal(msgA)
			if err != nil {
				panic(err)
			}

			fmt.Fprintf(conn, "%v\n", string(dataEnc))
		default:
			panic("Received unknown message type")
		}

		conn.Close()

	}
}

func getIPFromString(addr string) string {
	r, _ := regexp.Compile("[0-9]*\\.[0-9]*\\.[0-9]*\\.[0-9]*")
	return r.FindString(addr)
}

func getIPUsername(addr string) string {
	for _, cp := range chatPeers {
		if cp.Address == addr {
			return cp.Username
		}
	}
	if *debugFlag {
		fmt.Printf("Failed to find match for %v", addr)
	}

	return defaultName
}

//check the IP we just received from to make sure we have it in our active list
func checkForNewAddress(addr string) bool {

	r, _ := regexp.Compile("[0-9]*\\.[0-9]*\\.[0-9]*\\.[0-9]*")
	newAddr := r.FindString(addr)
	for _, cp := range chatPeers {
		oldAddr := cp.Address
		if oldAddr == newAddr {
			return false
		}
	}

	addPeerToList(defaultName, newAddr)

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

	if strings.TrimSpace(text) == "/printnames" {
		fmt.Printf("Peer list:\n")
		for _, cp := range chatPeers {
			fmt.Printf("[%v] %v\n", cp.Address, cp.Username)
		}
		return
	}

	msgA := &MessageObj{"message", text}
	dataEnc, err := json.Marshal(msgA)
	if err != nil {
		panic(err)
	}

	for _, cp := range chatPeers {
		addr := cp.Address
		conn, err := net.Dial("tcp", addr+":"+strconv.Itoa(chatPort))
		if err != nil {
			if *debugFlag {
				fmt.Print(err)
			}
		} else {
			// msg := username + " - " + text + "\n"
			fmt.Fprintf(conn, "%v\n", string(dataEnc))
			chatHistory = append(chatHistory, string(dataEnc))
			conn.Close()
		}

	}

}
