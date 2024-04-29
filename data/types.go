package data

type Press struct {
	UserId string `json:"userId"`
	Time   int64  `json:"time"`
}

type User struct {
	UserId string `json:"userId"`
	Name   string `json:"name"`
}
