package step

type Yard struct {
	dir string
}

func (yard *Yard) WorkingDir() string {
	return yard.dir
}

///

type CreationStep struct {
	Yard
}

func Create(workingDir string) *CreationStep {
	return &CreationStep{Yard{workingDir}}
}

///

type DestroyStep struct {
	Yard
}

func Destroy(create *CreationStep) *DestroyStep {
	return &DestroyStep{Yard{create.dir}}
}
