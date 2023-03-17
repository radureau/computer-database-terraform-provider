package cdb

import "fmt"

type Company struct {
	ID             string
	Name           string
	Location       string
	ComputerModels *[]ComputerModel
}
type ComputerModel struct {
	ID      string
	Name    string
	Release string
	Company *Company
}

func (c Company) WithComputerModels(computerModels ...ComputerModel) *Company {
	c.ComputerModels = &computerModels
	return &c
}

func (cm ComputerModel) WithCompany(company *Company) *ComputerModel {
	cm.Company = company
	return &cm
}

func (c Company) String() string {
	type alias Company
	return fmt.Sprintf("%+v", alias(c))
}

func (cm ComputerModel) String() string {
	type alias ComputerModel
	return fmt.Sprintf("%+v", alias(cm))
}
