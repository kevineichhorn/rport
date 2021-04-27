package users

const (
	Administrators = "Administrators"
)

// User represents API user.
type User struct {
	Username string   `json:"username"`
	Password string   `json:"-"`
	Groups   []string `json:"groups"`
}
