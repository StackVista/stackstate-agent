package driver

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-config-inspect/tfconfig"
	"github.com/hashicorp/terraform-exec/tfexec"
	"log"
	"os"
	"strconv"
	"strings"
)

const (
	DefaultTerraformBinary = "/usr/local/bin/terraform"
)

type TerraformContext interface {
	WorkingDir() string
	Variables() map[string]interface{}
}

type TerraformState struct {
	Module   *tfconfig.Module
	tf       *tfexec.Terraform
	planPath string
}

func TerraformApply(ctx TerraformContext, prompt bool) {
	apply(ctx, false, prompt)
}

func TerraformDestroy(ctx TerraformContext, prompt bool) {
	apply(ctx, true, prompt)
}

func apply(ctx TerraformContext, destroy bool, prompt bool) {
	state := newTerraform(ctx)
	state.init()
	state.validate()
	log.Println(fmt.Sprintf("Variables: %v", ctx.Variables()))
	if state.plan(ctx.Variables(), destroy) {
		state.apply(prompt)
	} else {
		log.Println("No Terraform changes detected.")
	}
}

func newTerraform(ctx TerraformContext) *TerraformState {
	log.Println(fmt.Sprintf("Preparing Terraform to run from: %s", ctx.WorkingDir()))

	tf, err := tfexec.NewTerraform(ctx.WorkingDir(), DefaultTerraformBinary)
	if err != nil {
		log.Fatalf("Error running NewTerraform: %s", err)
	}

	// TODO debug
	//tf.SetStdout(os.Stdout)

	return &TerraformState{
		tf:     tf,
		Module: loadModule(tf.WorkingDir()),
	}
}

func loadModule(moduleDir string) *tfconfig.Module {
	module, diags := tfconfig.LoadModule(moduleDir)
	if diags.HasErrors() {
		log.Fatalf("Error reading module: %s", diags.Error())
	}
	return module
}

func (ts *TerraformState) init() {
	log.Println("Initializing Terraform workspace ...")
	err := ts.tf.Init(context.Background(), tfexec.Upgrade(true))
	if err != nil {
		log.Fatalf("Error running Init: %s", err)
	}
}

func (ts *TerraformState) validate() {
	log.Println("Validating Terraform manifests ...")
	validated, err := ts.tf.Validate(context.Background())
	if err != nil {
		log.Fatalf("Error running Validate: %s", err)
	}
	if !validated.Valid {
		for _, d := range validated.Diagnostics {
			log.Fatalf("Manifests not valid: %s\n\n  %s\n\n%s", d.Summary, d.Snippet.Code, d.Detail)
		}
	}
}

func (ts *TerraformState) plan(vars map[string]interface{}, destroy bool) bool {
	log.Println("Planning Terraform changes ...")
	var tfPlanOptions []tfexec.PlanOption
	for k, v := range vars {
		tfPlanOptions = append(tfPlanOptions, tfexec.Var(fmt.Sprintf("%s=%s", k, v)))
	}
	planPath := fmt.Sprintf("%s/tf.deploy", ts.Module.Path)
	tfPlanOptions = append(tfPlanOptions, tfexec.Out(planPath))

	if destroy {
		tfPlanOptions = append(tfPlanOptions, tfexec.Destroy(true))
	}

	hasChanges, err := ts.tf.Plan(context.Background(), tfPlanOptions...)
	if err != nil {
		log.Fatalf("Error running Plan: %s", err)
	}
	ts.planPath = planPath
	return hasChanges
}

func (ts *TerraformState) apply(prompt bool) {
	changes, err := ts.tf.ShowPlanFileRaw(context.Background(), ts.planPath)
	if err != nil {
		log.Fatalf("Error showing Plan: %s", err)
	}
	log.Println(changes)

	confirmed := true //by default, we do not ask for confirmation
	if prompt {
		confirmed = confirm()
	}

	if confirmed {
		log.Println("Applying Terraform changes ...")
		err = ts.tf.Apply(context.Background(), tfexec.DirOrPlan(ts.planPath))
		if err != nil {
			log.Fatalf("Error running Plan: %s", err)
		} else {
			log.Println("Terraform changes are done")
		}
	}
}

func confirm() bool {
	log.Println("Wanna continue? [y/N]")

	var response string
	_, err := fmt.Scanln(&response)
	if err != nil {
		log.Fatal(err)
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

func (ts *TerraformState) state() {
	state, err := ts.tf.Show(context.Background())
	if err != nil {
		log.Fatalf("Error running Show: %s", err)
	}

	fmt.Println(state.FormatVersion)
}

func (ts *TerraformState) output() {
	outs, err := ts.tf.Output(context.Background())
	if err != nil {
		log.Fatalf("Error running Output: %s", err)
	}
	for k, v := range outs {
		raw, err := v.Value.MarshalJSON()
		if err != nil {
			log.Fatalf("Error retrieving raw value '%s': %s", k, err)
		}
		fmt.Println(fmt.Sprintf("Writing output %s ...", k))
		outPath := fmt.Sprintf("%s/%s", ts.Module.Path, k)

		// output use -json option, but the out is not json, we just need to get the string and write it to a file
		strOut, err := strconv.Unquote(string(raw))
		if err != nil {
			log.Fatalf("Error converting raw value: %s", outPath)
		}

		//by default we assume all outputs to be sensitive
		err = os.WriteFile(outPath, []byte(strOut), 0600)
		if err != nil {
			log.Fatalf("Error writing output to: %s", outPath)
		}
	}
}
