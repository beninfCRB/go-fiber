package handler

// Handlers is an aggregate of all handler structs.
type Handlers struct {
	Auth *AuthHandler
	User *UserHandler
	Menu *MenuHandler
}

func NewHandlers(auth *AuthHandler, user *UserHandler, menu *MenuHandler) *Handlers {
	return &Handlers{Auth: auth, User: user, Menu: menu}
}
