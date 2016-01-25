/*
Package goping runs fping for a range of hosts simultaneously
in order to see whether those hosts are alive and how long
the pings are under their way.
This is accomplished by Go Routines. A nice start into concurrency
*/
package main

import (
	"fmt"
	"log"
	"os/exec"
	"strings"
)

const (
	IPStart = "192.168.178."
	IPEnd   = 500
	// Extern  = "8.8.8.8"
	Extern = "www.google.com"
)

//collecting the list of channels into one
func fanIn(input [IPEnd]<-chan string) <-chan string {
	c := make(chan string)
	go func() {
		for {
			for i := 0; i < len(input); i++ {
				c <- <-input[i]
			}
		}
	}()
	return c
}

//pings each ip using a go routine
func pinger(i int) <-chan string {
	c := make(chan string)
	ip := IPStart + fmt.Sprint(i)
	go func() {
		out, _ := exec.Command("fping", "-de",
			ip).CombinedOutput()
		res := string(out)
		if strings.Contains(res, "alive") {
			res = strings.Replace(res, "is alive", " ", 1)
			split := strings.SplitN(res, " ", 3)
			c <- fmt.Sprintf("%15s  %30s %s %s",
				ip, split[0], split[1], split[2])
		} else {
			c <- fmt.Sprint("")
		}
	}()
	return c
}

//now read the results which are piped into the channel
func reader(c <-chan string) {
	count := 0
	for i := 0; i < IPEnd; i++ {
		if res := <-c; res != "" {
			count++
			fmt.Printf("%3d  %s", count, res)
		}
	}
	fmt.Println("found: ", count, "hosts")
}

//check an external IP in order to see that the Internet is reachable
func checkExtern() {
	out0, err := exec.Command("ping", "-c1", Extern).Output()
	if err != nil {
		log.Fatal("extern is not accessible")
	}
	s := strings.SplitAfter(string(out0), "--- ")
	fmt.Printf("%s", s[0])
}

//generate a cannel array running the pings
func fillIpArray() [IPEnd]<-chan string {
	var cs [IPEnd]<-chan string
	for i := 0; i < IPEnd; i++ {
		cs[i] = pinger(i + 1)
	}
	return cs
}

func main() {
	fmt.Println("first pinging my dns server")
	checkExtern()
	fmt.Println("Now running fping as Go-Routines")
	c := fanIn(fillIpArray())
	reader(c)
}
