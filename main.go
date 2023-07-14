package main

import (
	"encoding/binary"
	"fmt"
	"math/rand"
	"net"
	"strings"
)

var TYPE_A uint16 = 1
var CLASS_IN uint16 = 1

type DNSHeader struct {
	ID             uint16
	Flags          uint16
	NumQuestions   uint16
	NumAnswers     uint16
	NumAuthorities uint16
	NumAdditionals uint16
}

type DNSQuestion struct {
	Name  []byte
	Type  uint16
	Class uint16
}

func HeaderToBytes(header DNSHeader) []byte {
	buffer := make([]byte, 12)
	binary.BigEndian.PutUint16(buffer[0:2], header.ID)
	binary.BigEndian.PutUint16(buffer[2:4], header.Flags)
	binary.BigEndian.PutUint16(buffer[4:6], header.NumQuestions)
	binary.BigEndian.PutUint16(buffer[6:8], header.NumAnswers)
	binary.BigEndian.PutUint16(buffer[8:10], header.NumAuthorities)
	binary.BigEndian.PutUint16(buffer[10:12], header.NumAdditionals)
	return buffer
}

func QuestionToBytes(question DNSQuestion) []byte {
	nameLen := len(question.Name)
	buffer := make([]byte, nameLen+4)
	copy(buffer[0:nameLen], question.Name)
	binary.BigEndian.PutUint16(buffer[nameLen:nameLen+2], question.Type)
	binary.BigEndian.PutUint16(buffer[nameLen+2:nameLen+4], question.Class)
	return buffer
}

func encodeDNSName(domainName string) []byte {
	var encoded []byte
	parts := strings.Split(domainName, ".")
	for _, part := range parts {
		encoded = append(encoded, byte(len(part)))
		encoded = append(encoded, []byte(part)...)
	}
	encoded = append(encoded, byte(0))
	return encoded
}

func buildQuery(domainName string, recordType uint16) []byte {
	r := rand.New(rand.NewSource(1))
	name := encodeDNSName(domainName)
	id := uint16(r.Intn(65536))
	RECURSION_DESIRED := uint16(1 << 8)
	header := DNSHeader{
		ID:           id,
		NumQuestions: 1,
		Flags:        RECURSION_DESIRED,
	}
	question := DNSQuestion{
		Name:  []byte(name),
		Type:  recordType,
		Class: 1,
	}
	return append(HeaderToBytes(header), QuestionToBytes(question)...)
}

func main() {

	query := buildQuery("example.com", TYPE_A)
	fmt.Printf("%#v\n", query)

	// Create a UDP socket
	sock, err := net.Dial("udp", "8.8.8.8:53")
	if err != nil {
		fmt.Println("Error creating socket:", err)
		return
	}
	defer sock.Close()

	// Send the query to 8.8.8.8:53
	_, err = sock.Write([]byte(query))
	if err != nil {
		fmt.Println("Error sending query:", err)
		return
	}
}
