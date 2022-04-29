package step

import "fmt"

func (yard *Yard) Inventory() string {
	return fmt.Sprintf("%s/ansible_inventory", yard.dir)
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
	testsPath string
	hosts     []string
}

func Verify(prepare *PrepareStep, testsPath string, hosts []string) *VerificationStep {
	return &VerificationStep{
		prepare.Yard,
		testsPath,
		hosts,
	}
}

func (v *VerificationStep) TestsPath() string {
	return v.testsPath
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
