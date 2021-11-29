package handlers

type (
	ErrorResponse struct {
		// Message about the error
		Message string `json:"message" example:"Operation failed"`
	} // @name ErrorResponse

	UserLoginParams struct {
		Username string `json:"username" example:"andrey@gmail.com"`
		Password string `json:"password" example:"Password123"`
	} // @name UserLoginParams

	LoginResponse struct {
		// Token to access protected pages
		IdUser      int64  `json:"id_user" example:"9338554"`
		Username    string `json:"username" example:"andrey"`
		AccessToken string `json:"access_token" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE2MzYyMjUxNDcsImp0aSI6IjkxMzQ1NzQ5NzUifQ.hiQUF6DNwoOcYsBvo1-aRVEQShzRMvGYReHWKg6QY4I"`
	} // @name LoginResponse

)

const refreshToken = "refresh_token"
const xAuthToken = "X-Auth-Token"
const admin = "admin"
