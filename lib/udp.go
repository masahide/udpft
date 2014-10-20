package lib

import (
	"bufio"
	"encoding/binary"
	"log"
	"os"
)

func SendWorker(to string, queue <-chan DataQueue, result chan<- error) {
	log.Printf("start to:%v", to)
	var err error
	f, _ := os.Create(to)
	conn := bufio.NewWriter(f)
	nn := 0
	for data := range queue {
		buf := make([]byte, 4+len(data.Data))
		binary.BigEndian.PutUint32(buf, uint32(data.Seq))
		copy(buf[4:], data.Data)
		n, err := conn.Write(buf)
		log.Printf("SendWorker dequeue to:%v Seq:%05d, len:%07d write:%07d", to, data.Seq, len(buf), n)
		nn += n
		if err != nil {
			log.Printf("conn.Write err: %v", err)
			break
		}
	}
	conn.Flush()
	f.Close()
	log.Printf("to:%v nn:%010d", to, nn)
	result <- err

}
