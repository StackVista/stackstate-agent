package step

import "fmt"

func (yard *Yard) Inventory() string {
	return fmt.Sprintf("%s/ansible_inventory", yard.workingDir)
}

func (yard *Yard) Playbook() string {
	return fmt.Sprintf("%s/playbook.yml", yard.workingDir)
}

///

type PrepareStep struct {
	Yard
}

func Prepare(create *CreationStep) *PrepareStep {
	return &PrepareStep{create.Yard}
}

///

type VerificationStep struct {
	Yard
	hosts []string
}

func Verify(prepare *PrepareStep, hosts []string) *VerificationStep {
	return &VerificationStep{
		prepare.Yard,
		hosts,
	}
}

func (v *VerificationStep) TestsPath() string {
	return v.Yard.testingDir
}

func (v *VerificationStep) Hostnames() []string {
	return v.hosts
}

///

type CleanupStep struct {
	Yard
}

func Cleanup(prepare *PrepareStep) *CleanupStep {
	return &CleanupStep{prepare.Yard}
}
