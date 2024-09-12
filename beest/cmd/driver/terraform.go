package driver

import (
	"beest/cmd/step"
	"context"
	"errors"
	"fmt"
	"github.com/hashicorp/terraform-config-inspect/tfconfig"
	"github.com/hashicorp/terraform-exec/tfexec"
	"log"
	"os"
	"strconv"
	"strings"
)

const (
	DefaultTerraformBinary = "/usr/local/bin/terragrunt"
)

type TerraformProvisioner struct{}

type TerraformState struct {
	module      *tfconfig.Module
	tf          *tfexec.Terraform
	planPath    string
	workspaceId string
}

func (tp *TerraformProvisioner) Create(step *step.CreationStep, prompt bool) error {
	return apply(step.Yard, false, prompt)
}

func (tp *TerraformProvisioner) Destroy(step *step.DestroyStep, prompt bool) error {
	return apply(step.Yard, true, prompt)
}

func apply(yard step.Yard, destroy bool, prompt bool) error {
	state, newErr := newTerraform(yard)
	if newErr != nil {
		return newErr
	}
	if err := state.selectWorkspace(); err != nil {
		return err
	}
	if err := state.init(); err != nil {
		return err
	}

	if hasChanges, planErr := state.plan(yard.Variables(), destroy); planErr == nil {
		if hasChanges {
			if applyErr := state.apply(prompt); applyErr != nil {
				return applyErr
			}
		} else {
			log.Println("No Terraform changes detected.")
		}
		return nil
	} else {
		return planErr
	}
}

func newTerraform(yard step.Yard) (*TerraformState, error) {
	log.Println(fmt.Sprintf("Terraform will run from: %s", yard.WorkingDir()))

	tf, newErr := tfexec.NewTerraform(yard.WorkingDir(), DefaultTerraformBinary)
	if newErr != nil {
		log.Printf("Error running NewTerraform: %s\n", newErr)
		return nil, newErr
	}
	if err := tf.SetLogPath("/go/src/app/terraform.log"); err != nil {
		log.Printf("Error setting log path: %s\n", err)
		return nil, err
	}

	if module, err := loadModule(tf.WorkingDir()); err != nil {
		return nil, err
	} else {
		return &TerraformState{
			tf:          tf,
			module:      module,
			workspaceId: yard.RunId(),
		}, nil
	}
}

func loadModule(moduleDir string) (*tfconfig.Module, error) {
	module, diags := tfconfig.LoadModule(moduleDir)
	if diags.HasErrors() {
		log.Printf("Error loading module: %s\n", diags.Error())
		return nil, errors.New(diags.Error())
	}
	return module, nil
}

func (ts *TerraformState) init() error {
	log.Println("Initializing Terraform ...")
	if err := ts.tf.Init(context.Background(), tfexec.Upgrade(true)); err != nil {
		log.Printf("Error running Init: %s\n", err)
		return err
	}
	return nil
}

func (ts *TerraformState) selectWorkspace() error {
	log.Printf("Selecting workspace %s ...\n", ts.workspaceId)
	// create and switches to the new workspace
	if err := ts.tf.WorkspaceNew(context.Background(), ts.workspaceId); err != nil {
		log.Printf("Could not create new workspace: %s\n", err)
		log.Printf("Trying to use existing workspace")
		// if new returns an error means that the workspace exists, so we select it
		if err = ts.tf.WorkspaceSelect(context.Background(), ts.workspaceId); err != nil {
			log.Printf("Error selecting workspace: %s\n", err)
			return err
		}
	}
	return nil
}

func (ts *TerraformState) validate() error {
	log.Println("Validating Terraform manifests ...")
	validated, err := ts.tf.Validate(context.Background())
	if err != nil {
		log.Printf("Error running Validate: %s\n", err)
		return err
	}
	if !validated.Valid {
		summaryStr := "Manifests not valid:"
		for _, d := range validated.Diagnostics {
			diagnosticStr := fmt.Sprintf("%s\n\n%s\n\n%s", d.Summary, d.Snippet.Code, d.Detail)
			summaryStr = fmt.Sprintf("%s\n\n%s", summaryStr, diagnosticStr)
		}
		log.Println(summaryStr)
		return errors.New(summaryStr)
	}
	return nil
}

