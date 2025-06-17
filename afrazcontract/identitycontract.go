package main

import (
	"encoding/json"
	"fmt"

	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

// SmartContract provides functions for managing identities
type SmartContract struct {
	contractapi.Contract
}

// Identity describes personal identity details
type Identity struct {
	ID                  string `json:"id"`
	Title               string `json:"title"`
	FirstName           string `json:"firstName"`
	MiddleName          string `json:"middleName"`
	LastName            string `json:"lastName"`
	NameOnCard          string `json:"nameOnCard"`
	ElevenCharName      string `json:"elevenCharName"`
	CNIC                string `json:"cnic"`
	CNICIssueDate       string `json:"cnicIssueDate"`
	CNICExpiryDate      string `json:"cnicExpiryDate"`
	OldNIC              string `json:"oldNIC"`
	PassportNumber      string `json:"passportNumber"`
	Nationality         string `json:"nationality"`
	PassportIssueDate   string `json:"passportIssueDate"`
	PassportExpiryDate  string `json:"passportExpiryDate"`
	DateOfBirth         string `json:"dateOfBirth"`
	PlaceOfBirth        string `json:"placeOfBirth"`
	Gender              string `json:"gender"`
	FatherOrHusbandName string `json:"fatherOrHusbandName"`
	MotherMaidenName    string `json:"motherMaidenName"`
	MaritalStatus       string `json:"maritalStatus"`
	Education           string `json:"education"`
	PoliticalAffiliation string `json:"politicalAffiliation"`
	TaxPayer            string `json:"taxPayer"`
	Address             string `json:"address"`
	Landline            string `json:"landline"`
	PostalCode          string `json:"postalCode"`
	NoOfDependents      string `json:"noOfDependents"`
	NTN                 string `json:"ntn"`
	ResidenceType       string `json:"residenceType"`
	ApartmentOrHouse    string `json:"apartmentOrHouse"`
	ResidenceNature     string `json:"residenceNature"`
	MobileNumber        string `json:"mobileNumber"`
}

// InitLedger adds a base set of identities to the ledger
func (s *SmartContract) InitLedger(ctx contractapi.TransactionContextInterface) error {
	identities := []Identity{
		{
			ID: "identity1",
			Title: "Mr.",
			FirstName: "John",
			LastName: "Doe",
			CNIC: "12345-6789012-3",
			DateOfBirth: "01-01-1980",
			Gender: "Male",
			MobileNumber: "03001234567",
		},
	}

	for _, identity := range identities {
		identityJSON, err := json.Marshal(identity)
		if err != nil {
			return err
		}

		err = ctx.GetStub().PutState(identity.ID, identityJSON)
		if err != nil {
			return fmt.Errorf("failed to put to world state: %v", err)
		}
	}

	return nil
}

// CreateIdentity issues a new identity to the world state with given details.
func (s *SmartContract) CreateIdentity(ctx contractapi.TransactionContextInterface, id string, title string, firstName string, lastName string, cnic string, dob string, gender string, mobile string) error {
	exists, err := s.IdentityExists(ctx, id)
	if err != nil {
		return err
	}
	if exists {
		return fmt.Errorf("the identity %s already exists", id)
	}

	identity := Identity{
		ID:         id,
		Title:      title,
		FirstName:  firstName,
		LastName:   lastName,
		CNIC:       cnic,
		DateOfBirth: dob,
		Gender:     gender,
		MobileNumber: mobile,
	}
	identityJSON, err := json.Marshal(identity)
	if err != nil {
		return err
	}

	return ctx.GetStub().PutState(id, identityJSON)
}

// ReadIdentity returns the identity stored in the world state with given id.
func (s *SmartContract) ReadIdentity(ctx contractapi.TransactionContextInterface, id string) (*Identity, error) {
	identityJSON, err := ctx.GetStub().GetState(id)
	if err != nil {
		return nil, fmt.Errorf("failed to read from world state: %v", err)
	}
	if identityJSON == nil {
		return nil, fmt.Errorf("the identity %s does not exist", id)
	}

	var identity Identity
	err = json.Unmarshal(identityJSON, &identity)
	if err != nil {
		return nil, err
	}

	return &identity, nil
}

// UpdateIdentity updates an existing identity in the world state with provided parameters.
func (s *SmartContract) UpdateIdentity(ctx contractapi.TransactionContextInterface, id string, mobile string, address string) error {
	exists, err := s.IdentityExists(ctx, id)
	if err != nil {
		return err
	}
	if !exists {
		return fmt.Errorf("the identity %s does not exist", id)
	}

	// Get current identity
	identity, err := s.ReadIdentity(ctx, id)
	if err != nil {
		return err
	}

	// Update fields
	identity.MobileNumber = mobile
	identity.Address = address

	identityJSON, err := json.Marshal(identity)
	if err != nil {
		return err
	}

	return ctx.GetStub().PutState(id, identityJSON)
}

// DeleteIdentity deletes an given identity from the world state.
func (s *SmartContract) DeleteIdentity(ctx contractapi.TransactionContextInterface, id string) error {
	exists, err := s.IdentityExists(ctx, id)
	if err != nil {
		return err
	}
	if !exists {
		return fmt.Errorf("the identity %s does not exist", id)
	}

	return ctx.GetStub().DelState(id)
}

// IdentityExists returns true when identity with given ID exists in world state
func (s *SmartContract) IdentityExists(ctx contractapi.TransactionContextInterface, id string) (bool, error) {
	identityJSON, err := ctx.GetStub().GetState(id)
	if err != nil {
		return false, fmt.Errorf("failed to read from world state: %v", err)
	}

	return identityJSON != nil, nil
}

// GetAllIdentities returns all identities found in world state
func (s *SmartContract) GetAllIdentities(ctx contractapi.TransactionContextInterface) ([]*Identity, error) {
	resultsIterator, err := ctx.GetStub().GetStateByRange("", "")
	if err != nil {
		return nil, err
	}
	defer resultsIterator.Close()

	var identities []*Identity
	for resultsIterator.HasNext() {
		queryResponse, err := resultsIterator.Next()
		if err != nil {
			return nil, err
		}

		var identity Identity
		err = json.Unmarshal(queryResponse.Value, &identity)
		if err != nil {
			return nil, err
		}
		identities = append(identities, &identity)
	}

	return identities, nil
}

func main() {
	chaincode, err := contractapi.NewChaincode(&SmartContract{})
	if err != nil {
		fmt.Printf("Error creating identity chaincode: %v", err)
		return
	}

	if err := chaincode.Start(); err != nil {
		fmt.Printf("Error starting identity chaincode: %v", err)
	}
}