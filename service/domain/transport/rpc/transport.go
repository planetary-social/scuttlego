package rpc

import (
	"encoding/json"
	"fmt"

	"github.com/boreq/errors"
)

type RequestBody struct {
	Name []string        `json:"name"`
	Type string          `json:"type"`
	Args json.RawMessage `json:"args"`
}

const (
	transportStringForProcedureTypeSource = "source"
	transportStringForProcedureTypeDuplex = "duplex"
	transportStringForProcedureTypeAsync  = "async"
)

func decodeProcedureType(str string) ProcedureType {
	switch str {
	case transportStringForProcedureTypeSource:
		return ProcedureTypeSource
	case transportStringForProcedureTypeDuplex:
		return ProcedureTypeDuplex
	case transportStringForProcedureTypeAsync:
		return ProcedureTypeAsync
	default:
		return ProcedureTypeUnknown
	}
}

func encodeProcedureType(t ProcedureType) (string, error) {
	switch t {
	case ProcedureTypeSource:
		return transportStringForProcedureTypeSource, nil
	case ProcedureTypeDuplex:
		return transportStringForProcedureTypeDuplex, nil
	case ProcedureTypeAsync:
		return transportStringForProcedureTypeAsync, nil
	default:
		return "", fmt.Errorf("unknown procedure type %T", t)
	}
}

func MarshalRequestBody(req *Request) ([]byte, error) {
	encodedType, err := encodeProcedureType(req.Type())
	if err != nil {
		return nil, errors.Wrap(err, "could not encode the procedure type")
	}

	body := RequestBody{
		Name: req.Name().Components(),
		Type: encodedType,
		Args: req.Arguments(),
	}

	j, err := json.Marshal(body)
	if err != nil {
		return nil, errors.Wrap(err, "could not marshal the request body")
	}

	return j, nil
}

func MustMarshalRequestBody(req *Request) []byte {
	v, err := MarshalRequestBody(req)
	if err != nil {
		panic(err)
	}
	return v
}
