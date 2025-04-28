package domain

type User struct {
	Login  string `json:"login"`
	PubKey string `json:"pubkey"`
	ID     int    `json:"id"`
}

type UserList struct {
	Users []*User `json:"users"`
}

func NewUserList() *UserList {
	return &UserList{
		Users: make([]*User, 0),
	}
}

func NewUser(login string, id int) *User {
	return &User{
		Login: login,
		ID:    id,
	}
}

func (u *User) SetPubKey(pubKey string) {
	u.PubKey = pubKey
}
