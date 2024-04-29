package data

type Storage interface {
	PressButton(userId string) error
	GetLastPress() (Press, error)
	GetLastPressByUser(userId string) (Press, error)
}
