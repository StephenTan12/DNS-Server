package main

import (
	"encoding/binary"
	"fmt"
	"net"
)

type DomainName struct {
	prefix string
	domain string
	tld    string
}

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

type DNSQuestion struct {
	domainName DomainName
	typeRec    uint16
	classField uint16
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
		fmt.Printf("%d\n", int(receivedData[20:21][0]))

		header := getHeader(buf, 1, 1)
		question := getQuestion(receivedData)
		formattedHeader := getFormattedHeaderResponse(header)
		formattedQuestions := getFormattedQuestionResponse(question)

		fmt.Printf("%v\n", formattedQuestions)

		response := append(formattedHeader, formattedQuestions...)

		_, err = udpConn.WriteToUDP(response, source)
		if err != nil {
			fmt.Println("Failed to send response:", err)
		}
	}
}

func getHeader(buf []byte, isResponse uint8, numQuestion uint16) *DNSHeader {
	header := DNSHeader{
		ID:      binary.BigEndian.Uint16(buf[:2]),
		QR:      isResponse,
		OPCODE:  0,
		AA:      0,
		TC:      0,
		RD:      0,
		RA:      0,
		Z:       0,
		RCODE:   0,
		QDCOUNT: numQuestion,
		ANCOUNT: 0,
		NSCOUNT: 0,
		ARCOUNT: 0,
	}
	return &header
}

func getQuestion(data string) *DNSQuestion {
	question := DNSQuestion{
		domainName: getDomainName(data),
		typeRec:    1,
		classField: 1,
	}
	return &question
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

func getFormattedQuestionResponse(dnsQuestion *DNSQuestion) []byte {
	domainName := dnsQuestion.domainName
	sizeLabels := len(domainName.domain) + len(domainName.tld) + 3
	buf := make([]byte, sizeLabels+4)

	buf[0] = uint8(len(domainName.domain))

	domainPointer := 0
	for {
		if domainPointer >= len(domainName.domain) {
			break
		}
		buf[domainPointer+1] = uint8(domainName.domain[domainPointer])
		domainPointer += 1
	}

	fmt.Printf("%v\n", buf)

	buf[domainPointer+1] = uint8(len(domainName.tld))

	tldPointerStart := domainPointer + 2
	tldPointerEnd := 0

	fmt.Printf("%v\n", buf)

	for {
		if tldPointerEnd >= len(domainName.tld) {
			break
		}
		buf[tldPointerStart+tldPointerEnd] = uint8(domainName.tld[tldPointerEnd])
		tldPointerEnd += 1
	}

	fmt.Printf("%v\n", buf)

	binary.BigEndian.PutUint16(buf[sizeLabels:sizeLabels+2], dnsQuestion.typeRec)
	binary.BigEndian.PutUint16(buf[sizeLabels+2:sizeLabels+4], dnsQuestion.classField)
	return buf
}

func getDomainName(data string) DomainName {
	START := 13
	domainPointerEnd := START
	for {
		if int(data[domainPointerEnd]) == 3 {
			break
		}
		domainPointerEnd += 1
	}

	domain := data[START:domainPointerEnd]

	START = domainPointerEnd + 1
	tldPointerEnd := START
	for {
		if int(data[tldPointerEnd]) == 0 {
			break
		}

		tldPointerEnd += 1
	}

	tld := data[START:tldPointerEnd]

	return DomainName{
		domain: domain,
		tld:    tld,
	}
}
