package datastore

import (
	"encoding/csv"
	"io"
	"strings"
)

func (ds *DataStore) loadTLSCihperNameMap(f io.ReadCloser) {
	if f == nil {
		return
	}
	defer f.Close()
	reader := csv.NewReader(f)
	for {
		line, err := reader.Read()
		if err != nil {
			break
		}
		if len(line) < 2 {
			continue
		}
		id := strings.Replace(line[0], ",", "", 1)
		id = strings.Replace(id, "0x", "", 2)
		id = strings.ToLower(id)
		name := line[1]
		if strings.HasPrefix(name, "TLS_") {
			ds.tlsCSMap[id] = name
		}
	}
}

func (ds *DataStore) GetCipherSuiteName(id string) (string, bool) {
	r, ok := ds.tlsCSMap[id]
	return r, ok
}
