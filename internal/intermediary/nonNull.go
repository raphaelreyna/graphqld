package intermediary

type NonNullType struct {
	TypeName string
}

func (t NonNullType) Name() string {
	return t.TypeName
}

func (NonNullType) Description() string {
	defer func() {
		panic("(_NonNullType)Description should never be called")
	}()

	return ""
}

func (NonNullType) String() string {
	defer func() {
		panic("(_NonNullType)String should never be called")
	}()

	return ""
}

func (NonNullType) Error() error {
	defer func() {
		panic("(_NonNullType)Error should never be called")
	}()

	return nil
}

type NonNullInput struct {
	InputName string
}

func (i NonNullInput) Name() string {
	return i.InputName
}

func (NonNullInput) Description() string {
	defer func() {
		panic("(_NonNullInput)Description should never be called")
	}()

	return ""
}

func (NonNullInput) String() string {
	defer func() {
		panic("(_NonNullInput)String should never be called")
	}()

	return ""
}

func (NonNullInput) Error() error {
	defer func() {
		panic("(_NonNullInput)Error should never be called")
	}()

	return nil
}
