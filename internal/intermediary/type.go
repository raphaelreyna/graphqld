package intermediary

type Type struct {
	TypeName string
}

func (t Type) Name() string {
	return t.TypeName
}

func (Type) Description() string {
	defer func() {
		panic("(_type)Description should never be called")
	}()

	return ""
}

func (Type) String() string {
	defer func() {
		panic("(_type)String should never be called")
	}()

	return ""
}

func (Type) Error() error {
	defer func() {
		panic("(_type)Error should never be called")
	}()

	return nil
}
