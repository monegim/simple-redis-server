package main

type RESP struct {
	Message string
	Type    RespType
}
type RespType string

const (
	RESP_SIMPLE_STRING RespType = "RESP_SIMPLE_STRING"
	RESP_ERROR         RespType = "RESP_ERROR"
	RESP_INTEGER       RespType = "RESP_INTEGER"
	RESP_BULK_STRING   RespType = "RESP_BULK_STRING"
	RESP_ARRAYS        RespType = "RESP_ARRAYS"
)

func NewRESP(msg string) *RESP {
	t := getType(msg)
	return &RESP{
		Message: msg,
		Type:    t,
	}
}

func (r *RESP) ToString() string {
	switch r.Type {
	case RESP_SIMPLE_STRING:
		return "+" + r.Message
	case RESP_ERROR:
		return "-" + r.Message
	case RESP_INTEGER:
		return ":" + r.Message
	case RESP_BULK_STRING:
		return "$" + r.Message
	case RESP_ARRAYS:
		return "*" + r.Message
	default:
		return ""
	}
}
func getType(s string) RespType {
	c := s[0]
	switch c {
	case '+':
		return RESP_SIMPLE_STRING
	case '-':
		return RESP_ERROR
	case ':':
		return RESP_INTEGER
	case '$':
		return RESP_BULK_STRING
	case '*':
		return RESP_ARRAYS
	default:
		return ""
	}
}
func (r *RESP) Command() string {
	switch r.Message {
	case "PING":
		return "+PONG"
	default:
		return ""
	}
}
