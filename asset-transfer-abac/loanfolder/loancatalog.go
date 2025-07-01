package main

import (
    "fmt"
    "encoding/json"
    "github.com/hyperledger/fabric-contract-api-go/contractapi"
)

// SmartContract provides functions for managing applications
type SmartContract struct {
    contractapi.Contract
}

// InitLedger adds initial data to the ledger
func (s *SmartContract) InitLedger(ctx contractapi.TransactionContextInterface) error {
    return nil
}

// Example function
func (s *SmartContract) Ping(ctx contractapi.TransactionContextInterface) (string, error) {
    return "Pong", nil
}

func main() {
    chaincode, err := contractapi.NewChaincode(new(SmartContract))
    if err != nil {
        fmt.Printf("Error create chaincode: %s", err.Error())
        return
    }

    if err := chaincode.Start(); err != nil {
        fmt.Printf("Error starting chaincode: %s", err.Error())
    }
}
