package intermediary

type ListType struct {
	TypeName string
}

func (t ListType) Name() string {
	return t.TypeName
}

func (ListType) Description() string {
	defer func() {
		panic("(_listType)Description should never be called")
	}()

	return ""
}

func (ListType) String() string {
	defer func() {
		panic("(_listType)String should never be called")
	}()

	return ""
}

func (ListType) Error() error {
	defer func() {
		panic("(_listType)Error should never be called")
	}()

	return nil
}
