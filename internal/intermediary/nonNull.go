package intermediary

type NonNullType struct {
	TypeName string
}

func (t NonNullType) Name() string {
	return t.TypeName
}

func (NonNullType) Description() string {
	defer func() {
		panic("(_nonNullType)Description should never be called")
	}()

	return ""
}

func (NonNullType) String() string {
	defer func() {
		panic("(_nonNullType)String should never be called")
	}()

	return ""
}

func (NonNullType) Error() error {
	defer func() {
		panic("(_nonNullType)Error should never be called")
	}()

	return nil
}
