package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"strconv"
	"strings" // only needed below for sample processing
	"time"
)

var username string
var chatPort int
var reader *bufio.Reader = bufio.NewReader(os.Stdin)
var networks []string
var chatPeers []chatPeer

type chatPeer struct {
	loginTime time.Time
	username  string
	ip        string
}

func main() {
	// fmt.Printf("%v", getMyIPs())
	// networks = getMyIPs()
	// username = getUsername()
	chatPort = getChatPort()
	go server()
	// pingAddressForListen("127.0.0.1")
	// findPeers(networks)

	go client()
	for {
	}

}

func findPeers(networks []string) {
	for _, netInt := range networks {
		fmt.Printf("\nSearching for peers on network of %v port %v as %v", netInt, chatPort, username)

		netAddresses := getIPAddressFromNetwork(netInt)

		for _, netAd := range netAddresses {
			if _, ok := pingAddressForListen(netAd); ok {
				fmt.Print("Found")
			} else {
				fmt.Print(".")
			}
		}

	}

}

func pingAddressForListen(netAddr string) (chatPeer, bool) {
	conn, err := net.Dial("tcp", netAddr+":"+strconv.Itoa(chatPort))
	if err != nil {
		fmt.Print(err)
		return chatPeer{}, false
	}

	fmt.Printf("found one at %v", netAddr)

	conn.Close()

	return chatPeer{}, false

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
				prefixString := strconv.Itoa(prefixSize)
				retList = append(retList, ipnet.IP.String()+"/"+prefixString)
			}
		}
	}
	return retList
}

func getChatPort() int {
	fmt.Print("Enter chat port: ")
	un := 9002

	return un
}

func getIPAddressFromNetwork(netString string) []string {
	var retList []string
	ip, ipnet, err := net.ParseCIDR(netString)
	if err != nil {
		fmt.Println(err)
	}
	for ip := ip.Mask(ipnet.Mask); ipnet.Contains(ip); inc(ip) {
		retList = append(retList, string(ip))
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
	return un
}

func server() {
	fmt.Println("Launching server...")

	// listen on all interfaces
	ln, _ := net.Listen("tcp", ":"+strconv.Itoa(chatPort))

	// accept connection on port

	// run loop forever (or until ctrl-c)
	for {
		conn, _ := ln.Accept()
		message, _ := bufio.NewReader(conn).ReadString('\n')
		fmt.Printf("Got %v\n", message)
		// fmt.Print("got it")
		conn.Close()
		// go testReceive(conn)
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
		conn, _ := net.Dial("tcp", "127.0.0.1:8081")
		// send to socket
		fmt.Fprintf(conn, text+"\n")
		// listen for reply
		// message, _ := bufio.NewReader(conn).ReadString('\n')
		// fmt.Print("User: "+message)
		conn.Close()
	}
}
