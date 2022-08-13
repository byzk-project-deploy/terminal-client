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

func PromptPassword(label string) *promptui.Prompt {
	return &promptui.Prompt{
		Label: label,
		Validate: func(s string) error {
			if len(s) <= 0 {
				return fmt.Errorf("密码不能为空")
			}
			return nil
		},
		Mask: '*',
	}
}

func PromptConfirm(label string) *promptui.Prompt {
	return &promptui.Prompt{
		Label:     label,
		IsConfirm: true,
	}
}