func prettyVars(vars map[string]interface{}) string {
	var pretty []string
	for k, v := range vars {
		pretty = append(pretty, fmt.Sprintf("%s: %s", k, v))
	}
	return strings.Join(pretty, ", ")
}

func (ts *TerraformState) plan(vars map[string]interface{}, destroy bool) (bool, error) {
	log.Printf("Planning Terraform changes: destroy=%v, variables=[%s] ...\n", destroy, prettyVars(vars))
	var tfPlanOptions []tfexec.PlanOption
	for k, v := range vars {
		tfPlanOptions = append(tfPlanOptions, tfexec.Var(fmt.Sprintf("%s=%s", k, v)))
	}
	planPath := fmt.Sprintf("%s/tf.deploy", ts.module.Path)
	tfPlanOptions = append(tfPlanOptions, tfexec.Out(planPath))

	if destroy {
		tfPlanOptions = append(tfPlanOptions, tfexec.Destroy(true))
	}

	hasChanges, err := ts.tf.Plan(context.Background(), tfPlanOptions...)
	if err != nil {
		log.Printf("Error running Plan: %s\n", err)
		return false, err
	}
	ts.planPath = planPath
	return hasChanges, nil
}

func (ts *TerraformState) apply(prompt bool) error {
	if changes, err := ts.tf.ShowPlanFileRaw(context.Background(), ts.planPath); err != nil {
		log.Printf("Error showing Plan: %s\n", err)
		return err
	} else {
		log.Println(changes)
	}

	confirmed := true //by default, we do not ask for confirmation
	if prompt {
		confirmed = confirm()
	}

	if confirmed {
		log.Println("Applying Terraform changes ...")
		if err := ts.tf.Apply(context.Background(), tfexec.DirOrPlan(ts.planPath)); err != nil {
			log.Printf("Error running Plan: %s\n", err)
			return err
		} else {
			log.Println("Terraform changes are done")
		}
	}
	return nil
}

func confirm() bool {
	log.Println("Wanna continue? [y/N]")

	var response string
	_, err := fmt.Scanln(&response)
	if err != nil {
		log.Printf("Error reading input: %s\n", err)
		return false
	}

	switch strings.ToLower(response) {
	case "y", "yes":
		return true
	case "n", "no":
		return false
	default:
		return false
	}
}

func (ts *TerraformState) state() error {
	if state, err := ts.tf.Show(context.Background()); err != nil {
		log.Printf("Error running Show: %s\n", err)
		return err
	} else {
		fmt.Println(state.FormatVersion)
		return nil
	}
}

func (ts *TerraformState) output() []error {
	outs, err := ts.tf.Output(context.Background())
	if err != nil {
		log.Printf("Error running Output: %s\n", err)
		return []error{err}
	}
	var errs []error
	for k, v := range outs {
		raw, err := v.Value.MarshalJSON()
		if err != nil {
			log.Printf("Error retrieving raw value '%s': %s\n", k, err)
			errs = append(errs, err)
			continue
		}
		fmt.Printf("Writing output %s ...\n", k)
		outPath := fmt.Sprintf("%s/%s", ts.module.Path, k)

		// output use -json option, but the out is not json, we just need to get the string and write it to a file
		strOut, err := strconv.Unquote(string(raw))
		if err != nil {
			log.Printf("Error converting raw value: %s\n", outPath)
			errs = append(errs, err)
			continue
		}

		//by default we assume all outputs to be sensitive
		err = os.WriteFile(outPath, []byte(strOut), 0600)
		if err != nil {
			log.Printf("Error writing output to: %s\n", outPath)
			errs = append(errs, err)
		}
	}
	return errs
}
