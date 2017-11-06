package main

import "net"
import "fmt"
import "bufio"
import "os"

func main() {
  go client()
  for{}

}

func client(){
  // connect to this socket
  
  for { 
    // read in input from stdin
    reader := bufio.NewReader(os.Stdin)
    fmt.Print("Text to send: ")
    text, _ := reader.ReadString('\n')
    sendMessage(text)
  }
}



func sendMessage(text string){
  conn, _ := net.Dial("tcp", "127.0.0.1:8081")
  // send to socket
  fmt.Fprintf(conn, text + "\n")
  // listen for reply
  message, _ := bufio.NewReader(conn).ReadString('\n')
  fmt.Print("Message from server: "+message)
}