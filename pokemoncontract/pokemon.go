package main

import (
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/hyperledger/fabric-contract-api-go/contractapi"
	"github.com/hyperledger/fabric-chaincode-go/shim"
)

// SmartContract provides functions for managing Pokemon
type SmartContract struct {
	contractapi.Contract
}

// Pokemon defines the structure for a Pokemon asset
type Pokemon struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Type      string `json:"type"`
	Power     int    `json:"power"`
	Trainer   string `json:"trainer"`
	Evolved   bool   `json:"evolved"`
	Location  string `json:"location"`
}

// InitLedger adds initial Pokemons to the ledger
func (s *SmartContract) InitLedger(ctx contractapi.TransactionContextInterface) error {
	pokemons := []Pokemon{
		{ID: "poke1", Name: "Pikachu", Type: "Electric", Power: 55, Trainer: "Ash", Evolved: false, Location: "Pallet Town"},
		{ID: "poke2", Name: "Charmander", Type: "Fire", Power: 52, Trainer: "Red", Evolved: false, Location: "Cinnabar Island"},
		{ID: "poke3", Name: "Squirtle", Type: "Water", Power: 48, Trainer: "Misty", Evolved: false, Location: "Cerulean City"},
	}

	for _, p := range pokemons {
		pokeJSON, err := json.Marshal(p)
		if err != nil {
			return err
		}

		err = ctx.GetStub().PutState(p.ID, pokeJSON)
		if err != nil {
			return err
		}
	}
	return nil
}

// CreatePokemon adds a new Pokemon to the ledger
func (s *SmartContract) CreatePokemon(ctx contractapi.TransactionContextInterface, id, name, ptype, trainer, location string, power int) error {
	exists, err := s.PokemonExists(ctx, id)
	if err != nil {
		return err
	}
	if exists {
		return fmt.Errorf("Pokemon %s already exists", id)
	}

	p := Pokemon{
		ID:       id,
		Name:     name,
		Type:     ptype,
		Power:    power,
		Trainer:  trainer,
		Evolved:  false,
		Location: location,
	}

	pokeJSON, err := json.Marshal(p)
	if err != nil {
		return err
	}

	return ctx.GetStub().PutState(id, pokeJSON)
}

// ReadPokemon returns the Pokemon from the ledger
func (s *SmartContract) ReadPokemon(ctx contractapi.TransactionContextInterface, id string) (*Pokemon, error) {
	pokeJSON, err := ctx.GetStub().GetState(id)
	if err != nil {
		return nil, fmt.Errorf("failed to read from world state: %v", err)
	}
	if pokeJSON == nil {
		return nil, fmt.Errorf("Pokemon %s does not exist", id)
	}

	var poke Pokemon
	err = json.Unmarshal(pokeJSON, &poke)
	if err != nil {
		return nil, err
	}
	return &poke, nil
}

// UpdatePokemon modifies the power or trainer of a Pokemon
func (s *SmartContract) UpdatePokemon(ctx contractapi.TransactionContextInterface, id, trainer string, power int) error {
	p, err := s.ReadPokemon(ctx, id)
	if err != nil {
		return err
	}

	p.Power = power
	p.Trainer = trainer

	pokeJSON, err := json.Marshal(p)
	if err != nil {
		return err
	}

	return ctx.GetStub().PutState(id, pokeJSON)
}

// EvolvePokemon marks a Pokemon as evolved
func (s *SmartContract) EvolvePokemon(ctx contractapi.TransactionContextInterface, id string) error {
	p, err := s.ReadPokemon(ctx, id)
	if err != nil {
		return err
	}

	if p.Evolved {
		return fmt.Errorf("Pokemon %s is already evolved", id)
	}

	p.Evolved = true
	p.Power += 30 // bonus power on evolution

	pokeJSON, err := json.Marshal(p)
	if err != nil {
		return err
	}

	return ctx.GetStub().PutState(id, pokeJSON)
}

// DeletePokemon removes a Pokemon from ledger
func (s *SmartContract) DeletePokemon(ctx contractapi.TransactionContextInterface, id string) error {
	exists, err := s.PokemonExists(ctx, id)
	if err != nil {
		return err
	}
	if !exists {
		return fmt.Errorf("Pokemon %s does not exist", id)
	}

	return ctx.GetStub().DelState(id)
}

func (s *SmartContract) GetHistory(ctx contractapi.TransactionContextInterface, id string) ([]string, error) {
	resultsIterator, err := ctx.GetStub().GetHistoryForKey(id)
	if err != nil {
		return nil, err
	}
	defer resultsIterator.Close()

	var history []string

	for resultsIterator.HasNext() {
		resp, err := resultsIterator.Next()
		if err != nil {
			return nil, err
		}
		var tx string
		if resp.IsDelete {
			tx = fmt.Sprintf("Deleted at TxID: %s", resp.TxId)
		} else {
			tx = fmt.Sprintf("TxID: %s, Data: %s", resp.TxId, string(resp.Value))
		}
		history = append(history, tx)
	}

	return history, nil
}

func (s *SmartContract) PokemonExists(ctx contractapi.TransactionContextInterface, id string) (bool, error) {
	pokeJSON, err := ctx.GetStub().GetState(id)
	if err != nil {
		return false, err
	}
	return pokeJSON != nil, nil
}

func main() {
	cc, err := contractapi.NewChaincode(new(SmartContract))
	if err != nil {
		panic(fmt.Sprintf("Error creating Pokemon chaincode: %v", err))
	}

	if err := cc.Start(); err != nil {
		panic(fmt.Sprintf("Error starting chaincode: %v", err))
	}
}

