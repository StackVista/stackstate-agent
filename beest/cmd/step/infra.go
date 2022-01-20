package step

type Yard struct {
	dir  string
	vars map[string]interface{}
}

func (yard *Yard) WorkingDir() string {
	return yard.dir
}

func (yard *Yard) Variables() map[string]interface{} {
	return yard.vars
}

///

type CreationStep struct {
	Yard
}

func Create(workingDir string, vars map[string]interface{}) *CreationStep {
	return &CreationStep{Yard{workingDir, vars}}
}

///

type DestroyStep struct {
	Yard
}

func Destroy(create *CreationStep) *DestroyStep {
	return &DestroyStep{Yard{create.dir, create.vars}}
}
