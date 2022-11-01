package main

import (
	"bufio"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"log"
	"math"
	"net"
	"strconv"
	"sync"
	"sync/atomic"
	"time"
)

type websockets_table struct {
	table map[string]chan []byte
	sync.RWMutex
}

//type websockets_table_shards []*websockets_table

var mock_users *websockets_table = &websockets_table{
	table: make(map[string]chan []byte),
}

type router_conn_obj struct {
	raddr *net.TCPAddr
	conn  *net.TCPConn
}

const sock_id = "SOCK_test"
const CONN_GROUP = "AZ1_main"
const BUFFER_SIZE = 1000000 //1 mb
var router_addr []string = []string{"127.0.0.1:10000"}
var HEARTBEAT_INTERVAL = 3 * time.Second
var HEARTBEAT_MAX_DELAY = HEARTBEAT_INTERVAL

//var HEARTBEAT_MAX_DELAY = 500 * time.Millisecond
var RETRY_INTERVAL = 1 * time.Second

func main() {

	var recover_bool_input string

	fmt.Println("'2' for new socket, '3' for recover socket")
	fmt.Scanln(&recover_bool_input)

	switch recover_bool_input {
	case "2":
		for i := 0; i < len(router_addr); i++ {
			connect_new_socket(router_addr[i], "nil")
		}
	case "3":
		for i := 0; i < len(router_addr); i++ {
			raddr, _ := net.ResolveTCPAddr("tcp", router_addr[i])
			recover_socket(raddr)
		}
	}
}

func connect_new_socket(addr string, old_conn_id string) {

	raddr, _ := net.ResolveTCPAddr("tcp", addr)
	conn, err := net.DialTCP("tcp", nil, raddr)

	if err != nil {
		handle_panic_err(conn, err)
	}

	defer conn.Close()

	_, err = conn.Write(package_op(2, []string{sock_id, CONN_GROUP, old_conn_id}, nil))
	if err != nil {
		handle_panic_err(conn, err)
	}

	res_b := make([]byte, 6)
	_, err = conn.Read(res_b)
	if err != nil {
		handle_panic_err(conn, err)
	}

	if res_b[0] != 2 {
		handle_panic_err(conn, err)
	}

	if string(res_b[5]) == "0" {
		handle_panic_err(conn, errors.New("sock_id already exists, choose different one"))
	}

	router_conn := &router_conn_obj{
		raddr: raddr,
		conn:  conn,
	}

	read_tcp(router_conn)

}

func recover_socket(raddr *net.TCPAddr) {

	defer func() {
		time.Sleep(RETRY_INTERVAL)
		log_err(raddr.String() + " recovering " + sock_id + "...")
		recover_socket(raddr)
	}()

	conn, err := net.DialTCP("tcp", nil, raddr)
	if err != nil {
		log_err("recovering " + sock_id + ": " + err.Error())
		return
	}

	defer conn.Close()

	_, err = conn.Write(package_op(3, []string{sock_id}, nil))
	if err != nil {
		log_err("recovering " + sock_id + ": " + err.Error())
		return
	}

	res_b := make([]byte, 6)
	_, err = conn.Read(res_b)
	if err != nil {
		log_err("recovering " + sock_id + ": " + err.Error())
		return
	}

	if res_b[0] != 3 {
		log_err("recovering " + sock_id + ": " + "unexpected op: " + string(res_b[0]))
		return
	}

	if string(res_b[5]) == "0" {
		log_err("recovering " + sock_id + ": " + "recovery failed")
		return
	}

	router_conn := &router_conn_obj{
		raddr: raddr,
		conn:  conn,
	}

	read_tcp(router_conn)
}

