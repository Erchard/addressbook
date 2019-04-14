package book

import (
	"bufio"
	"github.com/pkg/errors"
	"github.com/syndtr/goleveldb/leveldb"
	"log"
	"net"
)

func Start() {
	db, err := leveldb.OpenFile("path/to/db", nil)
	if err != nil {
		log.Fatal(err)
	}
	err = db.Put([]byte("key"), []byte("value"), nil)

	data, err := db.Get([]byte("key"), nil)

	println(string(data))

}

func Open(addr string) (*bufio.ReadWriter, error) {
	log.Println("Dial " + addr)
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		return nil, errors.Wrap(err, "Dialing "+addr+" failed")
	}
	return bufio.NewReadWriter(bufio.NewReader(conn), bufio.NewWriter(conn)), nil
}
