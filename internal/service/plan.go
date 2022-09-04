package service

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"reflect"
	"strings"

	"github.com/AlecAivazis/survey/v2"
	"golang.org/x/exp/slices"
)

//type resource struct {
//Address         string                 `json:"address"`
//Mode            string                 `json:"mode"`
//TFType          string                 `json:"type"`
//Name            string                 `json:"name"`
//Index           interface{}            `json:"index"`
//ProviderName    string                 `json:"provider_name"`
//SchemaVersion   int                    `json:"schema_version"`
//Values          map[string]interface{} `json:"values"`
//SensitiveValues map[string]bool        `json:"sensitive_values"`
//}

//type module struct {
//Resources []resource `json:"resources"`
//}

type afterUnknown struct {
	ID bool `json:"id"`
}

type Change struct {
	Actions      []string               `json:"actions"`
	Before       map[string]interface{} `json:"before"`
	After        map[string]interface{} `json:"after"`
	AfterUnknown afterUnknown           `json:"after_unknown"`
}

type ResourceChange struct {
	Address         string      `json:"address"`
	PreviousAddress string      `json:"previous_address"`
	ModuleAddress   string      `json:"module_adderss"`
	Mode            string      `json:"mode"`
	TFType          string      `json:"type"` // type is an illegal name, so tftype it is
	Name            string      `json:"name"`
	Change          Change      `json:"change"`
	Index           interface{} `json:"index"`
	Deposed         string      `json:"deposed"`
	ActionReason    string      `json:"action_reason"`
}

type plan struct {
	ResourceChanges []ResourceChange `json:"resource_changes"`
}

type Service struct {
}

func New() *Service {
	return &Service{}
}

func (s *Service) GeneratePlan() error {
	output, err := exec.Command("terraform", "plan", "-out", "tftargeter-plan").CombinedOutput()
	if err != nil {
		return fmt.Errorf(string(output), err)
	}

	return nil
}

func (s *Service) LoadPlan(userProvidedPlanFile string) (plan, error) {
	planfile := "tftargeter-plan"
	if userProvidedPlanFile != "" {
		planfile = userProvidedPlanFile
	}

	output, err := exec.Command("terraform", "show", "-json", planfile).CombinedOutput()
	if err != nil {
		return plan{}, fmt.Errorf(string(output), err)
	}

	var p plan
	err = json.Unmarshal(output, &p)
	if err != nil {
		return plan{}, fmt.Errorf("Got error while unmarshalling tf-plan to json: %v", err)
	}

	return p, err
}

func (s *Service) IsChanging(c Change) bool {
	if len(c.Actions) == 0 {
		return false
	}

	for _, a := range c.Actions {
		as := strings.ToLower(a)
		if as == "create" || as == "delete" || as == "update" {
			return true
		}
	}

	return false
}

func (s *Service) DeletePlan() error {
	output, err := exec.Command("rm", "tftargeter-plan").CombinedOutput()
	if err != nil {
		return fmt.Errorf(string(output), err)
	}

	return nil
}

func (s *Service) FormatCommand(targets []string) string {
	cmd := []string{"terraform apply"}
	for i, t := range targets {
		if i == len(targets)-1 {
			cmd = append(cmd, fmt.Sprintf("-target '%v'", t))
			continue
		}
		cmd = append(cmd, fmt.Sprintf("-target '%v' \\\n", t))
	}
	return strings.Join(cmd, " ")
}

func (s *Service) PrettyPrintDiff(changes []ResourceChange) {
	fmt.Println("Theses are the changes planned for the selected resources:")
	for i, c := range changes {
		fmt.Println(c.Address)
		if reflect.DeepEqual(c.Change.Actions, []string{"create"}) {
			// Pure create
			for k, v := range c.Change.After {
				fmt.Printf("+ %v: %v\n", k, v)
			}
		} else if reflect.DeepEqual(c.Change.Actions, []string{"delete"}) {
			// Pure delete
			for k, v := range c.Change.After {
				fmt.Printf("- %v: %v\n", k, v)
			}
		} else {
			// Update, or delete + create
			for k, val := range c.Change.After {
				beforeVal := c.Change.Before[k]
				// This can almost certainly be made nicer with a visual for create + delete as opposed to just a change
				if beforeVal == nil {
					// create
					fmt.Printf("+ %v: %v -> %v \n", k, beforeVal, val)
				} else if val == nil {
					// delete
					fmt.Printf("- %v: %v -> %v \n", k, beforeVal, val)
				} else {
					// update
					fmt.Printf("~ %v: %v -> %v \n", k, beforeVal, val)
				}
			}
		}

		if i < len(changes)-1 {
			fmt.Println("============================================")
		}
	}
}

func (s *Service) GetAndConfirmTargets(p plan) []string {
	var confirmed bool
	var names []string
	var targets []string

	for _, c := range p.ResourceChanges {
		if s.IsChanging(c.Change) {
			names = append(names, c.Address)
		}
	}

	for !confirmed {
		targets = []string{}
		prompt := &survey.MultiSelect{
			Options:  names,
			PageSize: 20,
		}

		if err := survey.AskOne(prompt, &targets, survey.WithKeepFilter(true)); err != nil {
			panic(err)
		}

		var changes []ResourceChange
		for _, c := range p.ResourceChanges {
			if slices.Contains(targets, c.Address) {
				changes = append(changes, c)
			}
		}

		s.PrettyPrintDiff(changes)

		confirm := &survey.Confirm{
			Message: "Are these the target resources you wanted to select?",
		}

		if err := survey.AskOne(confirm, &confirmed); err != nil {
			panic(err)
		}
	}
	return targets
}
