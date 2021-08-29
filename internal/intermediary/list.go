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

type ListInput struct {
	InputName string
}

func (t ListInput) Name() string {
	return t.InputName
}

func (ListInput) Description() string {
	defer func() {
		panic("(_listInput)Description should never be called")
	}()

	return ""
}

func (ListInput) String() string {
	defer func() {
		panic("(_listInput)String should never be called")
	}()

	return ""
}

func (ListInput) Error() error {
	defer func() {
		panic("(_listInput)Error should never be called")
	}()

	return nil
}
