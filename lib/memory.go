package lib

var memoryBuf map[int]DataBlock

type DataQueue struct {
	Seq  int
	Data []byte
}

func FileToMemory(done chan struct{}, path string, queue chan<- DataQueue, end <-chan int) <-chan error {
	datas, errc := FileRead(done, path)

	go func() {
		for {
			select {
			case data, ok := <-datas:
				if !ok {
					datas = nil
				} else {
					memoryBuf[data.id] = data
					memoryToQueue(queue, data.id)
				}
			case endID, ok := <-end:
				if !ok {
					end = nil
				} else {
					delete(memoryBuf, endID)
				}
			case <-done:
				return
			}
			if datas == nil && end == nil {
				break
			}
		}
	}()
	return errc

}

const SendSize = 1450

func memoryToQueue(queue chan<- DataQueue, dataId int) {
	size := memoryBuf[dataId].size
	seq := 0
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

}
