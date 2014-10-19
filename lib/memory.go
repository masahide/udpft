package lib

import "log"

var memoryBuf map[int]DataBlock

type DataQueue struct {
	Seq  int
	Data []byte
}

func FileToMemory(done chan struct{}, path string, queue chan<- DataQueue, end <-chan int) <-chan error {
	datas, errc := FileRead(done, path)

	memoryBuf = make(map[int]DataBlock)

	go func() {
		seq := 0
		for {
			select {
			case data, ok := <-datas:
				if !ok {
					datas = nil
				} else {
					memoryBuf[data.id] = data
					log.Printf("toMemory: %v, size: %v", data.id, len(data.data))
					seq = memoryToQueue(seq, queue, data.id)
				}
			case endID, ok := <-end:
				if !ok {
					end = nil
				} else {
					delete(memoryBuf, endID)
				}
			case <-done:
				close(queue)
				return
			}
			if datas == nil && end == nil {
				log.Printf("seq:%v", seq)
				close(queue)
				return
			}
		}
	}()
	return errc

}

const SendSize = 1450

func memoryToQueue(seq int, queue chan<- DataQueue, dataId int) int {
	size := memoryBuf[dataId].size
	for current := 0; current < size; current += SendSize {
		next := current + SendSize
		if next > size {
			next = size
		}
		sendBuf := make([]byte, next-current)
		copy(sendBuf, memoryBuf[dataId].data[current:next])
		queue <- DataQueue{Seq: seq, Data: sendBuf}
		seq++
	}
	return seq

}
