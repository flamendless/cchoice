package enums

func StringToPBEnum[T ~string, PB ~int32](val T, pbmap map[string]int32, dft PB) PB {
	v, ok := pbmap[string(val)]
	if !ok {
		return dft
	}
	return PB(v)
}
