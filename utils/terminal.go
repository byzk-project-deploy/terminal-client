package utils

import (
	"fmt"
	"github.com/byzk-project-deploy/promptui"
)

func PromptNotEmptyVerify(label string) *promptui.Prompt {
	return &promptui.Prompt{
		Label: label,
		Validate: func(s string) error {
			if s == "" {
				return fmt.Errorf("%s 不能为空", label)
			}
			return nil
		},
	}
}
