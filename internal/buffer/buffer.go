package buffer

var buff *Buffer

type buffer interface {
	BufferWorker()
}

type Buffer struct {
	Pipe     chan Messages
	Messages []Messages
}
type Messages struct {
	Msg  string `json:"msg"`
	User string `json:"user"`
}

func NewBuffer() *Buffer {
	if buff != nil {
		buff = nil
	}

	pool := []Messages{}
	newbuf := Buffer{
		Pipe:     make(chan Messages),
		Messages: pool,
	}
	buff = &newbuf
	return buff
}

func (b *Buffer) BufferWorker() {
	go func() {
		for {
			if len(b.Messages) == 100 {
				clear(b.Messages)
			}
			msg := <-b.Pipe

			b.Messages = append(b.Messages, msg)
		}
	}()
}
