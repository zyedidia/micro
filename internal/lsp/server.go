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

	lsp "go.lsp.dev/protocol"
	"go.lsp.dev/uri"
)

var activeServers map[string]*Server
var slock sync.Mutex

func init() {
	activeServers = make(map[string]*Server)
}

func GetServer(l Language, dir string) *Server {
	return activeServers[l.Command+"-"+dir]
}

type Server struct {
	cmd          *exec.Cmd
	stdin        io.WriteCloser
	stdout       *bufio.Reader
	language     *Language
	capabilities lsp.ServerCapabilities
	root         string
	lock         sync.Mutex
	Active       bool
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
	s := new(Server)

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

	s.cmd = c
	s.stdin = stdin
	s.stdout = bufio.NewReader(stdout)
	s.language = &l
	s.responses = make(map[int]chan []byte)

	return s, nil
}

// Initialize performs the LSP initialization handshake
// The directory must be an absolute path
func (s *Server) Initialize(directory string) {
	params := lsp.InitializeParams{
		ProcessID: float64(os.Getpid()),
		RootURI:   uri.File(directory),
		Capabilities: lsp.ClientCapabilities{
			Workspace: &lsp.WorkspaceClientCapabilities{
				WorkspaceEdit: &lsp.WorkspaceClientCapabilitiesWorkspaceEdit{
					DocumentChanges:    true,
					ResourceOperations: []string{"create", "rename", "delete"},
				},
				ApplyEdit: true,
			},
			TextDocument: &lsp.TextDocumentClientCapabilities{
				Formatting: &lsp.TextDocumentClientCapabilitiesFormatting{
					DynamicRegistration: false,
				},
				Completion: &lsp.TextDocumentClientCapabilitiesCompletion{
					DynamicRegistration: false,
					CompletionItem: &lsp.TextDocumentClientCapabilitiesCompletionItem{
						SnippetSupport:          false,
						CommitCharactersSupport: false,
						DocumentationFormat:     []lsp.MarkupKind{lsp.PlainText},
						DeprecatedSupport:       false,
						PreselectSupport:        false,
					},
					ContextSupport: false,
				},
				Hover: &lsp.TextDocumentClientCapabilitiesHover{
					DynamicRegistration: false,
					ContentFormat:       []lsp.MarkupKind{lsp.PlainText},
				},
			},
		},
	}

	activeServers[s.language.Command+"-"+directory] = s
	s.Active = true
	s.root = directory

	go s.receive()

	s.lock.Lock()
	go func() {
		resp, err := s.sendRequest("initialize", params)
		if err != nil {
			log.Println("[micro-lsp]", err)
			s.Active = false
			s.lock.Unlock()
			return
		}

		// todo parse capabilities
		log.Println("[micro-lsp] <<<", string(resp))

		var r RPCInit
		json.Unmarshal(resp, &r)

		s.lock.Unlock()
		err = s.sendNotification("initialized", struct{}{})
		if err != nil {
			log.Println("[micro-lsp]", err)
		}

		s.capabilities = r.Result.Capabilities
	}()
}

func (s *Server) receive() {
	for s.Active {
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
				log.Println("[micro-lsp] Got response for", r.ID)
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

	s.lock.Lock()
	defer s.lock.Unlock()
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
