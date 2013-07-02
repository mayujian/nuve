package nuve

import (
	"fmt"
)

type User struct {
	Name string `json:"name"`
	Role string `json:"role"`
}

type RoomOption struct {
}

type Room struct {
	Name string `json:"name"`
	Id   string `json:"_id"`
	RoomOption
}

type Service struct {
	Id    string `json:"_id"`
	Name  string `json:"name"`
	Key   string `json:"key"`
	Rooms []Room `json:"rooms"`
}

func (s Service) String() string {
	return fmt.Sprintf(`
{
    Id    %s
    Key   %s
    Name  %s
    Rooms %v
}`, s.Id, s.Key, s.Name, s.Rooms)
}

func (r Room) String() string {
	return fmt.Sprintf(`{Id: %s Name: %s}`, r.Id, r.Name)
}
