package cxcore

func opSerialize(expr *CXExpression, fp int) {
	inp1, out1 := expr.Inputs[0], expr.Outputs[0]
	out1Offset := GetFinalOffset(fp, out1)

	_ = inp1

	var slcOff int
	byts := SerializeCXProgram(PROGRAM, true)
	for _, b := range byts {
		slcOff = WriteToSlice(slcOff, []byte{b})
	}

	WriteI32(out1Offset, int32(slcOff))
}

func opDeserialize(expr *CXExpression, fp int) {
	inp := expr.Inputs[0]

	inpOffset := GetFinalOffset(fp, inp)

	off := Deserialize_i32(PROGRAM.Memory[inpOffset : inpOffset+TYPE_POINTER_SIZE])

	_l := PROGRAM.Memory[off+OBJECT_HEADER_SIZE : off+OBJECT_HEADER_SIZE+SLICE_HEADER_SIZE]
	l := Deserialize_i32(_l[4:8])

	Deserialize(PROGRAM.Memory[off+OBJECT_HEADER_SIZE+SLICE_HEADER_SIZE : off+OBJECT_HEADER_SIZE+SLICE_HEADER_SIZE+l]) // BUG : should be l * elt.TotalSize ?
}
