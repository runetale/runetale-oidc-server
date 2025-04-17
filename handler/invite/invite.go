package invite

import (
	"bytes"
	"io"
	"net/http"
	"sync"
	"text/template"

	grpcclient "github.com/runetale/runetale-oidc-server/grpc_client"
	"google.golang.org/grpc"
)

type Response struct {
	Status      int
	ContentType string
	Content     io.Reader
	Headers     map[string]string
}

type Action func(r *http.Request) *Response

func (response *Response) Write(rw http.ResponseWriter) {
	if response != nil {
		if response.ContentType != "" {
			rw.Header().Set("Content-Type", response.ContentType)
		}
		for k, v := range response.Headers {
			rw.Header().Set(k, v)
		}
		rw.WriteHeader(response.Status)
		_, err := io.Copy(rw, response.Content)

		if err != nil {
			rw.WriteHeader(http.StatusInternalServerError)
		}
	} else {
		rw.WriteHeader(http.StatusOK)
	}
}

type Handler interface {
	ServeHTTP(w http.ResponseWriter, r *http.Request)
}

type InviteHandler struct {
	grpcClient grpcclient.ServerClientImpl
	html       *template.Template
	mu         *sync.Mutex
}

func NewInviteHandler(conn *grpc.ClientConn, tmpl *template.Template) Handler {
	serverClient := grpcclient.NewServerClient(conn)
	return &InviteHandler{
		grpcClient: serverClient,
		mu:         &sync.Mutex{},
		html:       tmpl,
	}
}

func (h *InviteHandler) invite(r *http.Request) *Response {
	inviteCode := r.URL.Query().Get("code")
	inviter, err := h.grpcClient.GetInvitation(inviteCode)
	if err != nil {
		return responsehtml(http.StatusOK, h.html, "index.html", "", nil)
	}

	if inviteCode != inviter.InviteCode {
		return responsehtml(http.StatusOK, h.html, "index.html", "", nil)
	}

	val := map[string]string{
		"inviter":     inviter.Email,
		"invite_code": inviter.InviteCode,
	}

	h.mu.Lock()
	defer h.mu.Unlock()

	return responsehtml(http.StatusOK, h.html, "invite.html", val, nil)
}

func responsehtml(status int, t *template.Template, template string, data interface{}, headers map[string]string) *Response {
	var buf bytes.Buffer
	if err := t.ExecuteTemplate(&buf, template, data); err != nil {
		return nil
	}
	return &Response{
		Status:      status,
		ContentType: "text/html",
		Content:     &buf,
		Headers:     headers,
	}
}

func (h *InviteHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	action := Action(h.invite)
	response := action(r)
	response.Write(w)
}
