package api

import (
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/jstnf/the-button-server/data"
	"net/http"
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

type PressErrorResponse struct {
	Error string `json:"error"`
}

func (s *Server) handlePostPress(c *gin.Context) {
	c.Header("Access-Control-Allow-Origin", "*")
	var body PressRequestBody
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, PressErrorResponse{Error: "Invalid request body"})
		return
	}

	localPresses++
	whoPressed = body.UserId
	c.JSON(http.StatusOK, gin.H{"message": "pressed"})
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
