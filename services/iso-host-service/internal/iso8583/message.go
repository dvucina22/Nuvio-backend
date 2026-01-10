package iso8583

type Message struct {
	MTI    string
	Bitmap [8]byte
	Fields map[int][]byte
}
