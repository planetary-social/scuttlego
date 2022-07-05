package rpc

import (
	"encoding/json"
	"fmt"

	"github.com/boreq/errors"
	"github.com/planetary-social/scuttlego/service/domain/transport/rpc/transport"
)

type RequestBody struct {
	Name []string        `json:"name"`
	Type string          `json:"type"`
	Args json.RawMessage `json:"args"`
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

func unmarshalRequest(msg *transport.Message) (*Request, error) {
	var requestBody RequestBody
	if err := json.Unmarshal(msg.Body, &requestBody); err != nil {
		return nil, errors.Wrap(err, "could not unmarshal the request body")
	}

	procedureName, err := NewProcedureName(requestBody.Name)
	if err != nil {
		return nil, errors.Wrap(err, "could not create a procedure name")
	}

	procedureType := decodeProcedureType(requestBody.Type)

	req, err := NewRequest(
		procedureName,
		procedureType,
		requestBody.Args,
	)
	if err != nil {
		return nil, errors.Wrap(err, "could not create a request")
	}

	return req, err
}

func marshalRequest(req *Request, requestNumber uint32) (*transport.Message, error) {
	j, err := MarshalRequestBody(req)
	if err != nil {
		return nil, errors.Wrap(err, "could not marshal the request body")
	}

	flags, err := transport.NewMessageHeaderFlags(guessStream(req.Type()), false, transport.MessageBodyTypeJSON)
	if err != nil {
		return nil, errors.Wrap(err, "could not create message header flags")
	}

	header, err := transport.NewMessageHeader(
		flags,
		uint32(len(j)),
		int32(requestNumber),
	)
	if err != nil {
		return nil, errors.Wrap(err, "could not create a message header")
	}

	msg, err := transport.NewMessage(header, j)
	if err != nil {
		return nil, errors.Wrap(err, "could not create a message")
	}

	return &msg, nil
}

func guessStream(procedureType ProcedureType) bool {
	switch procedureType {
	case ProcedureTypeDuplex:
		return true
	case ProcedureTypeSource:
		return true
	default:
		return false
	}
}

// sendCloseStream closes the specified stream. Number is put directly into the
// request number field of the sent message. If you are closing a stream that
// you initiated then number will be a positive value, if you are closing a
// stream that a peer initiated then number will be a negative value.
func sendCloseStream(raw MessageSender, number int, errToSent error) error {
	// todo do this correctly? are the flags correct?
	flags, err := transport.NewMessageHeaderFlags(true, true, transport.MessageBodyTypeJSON)
	if err != nil {
		return errors.Wrap(err, "could not create message header flags")
	}

	var content []byte
	if errToSent == nil {
		content = []byte("true") // todo why true, is there any reason for this? do we have to send something specific? is this documented?
	} else {
		var mErr error
		content, mErr = json.Marshal(struct {
			Error string `json:"error"`
		}{errToSent.Error()})
		if mErr != nil {
			panic(mErr) // tests would have caught this eg. TestPrematureTerminationByRemote
		}
	}

	header, err := transport.NewMessageHeader(flags, uint32(len(content)), int32(number))
	if err != nil {
		return errors.Wrap(err, "could not create a message header")
	}

	msg, err := transport.NewMessage(header, content)
	if err != nil {
		return errors.Wrap(err, "could not create a message")
	}

	if err := raw.Send(&msg); err != nil {
		return errors.Wrap(err, "could not send a message")
	}

	return nil
}
