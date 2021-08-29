package intermediary

type Type struct {
	TypeName string
}

func (t Type) Name() string {
	return t.TypeName
}

func (Type) Description() string {
	defer func() {
		panic("(_Type)Description should never be called")
	}()

	return ""
}

func (Type) String() string {
	defer func() {
		panic("(_Type)String should never be called")
	}()

	return ""
}

func (Type) Error() error {
	defer func() {
		panic("(_Type)Error should never be called")
	}()

	return nil
}
