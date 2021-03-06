package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
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

func main() {
	// fmt.Printf("%v", getMyIPs())

	username = getUsername()
	chatPort = getChatPort()

	// pingAddressForListen("127.0.0.1")
	// start := time.Now()

	networks = getMyIPs()
	chatPeers := findPeers(networks)

	fmt.Printf("\nFound the following peers %v", chatPeers)

	// fmt.Printf("Time taken %v", time.Since(start))
	// fmt.Println("\ndone")
	// go server()
	go client()

	for {
	}

}

func findPeers(networks []string) []string {
	for _, netInt := range networks {
		fmt.Printf("\nSearching for peers on network of %v port %v as %v", netInt, chatPort, username)

		netAddresses := getIPAddressFromNetwork(netInt)
		for _, netAd := range netAddresses {
			if ok := pingAddressForListen(netAd); ok {
				chatPeers = append(chatPeers, netAd)
			}
		}

	}

	return chatPeers

}

func pingAddressForListen(netAddr string) bool {
	conn, err := net.DialTimeout("tcp", netAddr+":"+strconv.Itoa(chatPort), time.Microsecond*150)

	if err != nil {
		// fmt.Println(err)
		return false
	}

	fmt.Fprintf(conn, username+" joined the chat\n")

	conn.Close()

	return true

}

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

func getChatPort() int {
	fmt.Print("Enter chat port: ")
	un := 9002
	if userIn, _ := reader.ReadString('\n'); userIn != "" {
		userIn = strings.TrimRight(userIn, "\n")
		userInParse, err := strconv.ParseInt(userIn, 10, 64)
		if err == nil {
			return int(userInParse)
		}
	}
	return un
}

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

func server() {
	// fmt.Println("Launching server...")

	// listen on all interfaces
	ln, _ := net.Listen("tcp", ":"+strconv.Itoa(chatPort))

	// accept connection on port

	// run loop forever (or until ctrl-c)
	for {
		conn, _ := ln.Accept()
		message, _ := bufio.NewReader(conn).ReadString('\n')
		fmt.Printf("%v", message)

		conn.Close()

	}
}

func testReceive(conn net.Conn) {
	// will listen for message to process ending in newline (\n)
	message, _ := bufio.NewReader(conn).ReadString('\n')
	// output message received
	fmt.Print("User: ", string(message))
	// sample process for string received
	newmessage := strings.ToUpper(message)
	// send new string back to client
	conn.Write([]byte(newmessage + "\n"))
	conn.Close()
}

func client() {
	// connect to this socket

	for {

		// read in input from stdin
		reader := bufio.NewReader(os.Stdin)
		text, _ := reader.ReadString('\n')

		go sendMessage(text)

	}
}

func sendMessage(text string) {

	if utf8.RuneCountInString(text) < 1 {
		return
	}

	for _, addr := range chatPeers {
		conn, _ := net.Dial("tcp", addr+":"+strconv.Itoa(chatPort))
		msg := username + " - " + text + "\n"
		fmt.Fprintf(conn, msg)
		chatHistory = append(chatHistory, msg)
		conn.Close()
	}

}
