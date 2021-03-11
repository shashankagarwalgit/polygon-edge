package jsonrpc

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
	"unicode"

	"github.com/hashicorp/go-hclog"
)

var (
	invalidJSONRequest = &ErrorObject{Code: -32600, Message: "invalid json request"}
	internalError      = &ErrorObject{Code: -32603, Message: "internal error"}
)

func invalidMethod(method string) error {
	return &ErrorObject{Code: -32601, Message: fmt.Sprintf("The method %s does not exist/is not available", method)}
}

func invalidArguments(method string) error {
	return &ErrorObject{Code: -32602, Message: fmt.Sprintf("invalid arguments to %s", method)}
}

type serviceData struct {
	sv      reflect.Value
	funcMap map[string]*funcData
}

type funcData struct {
	inNum int
	reqt  []reflect.Type
	fv    reflect.Value
}

type endpoints struct {
	Eth  *Eth
	Web3 *Web3
	Net  *Net
}

type enabledEndpoints map[string]struct{}

// Dispatcher handles jsonrpc requests
type Dispatcher struct {
	logger        hclog.Logger
	store         blockchainInterface
	serviceMap    map[string]*serviceData
	endpoints     endpoints
	filterManager *FilterManager
}

func newDispatcher(logger hclog.Logger, store blockchainInterface) *Dispatcher {
	d := &Dispatcher{
		logger: logger.Named("dispatcher"),
	}
	d.registerEndpoints()
	if store != nil {
		d.filterManager = NewFilterManager(logger, store)
		go d.filterManager.Run()
	}
	return d
}

func (d *Dispatcher) registerEndpoints() {
	d.endpoints.Eth = &Eth{d}
	d.endpoints.Net = &Net{d}
	d.endpoints.Web3 = &Web3{d}

	d.registerService("eth", d.endpoints.Eth)
	d.registerService("net", d.endpoints.Net)
	d.registerService("web3", d.endpoints.Web3)
}

func (d *Dispatcher) getFnHandler(typ serverType, req Request, params int) (*serviceData, *funcData, error) {
	callName := strings.SplitN(req.Method, "_", 2)
	if len(callName) != 2 {
		return nil, nil, invalidMethod(req.Method)
	}

	serviceName, funcName := callName[0], callName[1]

	service, ok := d.serviceMap[serviceName]
	if !ok {
		return nil, nil, invalidMethod(req.Method)
	}
	fd, ok := service.funcMap[funcName]
	if !ok {
		return nil, nil, invalidMethod(req.Method)
	}
	if params != fd.inNum-1 {
		return nil, nil, invalidArguments(req.Method)
	}
	return service, fd, nil
}

type wsConn interface {
	WriteMessage(b []byte) error
}

func (d *Dispatcher) handleSubscribe(req Request, conn wsConn) (string, error) {
	var params []interface{}
	if err := json.Unmarshal(req.Params, &params); err != nil {
		return "", invalidJSONRequest
	}
	if len(params) == 0 {
		return "", invalidJSONRequest
	}

	subscribeMethod, ok := params[0].(string)
	if !ok {
		return "", fmt.Errorf("subscribe method '%s' not found", params[0])
	}

	var filterID string
	if subscribeMethod == "newHeads" {
		filterID = d.filterManager.NewBlockFilter(conn)

	} else if subscribeMethod == "logs" {
		logFilter, err := decodeLogFilterFromInterface(params[1])
		if err != nil {
			return "", err
		}
		filterID = d.filterManager.NewLogFilter(logFilter, conn)

	} else {
		return "", fmt.Errorf("subscribe method %s not found", subscribeMethod)
	}

	return filterID, nil
}

func (d *Dispatcher) handleUnsubscribe(req Request) (bool, error) {
	var params []interface{}
	if err := json.Unmarshal(req.Params, &params); err != nil {
		return false, invalidJSONRequest
	}
	if len(params) != 1 {
		return false, invalidJSONRequest
	}

	filterID, ok := params[0].(string)
	if !ok {
		return false, fmt.Errorf("unsubscribe filter not found")
	}

	return d.filterManager.Uninstall(filterID), nil
}

