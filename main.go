//Garrett Allen
package main

import (
	"bufio"
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

func main() {
	// fmt.Printf("%v", getMyIPs())
	fmt.Printf("### Welcome to lanchat! ###\n")
	username = getUsername()
	chatPort = getChatPort()

	// pingAddressForListen("192.168.29.113")
	// start := time.Now()
	// return

	networks = getMyIPs()

	//set my ip list for avoiding loopbacks
	r, _ := regexp.Compile("[0-9]*\\.[0-9]*\\.[0-9]*\\.[0-9]*")
	for _, myip := range networks {
		newAddr := r.FindString(myip)
		myIPs = append(myIPs, newAddr)
	}

	go findPeers(networks)

	// fmt.Printf("Time taken %v", time.Since(start))
	fmt.Println("\nJoining the chat room\n")
	go server()
	go client()

	for {
	}

}

func addPeerToList(addr string) {
	for _, myip := range myIPs {
		if myip == addr {
			// return
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

	if err != nil {
		// fmt.Println(err)
		return false
	}

	fmt.Fprintf(conn, username+" joined the chat\n")

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
				// fmt.Print(ipnet.IP.DefaultMask().String())
				mask := net.IPMask(ipnet.IP.DefaultMask()) // If you have the mask as a string
				//mask := net.IPv4Mask(255,255,255,0) // If you have the mask as 4 integer values

				prefixSize, _ := mask.Size()

				if prefixSize >= 24 {
					prefixString := strconv.Itoa(prefixSize)
					retList = append(retList, ipnet.IP.String()+"/"+prefixString)
				}

			}
		}
	}

	//TESTING STUFFZ
	// retList = append(retList, "127.0.0.1/32")
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

		retList = append(retList, "127.0.0.1")
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
	ln, _ := net.Listen("tcp", ":"+strconv.Itoa(chatPort))

	// accept connection on port

	// run loop forever (or until ctrl-c)
	for {
		conn, _ := ln.Accept()
		checkForNewAddress(conn.LocalAddr().String())
		message, _ := bufio.NewReader(conn).ReadString('\n')
		chatHistory = append(chatHistory, message)
		fmt.Printf("%v", message)

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

	for _, addr := range chatPeers {
		conn, err := net.Dial("tcp", addr+":"+strconv.Itoa(chatPort))
		if err != nil {
			// fmt.Print(err)
		} else {
			msg := username + " - " + text + "\n"
			fmt.Fprintf(conn, msg)
			chatHistory = append(chatHistory, msg)
			conn.Close()
		}

	}

}
