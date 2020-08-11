package lsp

import (
	"bufio"
	"encoding/json"
	"io"
	"log"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"sync"

	"github.com/sourcegraph/go-lsp"
	"github.com/zyedidia/micro/v2/internal/util"
)

var activeServers map[string]*Server
var slock sync.Mutex

func init() {
	activeServers = make(map[string]*Server)
}

type Server struct {
	cmd          *exec.Cmd
	stdin        io.WriteCloser
	stdout       *bufio.Reader
	language     *Language
	capabilities lsp.ServerCapabilities
	root         string
	lock         sync.Mutex
	active       bool
	requestID    int
	responses    map[int]chan ([]byte)
}

type RPCRequest struct {
	RPCVersion string      `json:"jsonrpc"`
	ID         int         `json:"id"`
	Method     string      `json:"method"`
	Params     interface{} `json:"params"`
}

type RPCNotification struct {
	RPCVersion string      `json:"jsonrpc"`
	Method     string      `json:"method"`
	Params     interface{} `json:"params"`
}

type RPCInit struct {
	RPCVersion string               `json:"jsonrpc"`
	ID         int                  `json:"id"`
	Result     lsp.InitializeResult `json:"result"`
}

type RPCResult struct {
	RPCVersion string `json:"jsonrpc"`
	ID         int    `json:"id,omitempty"`
	Method     string `json:"method,omitempty"`
}

func StartServer(l Language) (*Server, error) {
	c := exec.Command(l.Command, l.Args...)

	c.Stderr = log.Writer()

	stdin, err := c.StdinPipe()
	if err != nil {
		log.Println("[micro-lsp]", err)
		return nil, err
	}

	stdout, err := c.StdoutPipe()
	if err != nil {
		log.Println("[micro-lsp]", err)
		return nil, err
	}

	err = c.Start()
	if err != nil {
		log.Println("[micro-lsp]", err)
		return nil, err
	}

	s := new(Server)
	s.cmd = c
	s.stdin = stdin
	s.stdout = bufio.NewReader(stdout)
	s.language = &l
	s.responses = make(map[int]chan []byte)

	// activeServers[l.Command] = s

	return s, nil
}

// Initialize performs the LSP initialization handshake
// The directory must be an absolute path
func (s *Server) Initialize(directory string) {
	params := lsp.InitializeParams{
		ProcessID: os.Getpid(),
		RootURI:   lsp.DocumentURI("file://" + directory),
		ClientInfo: lsp.ClientInfo{
			Name:    "micro",
			Version: util.Version,
		},
		Trace: "off",
		Capabilities: lsp.ClientCapabilities{
			Workspace: lsp.WorkspaceClientCapabilities{
				WorkspaceEdit: struct {
					DocumentChanges    bool     `json:"documentChanges,omitempty"`
					ResourceOperations []string `json:"resourceOperations,omitempty"`
				}{
					DocumentChanges:    true,
					ResourceOperations: []string{"create", "rename", "delete"},
				},
				ApplyEdit: true,
			},
			TextDocument: lsp.TextDocumentClientCapabilities{
				Formatting: &struct {
					DynamicRegistration bool `json:"dynamicRegistration,omitempty"`
				}{
					DynamicRegistration: true,
				},
				Completion: struct {
					CompletionItem struct {
						DocumentationFormat []lsp.DocumentationFormat `json:"documentationFormat,omitempty"`
						SnippetSupport      bool                      `json:"snippetSupport,omitempty"`
					} `json:"completionItem,omitempty"`

					CompletionItemKind struct {
						ValueSet []lsp.CompletionItemKind `json:"valueSet,omitempty"`
					} `json:"completionItemKind,omitempty"`

					ContextSupport bool `json:"contextSupport,omitempty"`
				}{
					CompletionItem: struct {
						DocumentationFormat []lsp.DocumentationFormat `json:"documentationFormat,omitempty"`
						SnippetSupport      bool                      `json:"snippetSupport,omitempty"`
					}{
						DocumentationFormat: []lsp.DocumentationFormat{lsp.DFPlainText},
						SnippetSupport:      false,
					},
					ContextSupport: false,
				},
			},
			Window: lsp.WindowClientCapabilities{
				WorkDoneProgress: false,
			},
			Experimental: nil,
		},
	}

	activeServers[s.language.Command+"-"+directory] = s
	s.active = true

	go s.receive()

	resp, err := s.sendRequest("initialize", params)
	if err != nil {
		log.Println("[micro-lsp]", err)
		return
	}

	// todo parse capabilities
	log.Println("[micro-lsp] <<<", string(resp))

	var r RPCInit
	json.Unmarshal(resp, &r)

	err = s.sendNotification("initialized", struct{}{})
	if err != nil {
		log.Println("[micro-lsp]", err)
		return
	}

	s.capabilities = r.Result.Capabilities
	s.root = directory
}

func (s *Server) receive() {
	for s.active {
		resp, err := s.receiveMessage()
		if err != nil {
			log.Println("[micro-lsp]", err)
			continue
		}
		log.Println("[micro-lsp] <<<", string(resp))

		var r RPCResult
		err = json.Unmarshal(resp, &r)
		if err != nil {
			log.Println("[micro-lsp]", err)
			continue
		}

		switch r.Method {
		case "window/logMessage":
			// TODO
		case "textDocument/publishDiagnostics":
			// TODO
		case "":
			// Response
			if _, ok := s.responses[r.ID]; ok {
				s.responses[r.ID] <- resp
			}
		}
	}
}

func (s *Server) receiveMessage() ([]byte, error) {
	n := -1
	for {
		b, err := s.stdout.ReadBytes('\n')
		if err != nil {
			return nil, err
		}
		headerline := strings.TrimSpace(string(b))
		if len(headerline) == 0 {
			break
		}
		if strings.HasPrefix(headerline, "Content-Length:") {
			split := strings.Split(headerline, ":")
			if len(split) <= 1 {
				break
			}
			n, err = strconv.Atoi(strings.TrimSpace(split[1]))
			if err != nil {
				return nil, err
			}
		}
	}

	if n <= 0 {
		return []byte{}, nil
	}

	bytes := make([]byte, n)
	_, err := io.ReadFull(s.stdout, bytes)
	if err != nil {
		log.Println("[micro-lsp]", err)
	}
	return bytes, err
}

func (s *Server) sendNotification(method string, params interface{}) error {
	m := RPCNotification{
		RPCVersion: "2.0",
		Method:     method,
		Params:     params,
	}

	return s.sendMessage(m)
}

func (s *Server) sendRequest(method string, params interface{}) ([]byte, error) {
	id := s.requestID
	s.requestID++
	r := make(chan []byte)
	s.responses[id] = r

	m := RPCRequest{
		RPCVersion: "2.0",
		ID:         id,
		Method:     method,
		Params:     params,
	}

	err := s.sendMessage(m)
	if err != nil {
		return nil, err
	}

	bytes := <-r
	delete(s.responses, id)

	return bytes, nil
}

func (s *Server) sendMessage(m interface{}) error {
	msg, err := json.Marshal(m)
	if err != nil {
		return err
	}

	log.Println("[micro-lsp] >>>", string(msg))

	// encode header and proper line endings
	msg = append(msg, '\r', '\n')
	header := []byte("Content-Length: " + strconv.Itoa(len(msg)) + "\r\n\r\n")
	msg = append(header, msg...)

	_, err = s.stdin.Write(msg)
	return err
}
