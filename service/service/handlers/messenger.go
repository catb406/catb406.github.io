package handlers

import (
	"encoding/json"
	"github.com/labstack/echo/v4"
	"github.com/labstack/gommon/log"
	"gopkg.in/olahol/melody.v1"
	"net/http"
	"sync"
	"time"
)

var once sync.Once
var m *melody.Melody
var sessionCollection = make(map[int64]*melody.Session)
var ctx echo.Context

type (
	Message struct {
		IdTo      int64     `json:"id_to" example:"123"`
		IdFrom    int64     `json:"id_from" example:"321"`
		Content   string    `json:"content,omitempty" example:"Hello!"`
		CreatedAt time.Time `json:"created_at,omitempty" example:"2021-11-29T00:16:01.367333+03:00"`
		Type      string    `json:"type" example:"personal"`
	}

	Request struct {
		IdTo      int64     `json:"id_to" example:"123"`
		IdFrom    int64     `json:"id_from" example:"321"`
		Type      string    `json:"type" example:"personal"`
		CreatedAt time.Time `json:"created_at,omitempty" example:"2021-11-29T00:16:01.367333+03:00"`
		Status    string    `json:"status,omitempty" example:"declined"`
		Seen      bool      `json:"seen,omitempty" example:"false"`
	}

	Dialog struct {
		IdTo   int64  `json:"id_to" example:"123"`
		IdFrom int64  `json:"id_from" example:"321"`
		Type   string `json:"type" example:"group"`
		Status string `json:"status"`
	}

	MessagesFilter struct {
		IdUsers      []int64   `json:"id_users"`
		CreatedAfter time.Time `json:"created_after,omitempty"`
	}
)

// MessengerHandler godoc
// @Summary Send message to a user
// @ID sendMessage
// @Tags Messenger
// @Failure 500 {object} ErrorResponse
// @Router /messenger/dialogs [any]
func (handler *handler) MessengerHandler(c echo.Context) error {
	once.Do(func() {
		m = melody.New()
	})
	ctx = c
	m.HandleConnect(handler.handleConnection)
	m.HandleClose(handler.handleClose)
	m.HandleMessage(handler.handleMessage)

	return m.HandleRequest(c.Response().Writer, c.Request())
}

func (msg *Message) Serialize() []byte {
	res, _ := json.Marshal(msg)
	return res
}

func (dialog *Dialog) Serialize() []byte {
	res, _ := json.Marshal(dialog)
	return res
}

func (req *Request) Serialize() []byte {
	res, _ := json.Marshal(req)
	return res
}

func (handler *handler) handleConnection(session *melody.Session) {
	idUser, err := handler.getIdFromContext(ctx)
	if err == nil {
		session.Keys = make(map[string]interface{})
		session.Keys["id_user"] = idUser
		sessionCollection[idUser] = session
	} else {
		log.Error("failed to get id user")
	}
}

func (handler *handler) handleClose(session *melody.Session, code int, msg string) error {
	delete(sessionCollection, session.Keys["id_user"].(int64))
	return ctx.JSON(code, msg)
}

func (handler *handler) handleMessage(session *melody.Session, msg []byte) {
	var message Message
	err := json.Unmarshal(msg, &message)
	if err != nil {
		log.Error(err)
		return
	}
	message.CreatedAt = time.Now()
	err = handler.messenger.AddMessage(&message)
	if err != nil {
		log.Error(err)
		return
	}
	s, ok := sessionCollection[message.IdTo]
	if ok {
		m.BroadcastMultiple(message.Serialize(), []*melody.Session{s})
	} else {
		log.Error("failed to get recipient session")
		log.Info(sessionCollection)
	}
}

// GetDialogsHandler godoc
// @Summary Get user's dialogs and requests
// @ID getDialogs
// @Tags Messenger
// @Produce json
// @Success 200 {object} []Request
// @Failure 400,401 {object} ErrorResponse
// @Router /messenger/dialogs [get]
func (handler *handler) GetDialogsHandler(c echo.Context) error {
	idUser, err := handler.getIdFromContext(c)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, ErrorResponse{err.Error()})
	}
	dialogs, err := handler.messenger.GetDialogs(idUser)
	if err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{err.Error()})
	}
	return c.JSON(http.StatusOK, dialogs)
}

