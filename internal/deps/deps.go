package deps

var dependencies = &Deps{}

type Deps struct{}

func NewDeps() {
	dependencies = &Deps{}
}

func GetDeps() *Deps {
	return dependencies
}
