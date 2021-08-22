package intermediary

type Input struct {
	TypeName string
}

func (i Input) Name() string {
	return i.TypeName
}

func (Input) Description() string {
	defer func() {
		panic("(_type)Description should never be called")
	}()

	return ""
}

func (Input) String() string {
	defer func() {
		panic("(_type)String should never be called")
	}()

	return ""
}

func (Input) Error() error {
	defer func() {
		panic("(_type)Error should never be called")
	}()

	return nil
}
