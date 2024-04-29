package api

import (
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/jstnf/the-button-server/data"
	"net/http"
	"time"
)

type Server struct {
	listenAddr string
	Store      data.Storage
}

func NewAPIServer(listenAddr string, storage data.Storage) *Server {
	return &Server{
		listenAddr: listenAddr,
		Store:      storage,
	}
}

func (s *Server) Run() error {
	router := gin.Default()

	v1 := router.Group("/api/v1")
	{
		v1.POST("/press", s.handlePostPress)
		v1.GET("/data", s.handleGetData)
	}

	router.Use(cors.Default())
	return router.Run(s.listenAddr)
}

var localPresses int64 = 0
var whoPressed string = "no one"

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

	c.JSON(http.StatusOK, PressSuccessResponse{Time: t})
}

type DataResponse struct {
	Presses    int64  `json:"presses"`
	WhoPressed string `json:"whoPressed"`
	Expiry     int64  `json:"expiry"`
}

func NewDataResponse(presses int64, whoPressed string, expiry int64) *DataResponse {
	return &DataResponse{
		Presses:    presses,
		WhoPressed: whoPressed,
		Expiry:     expiry,
	}
}

func (s *Server) handleGetData(c *gin.Context) {
	c.Header("Access-Control-Allow-Origin", "*")
	c.JSON(http.StatusOK, NewDataResponse(localPresses, whoPressed, 1716015600000))
}
