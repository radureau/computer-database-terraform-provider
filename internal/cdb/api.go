package cdb

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

type companyJSON struct {
	ID                 string   `json:"id"`
	Name               string   `json:"name"`
	Location           string   `json:"location"`
	ComputerModelsURIs []string `json:"computerModels"`
}

type computerModelJSON struct {
	ID         string `json:"id"`
	Name       string `json:"name"`
	Release    string `json:"release"`
	CompanyURI string `json:"company"`
}

// leaves ComputerModelsURIs empty
func toCompanyJSON(company *Company) *companyJSON {
	return &companyJSON{
		ID:       company.ID,
		Name:     company.Name,
		Location: company.Location,
	}
}

func (c *companyJSON) Company(computerModels []ComputerModel) *Company {
	company := &Company{
		ID:       c.ID,
		Name:     c.Name,
		Location: c.Location,
	}
	if computerModels != nil {
		return company.WithComputerModels(computerModels...)
	}
	return company
}

// leaves CompanyURI empty
func toComputerModelJSON(computerModel *ComputerModel) *computerModelJSON {
	return &computerModelJSON{
		ID:      computerModel.ID,
		Name:    computerModel.Name,
		Release: computerModel.Release,
	}
}

func (c *computerModelJSON) ComputerModel(companyOpt ...*Company) *ComputerModel {
	company := (*Company)(nil)
	if len(companyOpt) > 0 {
		company = companyOpt[0]
	}
	return &ComputerModel{
		ID:      c.ID,
		Name:    c.Name,
		Release: c.Release,
		Company: company,
	}
}

type APIClient struct {
	client  *http.Client
	baseURL string
}

func NewAPIClient(baseURL string) *APIClient {
	return &APIClient{client: &http.Client{}, baseURL: baseURL}
}

func (c *APIClient) CreateCompany(company *Company) error {
	_company := toCompanyJSON(company)
	jsonData, err := json.Marshal(_company)
	if err != nil {
		return err
	}
	resp, err := c.client.Post(fmt.Sprintf("%s/companies", c.baseURL), "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusCreated {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}
	return nil
}

func (c *APIClient) GetCompany(id string) (*Company, error) {
	return c.getCompanyByURI(fmt.Sprintf("%s/companies/%s", c.baseURL, id), true)
}

func (c *APIClient) getCompanyByURI(URI string, load bool) (*Company, error) {
	resp, err := c.client.Get(URI)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}
	var _company *companyJSON
	err = json.NewDecoder(resp.Body).Decode(&_company)
	if err != nil {
		return nil, err
	}
	if !load {
		return _company.Company(nil), nil
	}
	computerModels := []ComputerModel{}
	for _, cmURI := range _company.ComputerModelsURIs {
		cm, err := c.getComputerModelByURI(cmURI, false)
		if err != nil {
			return nil, err
		}
		computerModels = append(computerModels, *cm)
	}
	return _company.Company(computerModels), nil
}

func (c *APIClient) UpdateCompany(company *Company) error {
	_company := toCompanyJSON(company)
	jsonData, err := json.Marshal(_company)
	if err != nil {
		return err
	}
	req, _ := http.NewRequest(http.MethodPut, fmt.Sprintf("%s/companies/%s", c.baseURL, company.ID), bytes.NewBuffer(jsonData))
	resp, err := c.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}
	return nil
}

func (c *APIClient) DeleteCompany(id string) error {
	req, _ := http.NewRequest(http.MethodDelete, fmt.Sprintf("%s/companies/%s", c.baseURL, id), nil)
	resp, err := c.client.Do(req)
	if err != nil {
		return err
	}
	if resp.StatusCode != http.StatusNoContent {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}
	return nil
}

func (c *APIClient) GetComputerModel(companyID, computerModelID string) (*ComputerModel, error) {
	return c.getComputerModelByURI(fmt.Sprintf("%s/companies/%s/computer-models/%s", c.baseURL, companyID, computerModelID), true)
}

func (c *APIClient) getComputerModelByURI(URI string, load bool) (*ComputerModel, error) {
	resp, err := c.client.Get(URI)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}
	var _computerModel *computerModelJSON
	err = json.NewDecoder(resp.Body).Decode(&_computerModel)
	if err != nil {
		return nil, err
	}
	if !load {
		return _computerModel.ComputerModel(), nil
	}
	company, _ := c.getCompanyByURI(_computerModel.CompanyURI, false)
	return _computerModel.ComputerModel(company), nil
}
