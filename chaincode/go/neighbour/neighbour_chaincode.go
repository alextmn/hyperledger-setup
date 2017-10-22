
package main

import (
	"fmt"

	"github.com/hyperledger/fabric/core/chaincode/shim"
	pb "github.com/hyperledger/fabric/protos/peer"
	"strings"
	"encoding/pem"
	"crypto/x509"
	"encoding/json"
)

type NeighbourChaincode struct {
}

type Lease struct {
	Key		LeaseKey 	`json:"key"`
	Value	LeaseValue 	`json:"value"`
}

type LeaseKey struct {
	Season	string 		`json:"season"`
	Lot		string 		`json:"lot"`
}

type LeaseValue struct {
	Leasor	string 		`json:"leasor"`
	Leasee	string 		`json:"leasee"`
	Terms	string 		`json:"terms"`
}

var logger = shim.NewLogger("NeighbourChaincode")

func (t *NeighbourChaincode) Init(stub shim.ChaincodeStubInterface) pb.Response {
	logger.Debug("Init")

	return shim.Success(nil)
}

func (t *NeighbourChaincode) Invoke(stub shim.ChaincodeStubInterface) pb.Response {
	logger.Debug("Invoke")

	function, args := stub.GetFunctionAndParameters()
	if function == "addLease" {
		return t.addLease(stub, args)
	} else if function == "amendLease" {
		return t.amendLease(stub, args)
	} else if function == "signLease" {
		return t.signLease(stub, args)
	} else if function == "query" {
		return t.query(stub, args)
	} else if function == "queryLease" {
		return t.queryLease(stub, args)
	}

	return pb.Response{Status:400, Message:"Invalid invoke function name"}
}

func (t *NeighbourChaincode) addLease(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	if len(args) != 3 {
		return pb.Response{Status:400, Message:"Incorrect number of arguments"}
	}

	season := args[0]
	lot := args[1]
	terms := args[2]

	ck, err := stub.CreateCompositeKey("Lease", []string{season, lot})
	if err != nil {
		return shim.Error(fmt.Sprintf("cannot CreateCompositeKey %v", err))
	}

	creatorBytes, err := stub.GetCreator()
	if err != nil {
		return shim.Error(fmt.Sprintf("cannot getCreator %v", err))
	}

	commonName, org := getCreator(creatorBytes)
	leasor := commonName + "@" + org

	leaseValue := LeaseValue{Leasor: leasor, Terms:terms}

	leaseValueBytes, err := json.Marshal(leaseValue)
	if err != nil {
		return shim.Error(fmt.Sprintf("cannot Marshal %v", err))
	}

	//TODO check with reference the lot is not leased

	err = stub.PutState(ck, leaseValueBytes)
	if err != nil {
		return shim.Error(fmt.Sprintf("cannot PutState %v", err))
	}

	leaseKeyBytes, err := json.Marshal(LeaseKey{Lot:lot, Season:season})
	if err != nil {
		return shim.Error(fmt.Sprintf("cannot Marshal %v", err))
	}

	return shim.Success(leaseKeyBytes)
}

func (t *NeighbourChaincode) amendLease(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	if len(args) != 3 {
		return pb.Response{Status:400, Message:"Incorrect number of arguments"}
	}

	season := args[0]
	lot := args[1]
	terms := args[2]

	ck, err := stub.CreateCompositeKey("Lease", []string{season, lot})
	if err != nil {
		return shim.Error(fmt.Sprintf("cannot CreateCompositeKey %v", err))
	}

	leaseValueBytes, err := stub.GetState(ck)
	if err != nil {
		return shim.Error(fmt.Sprintf("cannot GetState %v", err))
	}

	var leaseValue LeaseValue
	err = json.Unmarshal(leaseValueBytes, &leaseValue)
	if err != nil {
		return shim.Error(fmt.Sprintf("cannot Unmarshal %v", err))
	}

	if leaseValue.Leasee != "" {
		return pb.Response{Status:409, Message:"cannot amend signed lease"}
	}

	leaseValue.Terms = terms

	leaseValueBytes, err = json.Marshal(leaseValue)
	if err != nil {
		return shim.Error(fmt.Sprintf("cannot Marshal %v", err))
	}

	err = stub.PutState(ck, leaseValueBytes)
	if err != nil {
		return shim.Error(fmt.Sprintf("cannot PutState %v", err))
	}

	return shim.Success(nil)
}

