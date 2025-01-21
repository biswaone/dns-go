package main

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"log"
	"math/rand"
	"net"
	"os"
	"strconv"
	"strings"
)

const TYPE_A uint16 = 1
const TYPE_NS uint16 = 2
const CLASS_IN uint16 = 1

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

type DNSRecord struct {
	Name  []byte
	Type  uint16
	Class uint16
	TTL   uint32
	Data  []byte
}

type DNSPacket struct {
	header      DNSHeader
	questions   []DNSQuestion
	answers     []DNSRecord
	authorities []DNSRecord
	additionals []DNSRecord
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
	//RECURSION_DESIRED := uint16(1 << 8)
	header := DNSHeader{
		ID:           id,
		NumQuestions: 1,
		Flags:        0,
	}
	question := DNSQuestion{
		Name:  []byte(name),
		Type:  recordType,
		Class: 1,
	}
	return append(HeaderToBytes(header), QuestionToBytes(question)...)
}

func decodeName(reader *bytes.Reader) []byte {
	var parts [][]byte
	for {
		length, err := reader.ReadByte()
		if err != nil {
			fmt.Println("Error reading length:", err)
			return nil
		}

		if length == 0 {
			break
		} else if length&0xC0 != 0 {
			pointerBytes := []byte{length & 0x3F, 0}
			_, err := reader.Read(pointerBytes[1:])
			if err != nil {
				fmt.Println("Error reading pointer bytes:", err)
				return nil
			}

			pointer := binary.BigEndian.Uint16(pointerBytes)
			currentPos, _ := reader.Seek(0, 1)
			reader.Seek(int64(pointer), 0)
			result := decodeName(reader)
			reader.Seek(currentPos, 0)

			parts = append(parts, result)
			break
		} else {
			part := make([]byte, length)
			_, err := reader.Read(part)
			if err != nil {
				fmt.Println("Error reading part:", err)
				return nil
			}

			parts = append(parts, part)
		}
	}

	return bytes.Join(parts, []byte("."))
}

func parseHeader(reader *bytes.Reader) (DNSHeader, error) {
	var header DNSHeader
	err := binary.Read(reader, binary.BigEndian, &header)
	if err != nil {
		return DNSHeader{}, err
	}
	return header, nil
}

func parseQuestion(reader *bytes.Reader) (DNSQuestion, error) {
	name := decodeName(reader)

	data := make([]byte, 4)
	_, err := reader.Read(data)
	if err != nil {
		return DNSQuestion{}, err
	}

	var question DNSQuestion
	question.Name = name
	question.Type = binary.BigEndian.Uint16(data[0:2])
	question.Class = binary.BigEndian.Uint16(data[2:4])

	return question, nil
}

func parseRecord(reader *bytes.Reader) (DNSRecord, error) {
	name := decodeName(reader)
	data := make([]byte, 10)
	_, err := reader.Read(data)
	if err != nil {
		return DNSRecord{}, err
	}
	var dnsRecord DNSRecord
	dnsRecord.Name = name
	dnsRecord.Type = binary.BigEndian.Uint16(data[:2])
	dnsRecord.Class = binary.BigEndian.Uint16(data[2:4])
	dnsRecord.TTL = binary.BigEndian.Uint32(data[4:8])

	dataLen := binary.BigEndian.Uint16(data[8:10])
	dnsRecord.Data = make([]byte, dataLen)

	if dnsRecord.Type == TYPE_NS {
		dnsRecord.Data = decodeName(reader)
	} else {
		_, err := reader.Read(dnsRecord.Data)
		if err != nil {
			return DNSRecord{}, err
		}
	}

	return dnsRecord, nil
}

