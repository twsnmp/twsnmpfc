package wol

import (
	"net"
)

// SendWakeOnLanPacket : send wake on lan magic packet to mac address
func SendWakeOnLanPacket(mac string) error {
	ra, err := net.ResolveUDPAddr("udp4", "255.255.255.255:9")
	if err != nil {
		return err
	}
	la, err := net.ResolveUDPAddr("udp4", ":0")
	if err != nil {
		return err
	}
	c, err := net.DialUDP("udp4", la, ra)
	if err != nil {
		return err
	}
	defer c.Close()
	hw, err := net.ParseMAC(mac)
	if err != nil {
		return err
	}
	packet := []byte{}
	prefix := []byte{0xff, 0xff, 0xff, 0xff, 0xff, 0xff}
	packet = append(packet, prefix...)
	for i := 0; i < 16; i++ {
		packet = append(packet, hw...)
	}
	_, err = c.Write(packet)
	return err
}
