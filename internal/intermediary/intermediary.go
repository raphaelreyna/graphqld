package intermediary

func IsIntermediary(i interface{}) bool {
	switch i.(type) {
	case Type, ListType, NonNullType:
		return true
	}
	return false
}
