package service

import "sync"

type UserStore interface {
	Save(user *User) error
	Find(username string) (*User, error)
}

type InMemoryUserStore struct {
	mutex sync.RWMutex
	users map[string]*User
}

func NewInMemoryUserStore() *InMemoryUserStore {
	return &InMemoryUserStore{
		users: make(map[string]*User),
	}
}

func (u *InMemoryUserStore) Save(user *User) error {
	u.mutex.Lock()
	defer u.mutex.Unlock()

	if u.users[user.Username] != nil {
		return ErrAlreadyExists
	}

	u.users[user.Username] = user.Clone()
	return nil
}

func (u *InMemoryUserStore) Find(username string) (*User, error) {
	u.mutex.RLock()
	defer u.mutex.RUnlock()

	user := u.users[username]
	if user == nil {
		return nil, nil
	}
	return user.Clone(), nil
}
