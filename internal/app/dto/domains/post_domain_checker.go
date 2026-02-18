package domains

type DomainCheckerRequest struct {
	Name    string   `json:"name" binding:"required"`
	Domains []string `json:"domains" binding:"required"`
}

type DomainCheckerResponse struct {
	Message string `json:"message"`
	Success bool   `json:"success"`
}
