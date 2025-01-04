package main

// viewData represents the root model used to dynamically update the app
type viewData struct {
	ViewType  string `json:"viewType" redis:"viewType"`
	PageTitle string
	AppName   string
	Stream    []*post
	Nonce     string
	Order     string `json:"order" redis:"order"`
}
