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

type upsertCompanyRequest struct {
	companyJSON
	ComputerModels []computerModelJSON `json:"computerModels"`
}

func toUpsertCompanyRequest(company *Company) (req upsertCompanyRequest) {
	req.companyJSON = *toCompanyJSON(company)
	if company.ComputerModels == nil {
		return
	}
	for _, cm := range *company.ComputerModels {
		req.ComputerModels = append(req.ComputerModels, *toComputerModelJSON(&cm))
	}
	return
}

type APIClient struct {
	client  *http.Client
	baseURL string
}

func NewAPIClient(baseURL string) *APIClient {
	return &APIClient{client: &http.Client{}, baseURL: baseURL}
}

func (c *APIClient) CreateCompany(company *Company) error {

	upsertRequest := toUpsertCompanyRequest(company)
	jsonData, err := json.Marshal(upsertRequest)
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
	var jsonableCompany *companyJSON
	err = json.NewDecoder(resp.Body).Decode(&jsonableCompany)
	if err != nil {
		return nil, err
	}
	if !load {
		return jsonableCompany.Company(nil), nil
	}
	computerModels := []ComputerModel{}
	for _, cmURI := range jsonableCompany.ComputerModelsURIs {
		cm, err := c.getComputerModelByURI(cmURI, false)
		if err != nil {
			return nil, err
		}
		computerModels = append(computerModels, *cm)
	}
	return jsonableCompany.Company(computerModels), nil
}

func (c *APIClient) UpdateCompany(company *Company) error {
	upsertRequest := toUpsertCompanyRequest(company)
	jsonData, err := json.Marshal(upsertRequest)
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
	var jsonableComputerModel *computerModelJSON
	err = json.NewDecoder(resp.Body).Decode(&jsonableComputerModel)
	if err != nil {
		return nil, err
	}
	if !load {
		return jsonableComputerModel.ComputerModel(), nil
	}
	company, _ := c.getCompanyByURI(jsonableComputerModel.CompanyURI, false)
	return jsonableComputerModel.ComputerModel(company), nil
}

// Requires computerModel.Company not null and computerModel.Company.ID not empty
func (c *APIClient) CreateComputerModel(computerModel *ComputerModel) error {
	jsonableComputerModel := toComputerModelJSON(computerModel)
	jsonData, err := json.Marshal(jsonableComputerModel)
	if err != nil {
		return err
	}
	resp, err := c.client.Post(
		fmt.Sprintf("%s/companies/%s/computer-models", c.baseURL, computerModel.Company.ID),
		"application/json", bytes.NewBuffer(jsonData),
	)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusCreated {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}
	return nil
}

func (c *APIClient) DeleteComputerModel(companyID, computerModelID string) error {
	req, _ := http.NewRequest(http.MethodDelete, fmt.Sprintf("%s/companies/%s/computer-models/%s", c.baseURL, companyID, computerModelID), nil)
	resp, err := c.client.Do(req)
	if err != nil {
		return err
	}
	if resp.StatusCode != http.StatusNoContent {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}
	return nil
}
