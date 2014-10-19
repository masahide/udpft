package lib

import (
	"encoding/binary"
	"os"
)

func SendWorker(to string, datas <-chan DataQueue, result chan<- error) {
	var err error
	conn, _ := os.Open(to)
	for data := range datas {
		buf := make([]byte, 4+SendSize)
		binary.BigEndian.PutUint32(buf, uint32(data.Seq))
		copy(buf[4:], data.Data)
		_, err := conn.Write(buf)
		if err != nil {
			break
		}
	}
	conn.Close()
	result <- err

}
