package credentials

type ResultUserCredentialRow interface {
	GetUser() *WebauthnUser
	GetCredential() *WebauthnCredential
}

func (r GetUserWithCredentialsByIDRow) GetUser() *WebauthnUser {
	return &r.WebauthnUser
}

func (r GetUserWithCredentialsByIDRow) GetCredential() *WebauthnCredential {
	return &r.WebauthnCredential
}

func (r GetUserWithCredentialsByRefRow) GetUser() *WebauthnUser {
	return &r.WebauthnUser
}

func (r GetUserWithCredentialsByRefRow) GetCredential() *WebauthnCredential {
	return &r.WebauthnCredential
}
