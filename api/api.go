package api

import (
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/jstnf/the-button-server/data"
	"net/http"
	"sync/atomic"
	"time"
)

type ButtonState struct {
	Presses   atomic.Int64
	LastPress atomic.Pointer[data.Press]
}

type Server struct {
	listenAddr string
	Store      data.Storage
	state      *ButtonState
}

func NewAPIServer(listenAddr string, storage data.Storage) *Server {
	return &Server{
		listenAddr: listenAddr,
		Store:      storage,
		state:      nil,
	}
}

func (s *Server) Run() error {
	s.state = &ButtonState{}
	// Initialize last button state
	presses, err := s.Store.GetNumberOfPresses()
	if err != nil {
		return err
	}
	lastPress, err := s.Store.GetLastPress()
	if err != nil {
		return err
	}
	if lastPress != nil {
		s.state.LastPress.Store(lastPress)
	}
	s.state.Presses.Store(presses)

	router := gin.Default()

	v1 := router.Group("/api/v1")
	{
		v1.POST("/press", s.handlePostPress)
		v1.GET("/data", s.handleGetData)
	}

	router.Use(cors.Default())
	return router.Run(s.listenAddr)
}

type PressRequestBody struct {
	UserId string `json:"userId"`
}

type PressSuccessResponse struct {
	Time int64 `json:"time"`
}

type PressErrorResponse struct {
	Error string `json:"error"`
}

func (s *Server) handlePostPress(c *gin.Context) {
	c.Header("Access-Control-Allow-Origin", "*")
	var body PressRequestBody
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, PressErrorResponse{Error: data.ErrorUserUnknown})
		return
	}

	// TODO authentication

	// Get last press in general - if it's from the same user, return an error
	lastPress, err := s.Store.GetLastPress()
	if err != nil {
		c.JSON(http.StatusInternalServerError, PressErrorResponse{Error: err.Error()})
		return
	}
	if lastPress != nil && lastPress.UserId == body.UserId {
		c.JSON(http.StatusBadRequest, PressErrorResponse{Error: data.ErrorPressedTwice})
		return
	}

	// Get last press by user - if it's been less than 15s, return an error
	lastPress, err = s.Store.GetLastPressByUser(body.UserId)
	if err != nil {
		c.JSON(http.StatusInternalServerError, PressErrorResponse{Error: err.Error()})
		return
	}
	if lastPress != nil && lastPress.Time > (time.Now().Unix()*1000-15000) {
		c.JSON(http.StatusBadRequest, PressErrorResponse{Error: data.ErrorPressedTooSoon})
		return
	}

	// Register press
	t, err := s.Store.PressButton(body.UserId)
	if err != nil {
		c.JSON(http.StatusInternalServerError, PressErrorResponse{Error: err.Error()})
		return
	}

	// Update cached state
	s.state.Presses.Add(1)
	// Record press only if time is greater than the last press (cover race conditions)
	if s.state.LastPress.Load() == nil || t > s.state.LastPress.Load().Time {
		s.state.LastPress.Store(&data.Press{UserId: body.UserId, Time: t})
	}

	c.JSON(http.StatusOK, PressSuccessResponse{Time: t})
}

type DataResponse struct {
	Presses    int64  `json:"presses"`
	WhoPressed string `json:"whoPressed"`
	Expiry     int64  `json:"expiry"`
}

func newDataResponse(presses int64, whoPressed string, expiry int64) *DataResponse {
	return &DataResponse{
		Presses:    presses,
		WhoPressed: whoPressed,
		Expiry:     expiry,
	}
}

func (s *Server) handleGetData(c *gin.Context) {
	c.Header("Access-Control-Allow-Origin", "*")
	var name = "no one"
	lastPress := s.state.LastPress.Load()
	if lastPress != nil {
		// TODO resolve userId to name
		name = lastPress.UserId
	}
	c.JSON(http.StatusOK, newDataResponse(s.state.Presses.Load(), name, 1716015600000))
}