func (d *Dispatcher) HandleWs(reqBody []byte, conn wsConn) ([]byte, error) {
	var req Request
	if err := json.Unmarshal(reqBody, &req); err != nil {
		return nil, invalidJSONRequest
	}

	// if the request method is eth_subscribe we need to create a
	// new filter with ws connection
	if req.Method == "eth_subscribe" {
		filterID, err := d.handleSubscribe(req, conn)
		if err != nil {
			return nil, err
		}

		resp := fmt.Sprintf(`{"jsonrpc":"2.0","id":%d,"result":"%s"}`, req.ID, filterID)
		return []byte(resp), nil
	}

	if req.Method == "eth_unsubscribe" {
		ok, err := d.handleUnsubscribe(req)
		if err != nil {
			return nil, err
		}

		res := "false"
		if ok {
			res = "true"
		}
		resp := fmt.Sprintf(`{"jsonrpc":"2.0","id":%d,"result":"%s"}`, req.ID, res)
		return []byte(resp), nil
	}

	// its a normal query that we handle with the dispatcher
	resp, err := d.handleReq(serverWS, req)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

func (d *Dispatcher) Handle(typ serverType, reqBody []byte) ([]byte, error) {
	var req Request
	if err := json.Unmarshal(reqBody, &req); err != nil {
		return nil, invalidJSONRequest
	}
	return d.handleReq(typ, req)
}

func (d *Dispatcher) handleReq(typ serverType, req Request) ([]byte, error) {
	d.logger.Debug("request", "method", req.Method, "id", req.ID, "typ", typ)

	var params []interface{}
	if err := json.Unmarshal(req.Params, &params); err != nil {
		return nil, invalidJSONRequest
	}

	service, fd, err := d.getFnHandler(typ, req, len(params))
	if err != nil {
		return nil, err
	}

	inArgs := make([]reflect.Value, fd.inNum)
	inArgs[0] = service.sv

	for i := 0; i < fd.inNum-1; i++ {
		elem := reflect.ValueOf(params[i])
		if elem.Type() != fd.reqt[i+1] {
			return nil, invalidArguments(req.Method)
		}
		inArgs[i+1] = elem
	}

	output := fd.fv.Call(inArgs)
	err = getError(output[1])
	if err != nil {
		return nil, internalError
	}

	var data []byte
	res := output[0].Interface()
	if res != nil {
		data, err = json.Marshal(res)
		if err != nil {
			return nil, internalError
		}
	}

	resp := Response{
		ID:     req.ID,
		Result: data,
	}
	respBytes, err := json.Marshal(resp)
	if err != nil {
		return nil, internalError
	}
	return respBytes, nil
}

func (d *Dispatcher) registerService(serviceName string, service interface{}) {
	if d.serviceMap == nil {
		d.serviceMap = map[string]*serviceData{}
	}
	if serviceName == "" {
		panic(fmt.Sprintf("jsonrpc: serviceName cannot be empty"))
	}

	st := reflect.TypeOf(service)
	if st.Kind() == reflect.Struct {
		panic(fmt.Sprintf("jsonrpc: service '%s' must be a pointer to struct", serviceName))
	}

	funcMap := make(map[string]*funcData)
	for i := 0; i < st.NumMethod(); i++ {
		mv := st.Method(i)
		if mv.PkgPath != "" {
			// skip unexported methods
			continue
		}

		name := lowerCaseFirst(mv.Name)
		funcName := serviceName + "_" + name
		fd := &funcData{
			fv: mv.Func,
		}
		var err error
		if fd.inNum, fd.reqt, err = validateFunc(funcName, fd.fv, true); err != nil {
			panic(fmt.Sprintf("jsonrpc: %s", err))
		}
		funcMap[name] = fd
	}

	d.serviceMap[serviceName] = &serviceData{
		sv:      reflect.ValueOf(service),
		funcMap: funcMap,
	}
}

func removePtr(t reflect.Type) reflect.Type {
	for t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	return t
}

func validateFunc(funcName string, fv reflect.Value, isMethod bool) (inNum int, reqt []reflect.Type, err error) {
	if funcName == "" {
		err = fmt.Errorf("funcName cannot be empty")
		return
	}

	ft := fv.Type()
	if ft.Kind() != reflect.Func {
		err = fmt.Errorf("function '%s' must be a function instead of %s", funcName, ft)
		return
	}

	inNum = ft.NumIn()
	outNum := ft.NumOut()

	if outNum != 2 {
		err = fmt.Errorf("unexpected number of output arguments in the function '%s': %d. Expected 2", funcName, outNum)
		return
	}
	if !isErrorType(ft.Out(1)) {
		err = fmt.Errorf("unexpected type for the second return value of the function '%s': '%s'. Expected '%s'", funcName, ft.Out(1), errt)
		return
	}

	reqt = make([]reflect.Type, inNum)
	for i := 0; i < inNum; i++ {
		reqt[i] = ft.In(i)
	}
	return
}

var errt = reflect.TypeOf((*error)(nil)).Elem()

func isErrorType(t reflect.Type) bool {
	return t.Implements(errt)
}

func (d *Dispatcher) funcExample(b string) (interface{}, error) {
	return nil, nil
}

func getError(v reflect.Value) error {
	if v.IsNil() {
		return nil
	}
	return v.Interface().(error)
}

func lowerCaseFirst(str string) string {
	for i, v := range str {
		return string(unicode.ToLower(v)) + str[i+1:]
	}
	return ""
}
