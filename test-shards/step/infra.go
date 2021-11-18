package step

type Hive struct {
	dir string
}

func (h *Hive) WorkingDir() string {
	return h.dir
}

///

type CreationStep struct {
	//Module   *tfconfig.Module
	//tf       *tfexec.Terraform
	//planPath string
	Hive
}

func Create(workingDir string) *CreationStep {
	return &CreationStep{Hive{workingDir}}
}

///

type DestroyStep struct {
	Hive
}

func Destroy(create *CreationStep) *DestroyStep {
	return &DestroyStep{Hive{create.dir}}
}
