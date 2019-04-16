package book

import (
	"bufio"
	"encoding/binary"
	"github.com/pkg/errors"
	"github.com/syndtr/goleveldb/leveldb"
	"log"
	"math"
	"net"
	"strconv"
)

const DB_PATH = "path/to/db"

var db *leveldb.DB

type NodeStatus struct {
	Address *string
	Status  *uint64
	Data    []byte
}

func init() {
	database, err := leveldb.OpenFile(DB_PATH, nil)
	if err != nil {
		log.Fatal(err)
	} else {
		db = database
		log.Println("Database " + DB_PATH + " connected")
	}
}

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
func Update(nodestatus *NodeStatus) {
	host, port, err := net.SplitHostPort(*nodestatus.Address)
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

	var timestamp uint64
	if nodestatus.Status != nil {
		timestamp = *nodestatus.Status
	} else {
		timestamp = uint64(math.MaxUint64)

	}
	binary.BigEndian.PutUint64(value, timestamp)

	err = db.Put(key, value, nil)
	nodestatus.Data = append(key, value...)
}

func restore(data []byte) *NodeStatus {

	nodestatus := NodeStatus{
		Data: data,
	}

	host := net.IP(data[:16]).String()
	port := binary.BigEndian.Uint16(data[16:18])
	address := host + ":" + strconv.Itoa(int(port))

	nodestatus.Address = &address

	timestamp := binary.BigEndian.Uint64(data[18:26])

	if timestamp != math.MaxUint64 {
		nodestatus.Status = &timestamp
	}

	return &nodestatus
}

//func GetAll() []NodeStatus {
//
//}

func Open(addr string) (*bufio.ReadWriter, error) {
	log.Println("Dial " + addr)
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		return nil, errors.Wrap(err, "Dialing "+addr+" failed")
	}
	return bufio.NewReadWriter(bufio.NewReader(conn), bufio.NewWriter(conn)), nil
}
