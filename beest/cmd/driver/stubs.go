package driver

import (
	"beest/cmd/step"
)

type StubProvisioner struct {
	createCalled, destroyCalled bool
	failCreate, failDestroy     error
}

func NewStubProvisioner(failCreate, failDestroy error) *StubProvisioner {
	return &StubProvisioner{failCreate: failCreate, failDestroy: failDestroy}
}
func (sp *StubProvisioner) Create(*step.CreationStep, bool) error {
	sp.createCalled = true
	if sp.failCreate != nil {
		return sp.failCreate
	}
	return nil
}
func (sp *StubProvisioner) Destroy(*step.DestroyStep, bool) error {
	sp.destroyCalled = true
	if sp.failDestroy != nil {
		return sp.failDestroy
	}
	return nil
}
func (sp *StubProvisioner) Created() bool {
	return sp.createCalled
}
func (sp *StubProvisioner) Destroyed() bool {
	return sp.destroyCalled
}

//

type StubDeployer struct {
	prepareCalled, cleanupCalled bool
	failPrepare, failCleanup     error
}

func NewStubDeployer(failPrepare, failCleanup error) *StubDeployer {
	return &StubDeployer{failPrepare: failPrepare, failCleanup: failCleanup}
}
func (sd *StubDeployer) Prepare(*step.PrepareStep, []string) error {
	sd.prepareCalled = true
	if sd.failPrepare != nil {
		return sd.failPrepare
	}
	return nil
}
func (sd *StubDeployer) Cleanup(*step.CleanupStep, []string) error {
	sd.cleanupCalled = true
	if sd.failCleanup != nil {
		return sd.failCleanup
	}
	return nil
}
func (sd *StubDeployer) Prepared() bool {
	return sd.prepareCalled
}
func (sd *StubDeployer) Cleaned() bool {
	return sd.cleanupCalled
}

//

type StubVerifier struct {
	verifyCalled bool
	failVerify   error
}

func NewStubVerifier(failVerify error) *StubVerifier {
	return &StubVerifier{failVerify: failVerify}
}

func (sv *StubVerifier) Verify(*step.VerificationStep, bool, string) error {
	sv.verifyCalled = true
	if sv.failVerify != nil {
		return sv.failVerify
	}
	return nil
}
func (sv *StubVerifier) Verified() bool {
	return sv.verifyCalled
}