func read_tcp(router_conn *router_conn_obj) {

	//SOCK
	var received uint32

	heartbeat_chan := make(chan []byte, 3)
	timeout_chan := make(chan byte)
	close_chan := make(chan struct{})

	defer func() {
		close_chan <- struct{}{}
	}()

	go heartbeat_check(router_conn, heartbeat_chan, timeout_chan, close_chan)
	go heartbeat_worker(router_conn, &received, timeout_chan)
	//go mock_writer(router_conn)
	go benchmark_simulation(router_conn)

	reader := bufio.NewReaderSize(router_conn.conn, BUFFER_SIZE)

	var (
		err     error
		header  []byte = make([]byte, 5)
		size    uint32
		op      byte
		b       []byte
		param   []string
		payload []byte
	)
	for {

		_, err = io.ReadFull(reader, header)
		if err != nil {
			handle_conn_err(router_conn, err)
			break
		}

		op = header[0]

		size = binary.BigEndian.Uint32(header[1:])
		b = nil

		if size > 0 {
			b = make([]byte, size)
			_, err = io.ReadFull(reader, b)
			if err != nil {
				handle_conn_err(router_conn, err)
				break
			}
		}

		fmt.Println("RECV:", header, b)

		if op == 1 {
			heartbeat_chan <- b

			//SOCK
			atomic.StoreUint32(&received, 0)
		} else {
			switch op {
			case 50:
				param, payload, err = byte_to_params_and_payload(b, 1, true)
				if err != nil {
					break
				}

				//fmt.Println("op: 50", param[0], payload[0], string(payload[1:]))

				atomic.AddUint32(&received, 1)

				switch payload[0] {
				case 20:
					param, _, err = byte_to_params_and_payload(payload[1:], 1, false)
					fmt.Println("connect_session res:", param[0])
				case 21:
					param, _, err = byte_to_params_and_payload(payload[1:], 1, false)
					fmt.Println("recover_session res:", param[0])
				case 30:
					param, _, err = byte_to_params_and_payload(payload[1:], 2, false)
					fmt.Println(param)
					router_conn.conn.Write(package_op(30, []string{param[0], param[1], "1"}, []byte("NEW_STATE_PAYLOAD")))
				case 31:
					param, payload, err = byte_to_params_and_payload(payload[1:], 1, true)
					fmt.Println(param[0], string(payload))
				case 100:
					_, payload, err = byte_to_params_and_payload(payload[1:], 0, true)
					//fmt.Println("Payload to", param[0], string(payload))
					// num, _ := strconv.ParseUint(string(payload), 10, 64)
					// temp_i+=int64(num)
					// temp_total--
					// temp_slice[num] = true
					// if temp_total <= 10 {
					// 	fmt.Println(temp_i)
					// 	for i := 0; i < len(temp_slice); i++ {
					// 		if !temp_slice[i] {
					// 			fmt.Println("Missing:", i)
					// 		}
					// 	}
					// }
					temp_count--
					if temp_count == 0 {
						fmt.Println("ALL RECEIVED:", (time.Now().UnixNano()-temp_start_time)/temp_total, "ns/op")
					}
				}
			}

			if err != nil {
				log_err(err.Error())
			}
		}

	}
}

const temp_total = math.MaxUint16 * 10

var temp_count int64 = temp_total
var temp_start_time int64

// var temp_i int64 = 0
// var temp_slice []bool = make([]bool, math.MaxUint16)

func heartbeat_worker(router_conn *router_conn_obj, received *uint32, timeout_chan chan byte) {

	var (
		heartbeat_ver                byte  = 1
		expected_heartbeat_unix_nano int64 = 0
	)

	//API
	// write_b := make([]byte, 14)
	// copy(write_b, []byte{1, 0, 0, 0, 9})

	//SOCKET
	write_b := make([]byte, 18)
	copy(write_b, []byte{1, 0, 0, 0, 13})

	for {

		expected_heartbeat_unix_nano = time.Now().UnixNano() + int64(HEARTBEAT_INTERVAL)

		write_b[5] = heartbeat_ver
		binary.BigEndian.PutUint64(write_b[6:14], uint64(expected_heartbeat_unix_nano))
		// SOCKET
		binary.BigEndian.PutUint32(write_b[14:18], atomic.LoadUint32(received))

		//fmt.Println("Heartbeat write:", write_b)

		_, err := router_conn.conn.Write(write_b)
		if err != nil {
			handle_conn_err(router_conn, err)
		}

		go func(ver byte) {
			time.Sleep(HEARTBEAT_MAX_DELAY)
			timeout_chan <- ver
		}(heartbeat_ver)

		time.Sleep(time.Duration(expected_heartbeat_unix_nano - time.Now().UnixNano()))
		heartbeat_ver++

	}

}

