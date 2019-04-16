package book

import (
	"../configuration"
	"encoding/binary"
	"fmt"
	"github.com/pkg/errors"
	"github.com/syndtr/goleveldb/leveldb"
	"log"
	"math"
	"net"
	"strconv"
)

var db *leveldb.DB

type NodeStatus struct {
	Address *string
	Status  *uint64
	Data    []byte
}

func init() {
	database, err := leveldb.OpenFile(configuration.Config.DbPath, nil)
	//defer db.Close()
	if err != nil {
		log.Fatal(err)
	} else {
		db = database
		log.Println("Database " + configuration.Config.DbPath + " connected")
	}

	seedstatus := NodeStatus{
		Address: &configuration.Config.Seed[0],
	}
	Update(&seedstatus)
	err = server()
	if err != nil {
		log.Fatalln(err)
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

func GetAll() []NodeStatus {

	nodeArray := make([]NodeStatus, 0)
	iter := db.NewIterator(nil, nil)
	for iter.Next() {
		nodeArray = append(
			nodeArray,
			*restore(
				append(
					iter.Key(),
					iter.Value()...)))
	}
	fmt.Println(nodeArray)
	return nodeArray
}

func server() error {
	var err error
	var port string
	if configuration.Config.PreferredPort != nil {
		port = ":" + strconv.Itoa(int(*configuration.Config.PreferredPort))
	} else {
		port = ":11111"
	}

	listener, err := net.Listen("tcp", port)
	if err != nil {
		return errors.Wrapf(err, "Unable to listen on port %s\n", port)
	}
	log.Println("Listen on", listener.Addr().String())
	for {
		log.Println("Accept a connection request.")
		conn, err := listener.Accept()
		if err != nil {
			log.Println("Failed accepting a connection request:", err)
			continue
		}
		log.Println("Handle incoming messages.")
		//go e.handleMessages(conn)
		fmt.Println(conn.RemoteAddr())
	}
}
