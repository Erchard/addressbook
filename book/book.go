package book

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"github.com/pkg/errors"
	"github.com/syndtr/goleveldb/leveldb"
	"log"
	"net"
	"strconv"
)

/*
func Start() {
	db, err := leveldb.OpenFile("path/to/db", nil)
	if err != nil {
		log.Fatal(err)
	}

	data, err := db.Get([]byte(0x00000000000000000000ffffc0a800660670), nil)

	println(string(data))
}
*/
func Update(nodeaddress *string, status *uint64) {
	host, port, err := net.SplitHostPort(*nodeaddress)
	if err != nil {
		log.Fatal(err)
	}

	ip := net.ParseIP(host)
	byteHost := []byte(ip)

	u, err := strconv.ParseInt(port, 10, 16)
	bytePort := make([]byte, 2)
	binary.BigEndian.PutUint16(bytePort, uint16(u))

	key := append(byteHost, bytePort...)

	value := make([]byte, 8)
	if status != nil {
		binary.BigEndian.PutUint64(value, *status)
	} else {
		value = []byte{255, 255, 255, 255, 255, 255, 255, 255}
	}

	db, err := leveldb.OpenFile("path/to/db", nil)
	if err != nil {
		log.Fatal(err)
	}
	err = db.Put(key, value, nil)

	fmt.Printf("key: %x value: %x \n", key, value)

}

func Open(addr string) (*bufio.ReadWriter, error) {
	log.Println("Dial " + addr)
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		return nil, errors.Wrap(err, "Dialing "+addr+" failed")
	}
	return bufio.NewReadWriter(bufio.NewReader(conn), bufio.NewWriter(conn)), nil
}
