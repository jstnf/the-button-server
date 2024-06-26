package api

import (
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/jstnf/the-button-server/data"
	"net/http"
	"sort"
	"sync/atomic"
	"time"
)

type ButtonState struct {
	Presses   atomic.Int64
	LastPress atomic.Pointer[data.Press]
}

type Server struct {
	listenAddr     string
	expiry         int64
	millisPerPress int64
	storage        data.Storage
	users          data.UserStorage
	state          *ButtonState
}

func NewAPIServer(listenAddr string, expiry int64, millisPerPress int64, storage data.Storage, users data.UserStorage) *Server {
	return &Server{
		listenAddr:     listenAddr,
		expiry:         expiry,
		millisPerPress: millisPerPress,
		storage:        storage,
		users:          users,
		state:          nil,
	}
}

func (s *Server) Run() error {
	s.state = &ButtonState{}
	// Initialize last button state
	presses, err := s.storage.GetNumberOfPresses()
	if err != nil {
		return err
	}
	lastPress, err := s.storage.GetLastPress()
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
		v1.GET("/whowaslast", s.handleWhoWasLast)
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

	// Compute if the button has expired
	// Formula is expiry - currentTime - (numPresses * env.MILLIS_DEDUCTED_PER_PRESS)
	if s.buttonExpired() {
		c.JSON(http.StatusBadRequest, PressErrorResponse{Error: data.ErrorButtonExpired})
		return
	}

	user, err := s.users.GetUserById(body.UserId)
	if err != nil {
		c.JSON(http.StatusBadRequest, PressErrorResponse{Error: data.ErrorUserUnknown})
		return
	}

	// Get last press in general - if it's from the same user, return an error
	lastPress, err := s.storage.GetLastPress()
	if err != nil {
		c.JSON(http.StatusInternalServerError, PressErrorResponse{Error: err.Error()})
		return
	}
	if lastPress != nil && lastPress.UserId == user.UserId {
		c.JSON(http.StatusBadRequest, PressErrorResponse{Error: data.ErrorPressedTwice})
		return
	}

	// Get last press by user - if it's been less than 15s, return an error
	lastPress, err = s.storage.GetLastPressByUser(user.UserId)
	if err != nil {
		c.JSON(http.StatusInternalServerError, PressErrorResponse{Error: err.Error()})
		return
	}
	if lastPress != nil && lastPress.Time > (time.Now().Unix()*1000-15000) {
		c.JSON(http.StatusBadRequest, PressErrorResponse{Error: data.ErrorPressedTooSoon})
		return
	}

	// Register press
	t, err := s.storage.PressButton(user.UserId)
	if err != nil {
		c.JSON(http.StatusInternalServerError, PressErrorResponse{Error: err.Error()})
		return
	}

	// Update cached state
	s.state.Presses.Add(1)
	// Record press only if time is greater than the last press (cover race conditions)
	if s.state.LastPress.Load() == nil || t > s.state.LastPress.Load().Time {
		s.state.LastPress.Store(&data.Press{UserId: user.UserId, Time: t})
	}

	c.JSON(http.StatusOK, PressSuccessResponse{Time: t})
}

type DataResponse struct {
	Presses        int64  `json:"presses"`
	WhoPressed     string `json:"whoPressed"`
	Expiry         int64  `json:"expiry"`
	MillisPerPress int64  `json:"millisPerPress"`
}

func newDataResponse(presses int64, whoPressed string, expiry int64, millisPerPress int64) *DataResponse {
	return &DataResponse{
		Presses:        presses,
		WhoPressed:     whoPressed,
		Expiry:         expiry,
		MillisPerPress: millisPerPress,
	}
}

func (s *Server) handleGetData(c *gin.Context) {
	c.Header("Access-Control-Allow-Origin", "*")
	var name = "no one"
	lastPress := s.state.LastPress.Load()
	if lastPress != nil {
		user, err := s.users.GetUserById(lastPress.UserId)
		if err == nil {
			name = user.Name
		}
	}
	c.JSON(http.StatusOK, newDataResponse(s.state.Presses.Load(), name, s.expiry, s.millisPerPress))
}

func (s *Server) buttonExpired() bool {
	return s.expiry-(time.Now().Unix()*1000)-(s.state.Presses.Load()*s.millisPerPress) < 0
}

type WhoWasLastUserEntry struct {
	Name string `json:"name"`
	Time int64  `json:"time"`
}

type WhoWasLastResponse struct {
	Users []WhoWasLastUserEntry `json:"users"`
}

func (s *Server) handleWhoWasLast(c *gin.Context) {
	c.Header("Access-Control-Allow-Origin", "*")
	var lastPressData []WhoWasLastUserEntry
	allUsers, err := s.users.GetAllUsers()
	if err != nil {
		c.JSON(http.StatusInternalServerError, PressErrorResponse{Error: err.Error()})
		return
	}
	for _, user := range allUsers {
		lastPress, err := s.storage.GetLastPressByUser(user.UserId)
		if err != nil {
			c.JSON(http.StatusInternalServerError, PressErrorResponse{Error: err.Error()})
			return
		}
		if lastPress != nil {
			i := sort.Search(len(lastPressData), func(i int) bool { return lastPressData[i].Time > lastPress.Time })
			lastPressData = append(lastPressData, WhoWasLastUserEntry{})
			copy(lastPressData[i+1:], lastPressData[i:])
			lastPressData[i] = WhoWasLastUserEntry{Name: user.Name, Time: lastPress.Time}
		}
	}
	c.JSON(http.StatusOK, WhoWasLastResponse{Users: lastPressData})
}
