package bitio

// leftShift shifts byte array for left by n bits.
func leftShift(p []byte, bits uint) {
	// byte copy
	if bits > 8 {
		size := int(bits / 8)
		bits %= 8

		for i := 0; i < len(p)-size; i++ {
			ii := size + i
			p[ii-size] = p[ii]
		}

		for i := 0; i < size; i++ {
			p[len(p)-1-i] = 0
		}
	}

	// bit carry
	for i := 0; i < len(p)-1; i++ {
		l := p[i+1] >> (8 - bits)
		h := p[i] << bits
		p[i] = h | l
	}

	// bit carry [right end]
	p[len(p)-1] <<= bits
}

// rightShift shifts byte array at right by n bits.
func rightShift(p []byte, bits uint) {
	// byte copy
	if bits > 8 {
		size := int(bits / 8)
		bits %= 8

		for i := 0; i < len(p)-size; i++ {
			ii := len(p) - 1 - i
			p[ii] = p[ii-size]
		}

		for i := 0; i < size; i++ {
			p[i] = 0
		}
	}

	// bit carry
	for i := len(p) - 1; i > 0; i-- {
		h := p[i-1] << (8 - bits)
		l := p[i] >> bits
		p[i] = h | l
	}

	// bit carry [left end]
	p[0] >>= bits
}
