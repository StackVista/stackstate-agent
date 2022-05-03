package step

type Yard struct {
	runId string
	dir   string
	vars  map[string]interface{}
}

func (yard *Yard) RunId() string {
	return yard.runId
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

func Create(runId, workingDir string, vars map[string]interface{}) *CreationStep {
	return &CreationStep{Yard{runId, workingDir, vars}}
}

///

type DestroyStep struct {
	Yard
}

func Destroy(create *CreationStep) *DestroyStep {
	return &DestroyStep{create.Yard}
}