// SendRequestHandler godoc
// @Summary Send friendship or joining a group training request
// @ID sendRequest
// @Tags Messenger
// @Consume json
// @Produce json
// @Param Body body Request true "The body of a request"
// @Success 200 {string} string
// @Failure 400,401,403,500 {object} ErrorResponse
// @Router /user/requests [post]
func (handler *handler) SendRequestHandler(c echo.Context) error {
	idUser, err := handler.getIdFromContext(c)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, ErrorResponse{err.Error()})
	}
	var req Request
	err = c.Bind(&req)
	if err != nil {
		log.Errorf("failed to unmarshal request: %s", err.Error())
		return c.JSON(http.StatusBadRequest, ErrorResponse{"Invalid request"})
	}
	if idUser != req.IdFrom {
		return c.JSON(http.StatusForbidden, ErrorResponse{"Cannot send request from another user"})
	}
	nullTime := time.Time{}
	if req.CreatedAt == nullTime {
		req.CreatedAt = time.Now()
	}
	err = handler.messenger.AddRequest(&req)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, ErrorResponse{err.Error()})
	}
	if s, ok := sessionCollection[req.IdTo]; ok {
		_ = m.BroadcastOthers(req.Serialize(), s)
	}
	return c.JSON(http.StatusOK, "Request has been sent")
}

// ReplyToRequestHandler godoc
// @Summary Accept or decline friendship or joining a group training request
// @ID replyToRequest
// @Tags Messenger
// @Param Body body Request true "The body of a request - status : declined/accepted"
// @Consume json
// @Produce json
// @Success 200 {string} string
// @Failure 400,401,500,403 {object} ErrorResponse
// @Router /messenger/request/reply [put]
func (handler *handler) ReplyToRequestHandler(c echo.Context) error {
	idUser, err := handler.getIdFromContext(c)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, ErrorResponse{err.Error()})
	}
	var req Request
	err = c.Bind(&req)
	if err != nil {
		log.Errorf("failed to unmarshal request: %s", err.Error())
		return c.JSON(http.StatusBadRequest, ErrorResponse{"Invalid request"})
	}
	if idUser != req.IdTo {
		return c.JSON(http.StatusForbidden, ErrorResponse{"Cannot reply to request of another user"})
	}
	res, err := handler.messenger.UpdateRequest(&req)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, ErrorResponse{err.Error()})
	}
	if s, ok := sessionCollection[req.IdTo]; ok {
		_ = m.BroadcastOthers(res.Serialize(), s)
	}
	return c.JSON(http.StatusOK, "Request has been updated")
}

// DeclinedRequestSeenHandler godoc
// @Summary Mark response to a request as seen
// @ID seenRequestReply
// @Tags Messenger
// @Param Body body Request true "The body of a request - seen : true"
// @Consume json
// @Produce json
// @Success 200 {string} string
// @Failure 400,401,403,500 {object} ErrorResponse
// @Router /messenger/request/seen [put]
func (handler *handler) DeclinedRequestSeenHandler(c echo.Context) error {
	idUser, err := handler.getIdFromContext(c)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, ErrorResponse{err.Error()})
	}
	var req Request
	err = c.Bind(&req)
	if err != nil {
		log.Errorf("failed to unmarshal request: %s", err.Error())
		return c.JSON(http.StatusBadRequest, ErrorResponse{"Invalid request"})
	}
	if idUser != req.IdFrom {
		return c.JSON(http.StatusForbidden, ErrorResponse{"Cannot modify request of another user"})
	}
	req.Seen = true
	req.Status = "declined"
	_, err = handler.messenger.UpdateRequest(&req)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, ErrorResponse{err.Error()})
	}
	return c.JSON(http.StatusOK, "Request has been updated")
}

// GetMessagesHandler godoc
// @Summary Get dialog
// @ID getMessages
// @Tags Messenger
// @Consume json
// @Produce json
// @Param Body body MessagesFilter true "filter to display messages (id_users have to contain to ID, if created_after not defined - considered as current moment)"
// @Success 200 {object} []Message
// @Failure 400,401 {object} ErrorResponse
// @Router /messenger/messages [get]
func (handler *handler) GetMessagesHandler(c echo.Context) error {
	idUser, err := handler.getIdFromContext(c)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, ErrorResponse{err.Error()})
	}
	var filter MessagesFilter
	err = c.Bind(&filter)
	if err != nil {
		log.Errorf("failed to unmarshal message filter: %s", err.Error())
		return c.JSON(http.StatusBadRequest, ErrorResponse{"Invalid request"})
	}
	if len(filter.IdUsers) != 2 {
		return c.JSON(http.StatusBadRequest, ErrorResponse{"You have to provide two user's IDs"})
	}
	if idUser != filter.IdUsers[0] && idUser != filter.IdUsers[1] {
		return c.JSON(http.StatusForbidden, ErrorResponse{"Access denied"})
	}
	msg, err := handler.messenger.GetMessages(filter.IdUsers, filter.CreatedAfter)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, ErrorResponse{err.Error()})
	}

	return c.JSON(http.StatusOK, msg)
}