func (t *NeighbourChaincode) signLease(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	if len(args) != 2 {
		return pb.Response{Status:400, Message:"Incorrect number of arguments"}
	}

	season := args[0]
	lot := args[1]

	ck, err := stub.CreateCompositeKey("Lease", []string{season, lot})
	if err != nil {
		return shim.Error(fmt.Sprintf("cannot CreateCompositeKey %v", err))
	}

	leaseValueBytes, err := stub.GetState(ck)
	if err != nil {
		return shim.Error(fmt.Sprintf("cannot GetState %v", err))
	}

	var leaseValue LeaseValue
	err = json.Unmarshal(leaseValueBytes, &leaseValue)
	if err != nil {
		return shim.Error(fmt.Sprintf("cannot Unmarshal %v", err))
	}

	creatorBytes, err := stub.GetCreator()
	if err != nil {
		return shim.Error(fmt.Sprintf("cannot getCreator %v", err))
	}

	commonName, org := getCreator(creatorBytes)
	leasee := commonName + "@" + org

	if leaseValue.Leasor == leasee {
		return pb.Response{Status:409, Message:"leasor cannot sign lease"}
	}

	leaseValue.Leasee = leasee

	leaseValueBytes, err = json.Marshal(leaseValue)
	if err != nil {
		return shim.Error(fmt.Sprintf("cannot Marshal %v", err))
	}

	err = stub.PutState(ck, leaseValueBytes)
	if err != nil {
		return shim.Error(fmt.Sprintf("cannot PutState %v", err))
	}

	return shim.Success(nil)
}

func (t *NeighbourChaincode) queryLease(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	if len(args) != 2 {
		return pb.Response{Status:400, Message:"Incorrect number of arguments"}
	}

	season := args[0]
	lot := args[1]

	ck, err := stub.CreateCompositeKey("Lease", []string{season, lot})
	if err != nil {
		return shim.Error(fmt.Sprintf("cannot CreateCompositeKey %v", err))
	}

	leaseValueBytes, err := stub.GetState(ck)
	if err != nil {
		return shim.Error(fmt.Sprintf("cannot GetState %v", err))
	}

	return shim.Success(leaseValueBytes)
}

func (t *NeighbourChaincode) query(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	var keys []string

	if len(args) > 1 {
		return pb.Response{Status:400, Message:"Incorrect number of arguments"}
	} else if len(args) == 1 {
		season := args[0]
		keys = []string{season}
	}

	it, err := stub.GetStateByPartialCompositeKey("Lease", keys)
	if err != nil {
		return shim.Error(err.Error())
	}
	defer it.Close()

	arr := []Lease{}
	for it.HasNext() {
		next, err := it.Next()
		if err != nil {
			return shim.Error(err.Error())
		}

		var leaseValue LeaseValue
		err = json.Unmarshal(next.Value, &leaseValue)
		if err != nil {
			return shim.Error(err.Error())
		}

		_, keys, err := stub.SplitCompositeKey(next.Key)
		if err != nil {
			return shim.Error(err.Error())
		}

		leaseKey := LeaseKey{Season: keys[0], Lot: keys[1]}

		lease := Lease{Key: leaseKey, Value: leaseValue}

		arr = append(arr, lease)
	}

	ret, err := json.Marshal(arr)
	if err != nil {
		return shim.Error(err.Error())
	}

	return shim.Success(ret)
}

var getCreator = func (certificate []byte) (string, string) {
	data := certificate[strings.Index(string(certificate), "-----"): strings.LastIndex(string(certificate), "-----")+5]
	block, _ := pem.Decode([]byte(data))
	cert, _ := x509.ParseCertificate(block.Bytes)
	organization := cert.Issuer.Organization[0]
	commonName := cert.Subject.CommonName
	logger.Debug("commonName: " + commonName + ", organization: " + organization)

	organizationShort := strings.Split(organization, ".")[0]

	return commonName, organizationShort
}

func main() {
	err := shim.Start(new(NeighbourChaincode))
	if err != nil {
		fmt.Printf("Error starting Simple chaincode: %s", err)
	}
}
