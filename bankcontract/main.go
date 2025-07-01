package main

import (
	"encoding/json"
	"fmt"

	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

type SmartContract struct {
	contractapi.Contract
}

type LoanApplication struct {
	ID           string  `json:"id"`
	Applicant    string  `json:"applicant"`
	Amount       int     `json:"amount"`
	Term         int     `json:"term"` // in months
	InterestRate float64 `json:"interestRate"`
	Status       string  `json:"status"`
}

// InitLedger initializes the ledger with some sample loan applications
func (s *SmartContract) InitLedger(ctx contractapi.TransactionContextInterface) error {
	loans := []LoanApplication{
		{ID: "loan1", Applicant: "Afraz", Amount: 10000, Term: 12, InterestRate: 5.5, Status: "Pending"},
		{ID: "loan2", Applicant: "Alam", Amount: 5000, Term: 6, InterestRate: 4.2, Status: "Approved"},
	}

	for _, loan := range loans {
		loanJSON, err := json.Marshal(loan)
		if err != nil {
			return err
		}
		err = ctx.GetStub().PutState(loan.ID, loanJSON)
		if err != nil {
			return fmt.Errorf("failed to put to world state: %v", err)
		}
	}

	return nil
}

// CreateLoanApplication adds a new loan application to the ledger
func (s *SmartContract) CreateLoanApplication(ctx contractapi.TransactionContextInterface, id, applicant string, amount, term int, interestRate float64) error {
	exists, err := s.LoanExists(ctx, id)
	if err != nil {
		return err
	}
	if exists {
		return fmt.Errorf("the loan application %s already exists", id)
	}

	loan := LoanApplication{
		ID:           id,
		Applicant:    applicant,
		Amount:       amount,
		Term:         term,
		InterestRate: interestRate,
		Status:       "Pending",
	}

	loanJSON, err := json.Marshal(loan)
	if err != nil {
		return err
	}

	return ctx.GetStub().PutState(id, loanJSON)
}

// ReadLoanApplication returns the loan application by ID
func (s *SmartContract) ReadLoanApplication(ctx contractapi.TransactionContextInterface, id string) (*LoanApplication, error) {
	loanJSON, err := ctx.GetStub().GetState(id)
	if err != nil {
		return nil, fmt.Errorf("failed to read from world state: %v", err)
	}
	if loanJSON == nil {
		return nil, fmt.Errorf("the loan application %s does not exist", id)
	}

	var loan LoanApplication
	err = json.Unmarshal(loanJSON, &loan)
	if err != nil {
		return nil, err
	}

	return &loan, nil
}

// UpdateLoanStatus changes the status of an existing loan application
func (s *SmartContract) UpdateLoanStatus(ctx contractapi.TransactionContextInterface, id, newStatus string) error {
	loan, err := s.ReadLoanApplication(ctx, id)
	if err != nil {
		return err
	}

	loan.Status = newStatus

	loanJSON, err := json.Marshal(loan)
	if err != nil {
		return err
	}

	return ctx.GetStub().PutState(id, loanJSON)
}

// DeleteLoanApplication removes a loan application from the ledger
func (s *SmartContract) DeleteLoanApplication(ctx contractapi.TransactionContextInterface, id string) error {
	exists, err := s.LoanExists(ctx, id)
	if err != nil {
		return err
	}
	if !exists {
		return fmt.Errorf("the loan application %s does not exist", id)
	}

	return ctx.GetStub().DelState(id)
}

// GetAllLoanApplications lists all loan applications in the ledger
func (s *SmartContract) GetAllLoanApplications(ctx contractapi.TransactionContextInterface) ([]*LoanApplication, error) {
	resultsIterator, err := ctx.GetStub().GetStateByRange("", "")
	if err != nil {
		return nil, err
	}
	defer resultsIterator.Close()

	var loans []*LoanApplication
	for resultsIterator.HasNext() {
		queryResponse, err := resultsIterator.Next()
		if err != nil {
			return nil, err
		}

		var loan LoanApplication
		err = json.Unmarshal(queryResponse.Value, &loan)
		if err != nil {
			return nil, err
		}
		loans = append(loans, &loan)
	}

	return loans, nil
}

// LoanExists checks if a loan with the given ID exists
func (s *SmartContract) LoanExists(ctx contractapi.TransactionContextInterface, id string) (bool, error) {
	loanJSON, err := ctx.GetStub().GetState(id)
	if err != nil {
		return false, err
	}
	return loanJSON != nil, nil
}

func main() {
	chaincode, err := contractapi.NewChaincode(&SmartContract{})
	if err != nil {
		fmt.Printf("Error creating loan application chaincode: %v\n", err)
		return
	}

	if err := chaincode.Start(); err != nil {
		fmt.Printf("Error starting loan application chaincode: %v\n", err)
	}
}
