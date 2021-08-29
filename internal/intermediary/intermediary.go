package intermediary

func IsIntermediary(i interface{}) bool {
	switch i.(type) {
	case Input, ListInput, NonNullInput:
		return true
	case Type, ListType, NonNullType:
		return true
	}
	return false
}