func heartbeat_check(router_conn *router_conn_obj, heartbeat_chan chan []byte, timeout_chan chan byte, close_chan chan struct{}) {

	var (
		heartbeat_ver byte = 1
		ver           byte
		b             []byte
	)
	for {
		select {
		case b = <-heartbeat_chan:
			//fmt.Println("Heartbeat:", ver)
			if b[0] != heartbeat_ver {
				handle_conn_err(router_conn, errors.New("heartbeat version mismatch. Incoming ver: "+fmt.Sprint(ver)+". Heartbeat ver: "+fmt.Sprint(heartbeat_ver)))
				return
			}
			heartbeat_ver++

			//API
			//fmt.Println("Amount received (API ONLY)", binary.BigEndian.Uint32(b[1:5]))

		case ver = <-timeout_chan:
			if ver == heartbeat_ver {
				handle_conn_err(router_conn, errors.New("heartbeat timeout. Timeout ver: "+fmt.Sprint(ver)+". Heartbeat ver: "+fmt.Sprint(heartbeat_ver)))
				return
			}
		case <-close_chan:
			return
		}
	}

}

func handle_panic_err(conn *net.TCPConn, err error) {
	conn.Close()
	log.Panicln("ERROR", err.Error())
}

func handle_conn_err(router_conn *router_conn_obj, err error) {
	router_conn.conn.Close()
	log_err("Conn err: " + err.Error())
	log_err(router_conn.raddr.String() + " closed, recovering...")

	panic("BLOCKED RESTART")
	//recover_socket(router_conn.raddr)
}

func log_err(err string) {
	fmt.Println(err)
	//LOG
}

func mock_writer(router_conn *router_conn_obj) {

	i := 0

	var (
		scanned string
		op      uint64
	)
	for {
		fmt.Println("Op: 20 or 21 or 100")
		fmt.Scanln(&scanned)

		op, _ = strconv.ParseUint(scanned, 10, 8)

		switch op {
		case 20:
			router_conn.conn.Write(package_op(20, []string{"uid1", "sid" + fmt.Sprint(i), "wid" + fmt.Sprint(i)}, nil))
			router_conn.conn.Write(package_op(20, []string{"uid2", "sid" + fmt.Sprint(i), "wid" + fmt.Sprint(i)}, nil))
			i++
		case 21:
			router_conn.conn.Write(package_op(21, []string{"uid2", "sid" + fmt.Sprint(i-1), "new_wid" + fmt.Sprint(i), "wid" + fmt.Sprint(i-1)}, nil))
		case 100:
			router_conn.conn.Write(package_op(100, []string{"sid0", "uid1", "uid2"}, []byte("100_PAYLOAD")))
		}

	}

}

func benchmark_simulation(router_conn *router_conn_obj) {

	var i int64
	var packaged packaged_op
	temp_start_time = time.Now().UnixNano()

	router_conn.conn.Write(package_op(20, []string{"uid1", "sid1", "wid1"}, nil))
	router_conn.conn.Write(package_op(20, []string{"uid2", "sid2", "wid2"}, nil))

	for i = 0; i < temp_total; i++ {
		if i%50 == 0 {
			time.Sleep(time.Microsecond)
		}
		if i%(temp_total/100) == 0 {
			fmt.Println("Progress:", math.Trunc((float64(i) / float64(temp_total) * 100)), "%")
		}
		packaged = package_op(100, []string{"sid1", "uid1", "uid2"}, []byte(strconv.FormatInt(i, 10)))
		n, err := router_conn.conn.Write(packaged)
		if err != nil || n == 0 {
			fmt.Println("ERROR AT:", i)
			panic("simulation error:" + err.Error())
		}
	}

	fmt.Println("Benchmark:", (time.Now().UnixNano()-temp_start_time)/temp_total, "ns/op")
}
