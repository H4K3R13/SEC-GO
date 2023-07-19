package main

import (
	//"bytes"
	"encoding/hex"
	"fmt"
	"log"
	"net"
	"os"
	"strings"
	"time"
	//"strconv"
	"github.com/google/gopacket"
	"github.com/google/gopacket/pcap"
	//FOR TUI
	"github.com/pterm/pterm"
	"github.com/pterm/pterm/putils"
)

var (
	iface   = "en0"
	snaplen = int32(320)
	promisc = true
	timeout = pcap.BlockForever
	//filter   = "tcp[13] == 0x11 or tcp[13] == 0x10 or tcp[13] == 0x18"
	filter   = "tcp"
	devFound = false
	results  = make(map[string]int)
)

func capture(iface, target string) {
	handle, err := pcap.OpenLive(iface, snaplen, promisc, timeout)
	if err != nil {
		log.Panicln(err)
	}
	defer handle.Close()
	if err := handle.SetBPFFilter(filter); err != nil {
		log.Panicln(err)
	}
	source := gopacket.NewPacketSource(handle, handle.LinkType())
	fmt.Println("Capturing packets")
	for packet := range source.Packets() {
		networkLayer := packet.NetworkLayer()
		if networkLayer == nil {
			continue
		}
		transportLayer := packet.TransportLayer()
		if transportLayer == nil {
			continue
		}
		srcHost := networkLayer.NetworkFlow().Src().String()
		srcPort := transportLayer.TransportFlow().Src().String()
		if srcHost != target {
			continue
		}
		results[srcPort] += 1

		pterm.DefaultBasicText.Println(packet)

		// Get the application layer (payload) of the packet
		appLayer := packet.ApplicationLayer()
		if appLayer != nil {
			pterm.DefaultBasicText.Println("Application Layer/Payload:")
			pterm.DefaultBasicText.Println(appLayer.Payload())
		}

		// Get the packet data in hex dump format
		pterm.DefaultBasicText.Println("Packet Data (Hex Dump):")
		pterm.DefaultBasicText.Println(pterm.Gray(hex.Dump(packet.Data())))

	}
}


func getServiceName(port string) string {
	serviceNames := map[string]string{
		"80":   "HTTP",
		"443":  "HTTPS",
		"8080": "HTTP Proxy",
		// Add more mappings as needed
	}

	service, found := serviceNames[port]
	if found {
		return service
	}
	return ""
}


func main() {

	_ = pterm.DefaultBigText.WithLetters(putils.LettersFromString("SEC-GO")).Render()
	pterm.DefaultCenter.Println(("Develped By @H4K3R (Github)"))
	if len(os.Args) != 4 {
		log.Fatalln("Usage: main.go <capture_iface> <target_ip> <port1,port2,port3>")
	}
	pterm.DefaultCenter.Print("Scanning", os.Args[2])

	devices, err := pcap.FindAllDevs()
	if err != nil {
		log.Panicln(err)
	}
	iface := os.Args[1]
	for _, device := range devices {
		if device.Name == iface {
			devFound = true
		}
	}
	if !devFound {
		log.Panicf("Device named '%s' does not exist\n", iface)
	}
	if devFound == true {
		log.Printf("Device Found '%s", iface)
	}

	ip := os.Args[2]
	go capture(iface, ip)
	time.Sleep(1 * time.Second)

	ports := strings.Split(os.Args[3], ",")
	fmt.Println(ports)
	totalSteps := len(ports)
	progressbar, _ := pterm.DefaultProgressbar.WithTotal(totalSteps).Start()
	for _, port := range ports {
		progressbar.Increment()
		target := fmt.Sprintf("%s:%s", ip, port)
		pterm.DefaultBasicText.Println(pterm.Red("\nTrying: ", target))
		c, err := net.DialTimeout("tcp", target, 1000*time.Millisecond)
		if err != nil {
			continue
		}
		c.Close()
	}

	time.Sleep(2 * time.Second)
	for port, confidence := range results {
		if confidence >= 1 {
			serviceName := getServiceName(port)
			fmt.Printf("Port %s open (confidence: %d)\n  Servivce : %s \n", port, confidence, serviceName)
		}
	}
}
