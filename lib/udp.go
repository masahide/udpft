package lib

import (
	"encoding/binary"
	"os"
)

func SendWorker(to string, queue <-chan DataQueue, result chan<- error) {
	var err error
	conn, _ := os.Create(to)
	for data := range queue {
		buf := make([]byte, 4+len(data.Data))
		binary.BigEndian.PutUint32(buf, uint32(data.Seq))
		copy(buf[4:], data.Data)
		_, err = conn.Write(buf)
		if err != nil {
			break
		}
	}
	conn.Close()
	result <- err

}
