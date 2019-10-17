package main

import (
	"bytes"
	"errors"
	"flag"
	"fmt"
	"net"
	"os"
	"time"
)

func ipv4Payload(b []byte) []byte {
	if len(b) < 20 {
		return b
	}
	hdrlen := int(b[0]&0x0f) << 2
	return b[hdrlen:]
}

func ping(address string) ([]byte, error) {
	conn, err := net.Dial("ip4:icmp", address)
	if err != nil {
		return nil, err
	}
	err = conn.SetDeadline(time.Now().Add(time.Second * 3))
	if err != nil {
		return nil, err
	}

	//type(8bit) code(8bit) sum(16bit) ID(16bit) Seq(16bit) data
	//type is 8, code is 0, two 0 as placeholder of sum
	b := []byte{8, 0, 0, 0}
	id, seq := os.Getpid()&0xffff, 1
	//append id and seq
	b = append(b, byte(id >> 8), byte(id & 0xff), byte(seq >> 8), byte(seq & 0xff))
	//append data
	b = append(b, bytes.Repeat([]byte("test"), 1)...)
	//calculate sum
	sum := len(b) - 1
	var s uint32
	for i := 0; i < sum; i += 2 {
		s += uint32(b[i+1])<<8 | uint32(b[i])
	}
	if sum & 1 == 0 {
		s += uint32(b[sum])
	}
	s = s>>16 + s&0xffff
	s = s + s>>16
	b[2] ^= byte(^s & 0xff)
	b[3] ^= byte(^s >> 8)

	wn, we := conn.Write(b)
	if we != nil {
		return nil, we
	}
	if wn != len(b) {
		return nil, errors.New("write too short")
	}
	rb := make([]byte, 20+len(b))
	for {
		if _, err = conn.Read(rb); err != nil {
			return nil, err
		}
		rb = ipv4Payload(rb)
		break
	}
	return rb, nil
}

func main()  {
	flag.Parse()
	address := flag.Arg(0)
	if address == "" {
		fmt.Println("address can not be empty")
		return
	}
	b, err := ping(address)
	if err != nil {
		fmt.Println("err:", err)
		return
	}
	fmt.Println("ping success, receive bytes:", b)
}