package datastore

import (
	"encoding/csv"
	"encoding/hex"
	"io"
	"strings"
)

// OUI Map
// Download oui.txt from
// http://standards-oui.ieee.org/oui/oui.txt
// をやめて
// https://maclookup.app/downloads/csv-database

// LoadOUIMap : Load OUI Data from io.ReadCloser
func loadOUIMap(f io.ReadCloser) {
	r := csv.NewReader(f)
	for {
		record, err := r.Read()
		if err == io.EOF {
			break
		}
		if len(record) < 2 {
			continue
		}
		oui := record[0]
		if !strings.Contains(oui, ":") {
			continue
		}
		oui = strings.TrimSpace(oui)
		oui = strings.ReplaceAll(oui, ":", "")
		ouiMap[oui] = record[1]
	}
}

// FindVendor : Find Vendor Name from MAC Address
func FindVendor(mac string) string {
	if mac == "" {
		return ""
	}
	mac = strings.TrimSpace(mac)
	mac = strings.ReplaceAll(mac, ":", "")
	mac = strings.ReplaceAll(mac, "-", "")
	if len(mac) > 6 {
		mac = strings.ToUpper(mac)
		if n, ok := ouiMap[mac[:6]]; ok {
			return n
		}
		if n, ok := ouiMap[mac[:7]]; ok {
			return n
		}
		if n, ok := ouiMap[mac[:9]]; ok {
			return n
		}
		if h, err := hex.DecodeString(mac); err == nil {
			if (h[0] & 0x02) == 0x02 {
				h[0] = h[0] & 0xfd
				mac = strings.ToUpper(hex.EncodeToString(h))
				if n, ok := ouiMap[mac[:6]]; ok {
					return n + "(Local)"
				}
				if n, ok := ouiMap[mac[:7]]; ok {
					return n + "(Local)"
				}
				if n, ok := ouiMap[mac[:9]]; ok {
					return n + "(Local)"
				}
				return "Local"
			}
		}
	}
	return "Unknown"
}
