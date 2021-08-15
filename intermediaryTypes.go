package main

type _type struct {
	name string
}

func (t _type) Name() string {
	return t.name
}

func (_type) Description() string {
	defer func() {
		panic("(_type)Description should never be called")
	}()

	return ""
}

func (_type) String() string {
	defer func() {
		panic("(_type)String should never be called")
	}()

	return ""
}

func (_type) Error() error {
	defer func() {
		panic("(_type)Error should never be called")
	}()

	return nil
}

type _listType struct {
	name string
}

func (t _listType) Name() string {
	return t.name
}

func (_listType) Description() string {
	defer func() {
		panic("(_listType)Description should never be called")
	}()

	return ""
}

func (_listType) String() string {
	defer func() {
		panic("(_listType)String should never be called")
	}()

	return ""
}

func (_listType) Error() error {
	defer func() {
		panic("(_listType)Error should never be called")
	}()

	return nil
}

type _nonNullType struct {
	name string
}

func (t _nonNullType) Name() string {
	return t.name
}

func (_nonNullType) Description() string {
	defer func() {
		panic("(_nonNullType)Description should never be called")
	}()

	return ""
}

func (_nonNullType) String() string {
	defer func() {
		panic("(_nonNullType)String should never be called")
	}()

	return ""
}

func (_nonNullType) Error() error {
	defer func() {
		panic("(_nonNullType)Error should never be called")
	}()

	return nil
}
