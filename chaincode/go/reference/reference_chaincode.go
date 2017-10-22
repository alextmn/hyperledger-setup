
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

type ReferenceChaincode struct {
}

type LeaseKey struct {
	Season	string 		`json:"season"`
	Lot		string 		`json:"lot"`
}

var logger = shim.NewLogger("ReferenceChaincode")

func (t *ReferenceChaincode) Init(stub shim.ChaincodeStubInterface) pb.Response {
	logger.Debug("Init")

	return shim.Success(nil)
}

func (t *ReferenceChaincode) Invoke(stub shim.ChaincodeStubInterface) pb.Response {
	logger.Debug("Invoke")

	function, args := stub.GetFunctionAndParameters()
	if function == "addLease" {
		return t.addLease(stub, args)
	} else if function == "query" {
		return t.query(stub, args)
	}

	return pb.Response{Status:400, Message:"Invalid invoke function name"}
}

func (t *ReferenceChaincode) addLease(stub shim.ChaincodeStubInterface, args []string) pb.Response {

	return shim.Success(nil)
}

func (t *ReferenceChaincode) query(stub shim.ChaincodeStubInterface, args []string) pb.Response {
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

	arr := []LeaseKey{}
	for it.HasNext() {
		next, err := it.Next()
		if err != nil {
			return shim.Error(err.Error())
		}

		_, keys, err := stub.SplitCompositeKey(next.Key)
		if err != nil {
			return shim.Error(err.Error())
		}

		leaseKey := LeaseKey{Season: keys[0], Lot: keys[1]}

		arr = append(arr, leaseKey)
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
	err := shim.Start(new(ReferenceChaincode))
	if err != nil {
		fmt.Printf("Error starting Simple chaincode: %s", err)
	}
}
