package main

import (
	"encoding/binary"
	"fmt"
	"net"
)

type DNSHeader struct {
	ID      uint16
	QR      uint8
	OPCODE  uint8
	AA      uint8
	TC      uint8
	RD      uint8
	RA      uint8
	Z       uint8
	RCODE   uint8
	QDCOUNT uint16
	ANCOUNT uint16
	NSCOUNT uint16
	ARCOUNT uint16
}

func main() {
	fmt.Println("Logs from your program will appear here!")

	udpAddr, err := net.ResolveUDPAddr("udp", "127.0.0.1:2053")
	if err != nil {
		fmt.Println("Failed to resolve UDP address:", err)
		return
	}

	udpConn, err := net.ListenUDP("udp", udpAddr)
	if err != nil {
		fmt.Println("Failed to bind to address:", err)
		return
	}
	defer udpConn.Close()

	buf := make([]byte, 512)

	for {
		size, source, err := udpConn.ReadFromUDP(buf)
		if err != nil {
			fmt.Println("Error receiving data:", err)
			break
		}

		receivedData := string(buf[:size])
		fmt.Printf("Received %d bytes from %s: %s\n", size, source, receivedData)

		header := getHeader(1)
		response := getFormattedHeaderResponse(header)

		_, err = udpConn.WriteToUDP(response, source)
		if err != nil {
			fmt.Println("Failed to send response:", err)
		}
	}
}

func getHeader(isResponse uint8) *DNSHeader {
	header := DNSHeader{
		ID:      1234,
		QR:      isResponse,
		OPCODE:  0,
		AA:      0,
		TC:      0,
		RD:      0,
		RA:      0,
		Z:       0,
		RCODE:   0,
		QDCOUNT: 0,
		ANCOUNT: 0,
		NSCOUNT: 0,
		ARCOUNT: 0,
	}
	return &header
}

func getFormattedHeaderResponse(dnsHeader *DNSHeader) []byte {
	buf := make([]byte, 12)

	var secondByte byte = 0
	secondByte |= (dnsHeader.QR << 7)
	secondByte |= (dnsHeader.OPCODE << 3)
	secondByte |= (dnsHeader.AA << 2)
	secondByte |= (dnsHeader.TC << 1)
	secondByte |= (dnsHeader.RD)

	var thirdByte byte = 0
	thirdByte |= (dnsHeader.RA << 7)
	thirdByte |= (dnsHeader.Z << 4)
	thirdByte |= (dnsHeader.RCODE)

	binary.BigEndian.PutUint16(buf[:2], dnsHeader.ID)
	buf[2] = secondByte
	buf[3] = thirdByte
	binary.BigEndian.PutUint16(buf[4:6], dnsHeader.QDCOUNT)
	binary.BigEndian.PutUint16(buf[6:8], dnsHeader.ANCOUNT)
	binary.BigEndian.PutUint16(buf[8:10], dnsHeader.NSCOUNT)
	binary.BigEndian.PutUint16(buf[10:12], dnsHeader.ARCOUNT)

	return buf
}