func parseDNSPacket(reader *bytes.Reader) DNSPacket {
	header, _ := parseHeader(reader)
	questions := make([]DNSQuestion, header.NumQuestions)
	answers := make([]DNSRecord, header.NumAnswers)
	authorities := make([]DNSRecord, header.NumAuthorities)
	additionals := make([]DNSRecord, header.NumAdditionals)

	for i := range questions {
		questions[i], _ = parseQuestion(reader)
	}

	for i := range answers {
		answers[i], _ = parseRecord(reader)
	}

	for i := range authorities {
		authorities[i], _ = parseRecord(reader)
	}

	for i := range additionals {
		additionals[i], _ = parseRecord(reader)
	}

	return DNSPacket{
		header:      header,
		questions:   questions,
		answers:     answers,
		authorities: authorities,
		additionals: additionals,
	}
}

func ipToString(ip []byte) string {
	var ipStrings []string
	for _, hex := range ip {
		decimal := strconv.Itoa(int(hex))
		ipStrings = append(ipStrings, decimal)
	}
	return strings.Join(ipStrings, ".")
}

func sendQuery(ipAddress string, domainName string, recordType uint16) []byte {
	query := buildQuery(domainName, recordType)
	port := ":53"
	address := ipAddress + port

	// Create a UDP socket
	sock, err := net.Dial("udp", address)
	if err != nil {
		log.Fatal(err)
	}
	defer sock.Close()

	// Send the query to 8.8.8.8:53
	_, err = sock.Write([]byte(query))
	if err != nil {
		log.Fatal(err)
	}

	// Read the response
	response := make([]byte, 1024)
	_, err = sock.Read(response)
	if err != nil {
		log.Fatal(err)
	}

	return response
}

func getAnswer(packet DNSPacket) string {
	for _, answer := range packet.answers {
		if answer.Type == TYPE_A {
			return ipToString(answer.Data)
		}
	}
	return ""
}

func getNameServerIP(packet DNSPacket) string {
	for _, authority := range packet.additionals {
		if authority.Type == TYPE_A {
			return ipToString(authority.Data)
		}
	}
	return ""
}

func getNameServer(packet DNSPacket) string {
	for _, authority := range packet.authorities {
		if authority.Type == TYPE_NS {
			return string(authority.Data[:])
		}
	}
	return ""
}

func resolve(domainName string, recordType uint16) string {
	nameServer := "198.41.0.4"
	for {
		fmt.Printf("Querying %s for %s\n", nameServer, domainName)
		response := sendQuery(nameServer, domainName, recordType)
		reader := bytes.NewReader(response)
		packet := parseDNSPacket(reader)

		if recordType == TYPE_NS {
			for _, answer := range packet.answers {
				if answer.Type == TYPE_NS {
					return string(answer.Data[:])
				}
			}
		}

		ip := getAnswer(packet)
		if ip != "" {
			return ip
		} else if nsIP := getNameServerIP(packet); nsIP != "" {
			nameServer = nsIP
		} else if nsDomain := getNameServer(packet); nsDomain != "" {
			nameServer = resolve(nsDomain, TYPE_A)
		} else {
			log.Fatal("something went wrong")
		}

	}
}

func parseRecordType(recordTypeStr string) (uint16, error) {
	switch strings.ToUpper(recordTypeStr) {
	case "A":
		return TYPE_A, nil
	case "NS":
		return TYPE_NS, nil
	default:
		return 0, fmt.Errorf("unsupported record type: %s", recordTypeStr)
	}
}

func main() {
	if len(os.Args) < 3 {
		fmt.Println("Usage: program <record_type> <domain_name>")
		fmt.Println("Record types: A, NS")
		os.Exit(1)
	}

	domainName := os.Args[2]
	recordTypeStr := os.Args[1]

	recordType, err := parseRecordType(recordTypeStr)
	if err != nil {
		log.Fatal(err)
	}

	switch recordType {
	case TYPE_A:
		ip := resolve(domainName, TYPE_A)
		fmt.Printf("The IP of %s is %s\n", domainName, ip)
	case TYPE_NS:
		ns := resolve(domainName, TYPE_NS)
		fmt.Printf("The nameserver for %s is %s\n", domainName, ns)
	}

}
