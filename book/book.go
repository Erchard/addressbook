package book

import (
	"../configuration"
	"bytes"
	"encoding/binary"
	"fmt"
	"github.com/pkg/errors"
	"github.com/syndtr/goleveldb/leveldb"
	"log"
	"math"
	"net"
	"strconv"
	"time"
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
	update(&seedstatus)
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
func update(nodestatus *NodeStatus) error {

	host, port, err := net.SplitHostPort(*nodestatus.Address)
	if err != nil {
		return errors.Wrapf(err, "Error: %s Split address %s\n", err, *nodestatus.Address)
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

	data, err := db.Get(key, nil)
	if err != nil {
		log.Println(err)
	}

	err = db.Put(key, value, nil)
	if err != nil {
		log.Println(err)
	}
	nodestatus.Data = append(key, value...)

	if bytes.Compare(data, value) != 0 {
		sendToAllOnline(nodestatus.Data)
	}
	return nil
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
		port = ":" + *configuration.Config.PreferredPort
	} else {
		port = ":0"
	}

	listener, err := net.Listen("tcp", port)
	if err != nil {
		return errors.Wrapf(err, "Unable to listen on port %s\n", port)
	}

	host, port, err := net.SplitHostPort(listener.Addr().String())
	if err != nil {
		log.Fatal(err)
	}
	configuration.Config.PreferredPort = &port
	log.Println("Host: ", host)
	log.Println("Listen on", listener.Addr().String())
	for {
		log.Println("Accept a connection request.")
		conn, err := listener.Accept()
		if err != nil {
			log.Println("Failed accepting a connection request:", err)
			continue
		}
		log.Println("Handle incoming messages.")
		go handleConnection(conn)
		fmt.Println(conn.RemoteAddr())
	}
}

func handleConnection(conn net.Conn) {
	address := conn.RemoteAddr().String()
	nodestatus := NodeStatus{
		Address: &address,
	}
	err := update(&nodestatus)
	if err != nil {
		log.Println(err)
	}
	var updateinfo = make([]byte, 26)

	err = update(restore(updateinfo))
	if err != nil {
		log.Println(err)
	}
	err = conn.Close()
	if err != nil {
		log.Println(err)
	}
}

func sendToAllOnline(msg []byte) {

	if len(msg) != 26 {
		log.Println("len(msg) = ", len(msg))
		return
	}

	for _, node := range GetAll() {
		if *node.Status == math.MaxUint64 {
			conn, err := net.Dial("tcp", *node.Address)
			if err != nil {
				log.Println(err, "Dialing "+*node.Address+" failed")
				timedisconnect := uint64(time.Now().Unix())
				node.Status = &timedisconnect
			}
			sendedbytes, err := conn.Write(msg)
			if sendedbytes != len(msg) {
				log.Println("sendedbytes != len(msg)")
				timedisconnect := uint64(time.Now().Unix())
				node.Status = &timedisconnect
			}
			if err != nil {
				log.Println(err)
				timedisconnect := uint64(time.Now().Unix())
				node.Status = &timedisconnect
			}
			if *node.Status != math.MaxUint64 {
				err = update(&node)
				if err != nil {
					log.Println(err)
				}
			}
			conn.Close()
		}
	}
}
