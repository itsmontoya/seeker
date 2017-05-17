package seeker

func getNewlineIndex(s []byte) (idx int) {
	for i, b := range s {
		if b == charNewline {
			return i
		}
	}

	return -1
}

func reverseByteSlice(bs []byte) {
	var n int
	c := len(bs) - 1
	for i := range bs {
		if n = c - i; n == i || n < i {
			break
		}

		bs[i], bs[n] = bs[n], bs[i]
	}
}
