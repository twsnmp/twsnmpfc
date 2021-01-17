package datastore

import (
	"bufio"
	"io"
	"strings"
)

// OUI Map
// Download oui.txt from
// http://standards-oui.ieee.org/oui/oui.txt

// LoadOUIMap : Load OUI Data from io.ReadCloser
func (ds *DataStore) loadOUIMap(f io.ReadCloser) {
	s := bufio.NewScanner(f)
	for s.Scan() {
		l := strings.TrimSpace(s.Text())
		if len(l) < 1 {
			continue
		}
		f := strings.Fields(l)
		if len(f) < 4 || f[1] != "(base" {
			continue
		}
		ds.ouiMap[f[0]] = strings.Join(f[3:], " ")
	}
}

// FindVendor : Find Vendor Name from MAC Address
func (ds *DataStore) FindVendor(mac string) string {
	mac = strings.TrimSpace(mac)
	mac = strings.ReplaceAll(mac, ":", "")
	mac = strings.ReplaceAll(mac, "-", "")
	if len(mac) > 6 {
		mac = strings.ToUpper(mac)
		if n, ok := ds.ouiMap[mac[:6]]; ok {
			return n
		}
	}
	return "Unknown"
}
