package handlers

import (
	"SB/catb406.github.io/service/config"
	"SB/catb406.github.io/service/repository/db"
	"SB/catb406.github.io/service/repository/messenger"
	"SB/catb406.github.io/service/repository/token"
	"SB/catb406.github.io/service/repository/training"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/dgrijalva/jwt-go"
	"github.com/labstack/echo/v4"
	"github.com/labstack/gommon/log"
	"net/http"
	"strconv"
	"time"
)

type (
	handler struct {
		userManager db.UserManager
		token       token.TokenManager
		trainingMgr training.TrainingManager
		messenger   messenger.Messenger
	}

	Handler interface {
		LoginHandler(c echo.Context) error
		LogoutHandler(c echo.Context) error
		SignupHandler(c echo.Context) error
		DeleteUserHandler(c echo.Context) error
		AccessMiddleware(next echo.HandlerFunc) echo.HandlerFunc
		GetUserProfileHandler(c echo.Context) error
		GetProfilesHandler(c echo.Context) error
		UpdateUserProfileHandler(c echo.Context) error
		RefreshToken(c echo.Context) error
		GetGroupTrainingsHandler(c echo.Context) error
		GetTrainingHandler(c echo.Context) error
		AddGroupTrainingHandler(c echo.Context) error
		UpdateGroupTrainingHandler(c echo.Context) error
		DeleteGroupTrainingHandler(c echo.Context) error
		GetUserTrainingsHandler(c echo.Context) error
		MessengerHandler(c echo.Context) error
		GetDialogsHandler(c echo.Context) error
		SendRequestHandler(c echo.Context) error
		ReplyToRequestHandler(c echo.Context) error
		DeclinedRequestSeenHandler(c echo.Context) error
		GetMessagesHandler(c echo.Context) error
	}
)

func NewHandler(usrMgr db.UserManager, tknMgr token.TokenManager, trainingMgr training.TrainingManager, messenger messenger.Messenger) Handler {
	return &handler{
		userManager: usrMgr,
		token:       tknMgr,
		messenger:   messenger,
		trainingMgr: trainingMgr,
	}
}

//TODO: add api for admin to administrate data

// LoginHandler godoc
// @Summary Login a user
// @ID authLogin
// @Tags Auth
// @Accept  json
// @Produce  json
// @Param Body body UserLoginParams true "The body to login a user"
// @Success 200 {object} LoginResponse
// @Failure 400,401,500 {object} ErrorResponse
// @Router /auth/login [post]
func (handler *handler) LoginHandler(c echo.Context) error {
	loginParams := new(UserLoginParams)
	err := c.Bind(&loginParams)
	if err != nil {
		log.Error(err)
		return c.JSON(http.StatusBadRequest, ErrorResponse{"Invalid login parameters"})
	}
	log.Info("login ", loginParams.Username)
	user, err := handler.userManager.Authenticate(loginParams.Username, loginParams.Password)
	if err != nil {
		log.Error("login failed: ", err)
		return c.JSON(http.StatusUnauthorized, ErrorResponse{"Failed to login"})
	}
	accessTkn, refreshTkn, err := handler.token.GenerateNewToken(user.GetId(), user.GetRole(), user.GetUsername())
	if err != nil {
		log.Error("failed to generate token", err)
		return c.JSON(http.StatusInternalServerError, ErrorResponse{"Failed to generate token"})
	}

	resp := LoginResponse{
		IdUser:      user.GetId(),
		Username:    user.GetUsername(),
		AccessToken: accessTkn.GetId(),
	}

	cookie := &http.Cookie{
		Name:     refreshToken,
		Value:    refreshTkn.GetId(),
		Expires:  refreshTkn.GetExpirationTime(),
		Path:     "/auth",
		HttpOnly: true,
	}

	c.SetCookie(cookie)
	return c.JSON(http.StatusOK, resp)
}

// LogoutHandler  godoc
// @Summary Logout a user
// @ID authLogout
// @Tags Auth
// @Success 200 {string} string "Logout successfully"
// @Failure 401,404,500 {object} ErrorResponse
// @Router /auth/logout [post]
func (handler *handler) LogoutHandler(c echo.Context) error {
	cookie, err := c.Cookie(refreshToken)
	if err != nil {
		log.Error("failed to get refresh token from cookie: ", err)
		return c.JSON(http.StatusUnauthorized, ErrorResponse{"Failed to logout"})
	}
	log.Info("logout ", cookie.Value)

	err = handler.token.Remove(cookie.Value)
	if err != nil {
		log.Error("failed to logout", err)
		return c.JSON(http.StatusNotFound, ErrorResponse{"No such token"})
	}
	cookie.MaxAge = -1
	c.SetCookie(cookie)
	return c.JSON(http.StatusOK, "Logout successfully")
}

// SignupHandler godoc
// @Summary Sign up a user
// @ID authSignup
// @Tags Auth
// @Accept  json
// @Produce  json
// @Param Body body UserLoginParams true "The body to sign up a user"
// @Success 200 {object} LoginResponse
// @Failure 400,401,500 {object} ErrorResponse
// @Router /auth/signup [post]
func (handler *handler) SignupHandler(c echo.Context) error {
	var signUpParams UserLoginParams
	err := c.Bind(&signUpParams)
	if err != nil {
		log.Error("failed to sign up: ", err)
		return c.JSON(http.StatusBadRequest, ErrorResponse{"Invalid sign up parameters"})
	}
	log.Info("sign up ", signUpParams.Username)
	user, err := handler.userManager.AddUser(signUpParams.Username, signUpParams.Password)
	if err != nil {
		log.Error("failed to sign up: ", err)
		return c.JSON(http.StatusBadRequest, ErrorResponse{"Failed to sign up: invalid user parameters or user already exists"})
	}

	accessTkn, refreshTkn, err := handler.token.GenerateNewToken(user.GetId(), user.GetRole(), user.GetUsername())
	if err != nil {
		log.Error("failed: to generate token", err)
		return c.JSON(http.StatusInternalServerError, ErrorResponse{"Failed to generate token"})
	}

	resp := LoginResponse{
		IdUser:      user.GetId(),
		Username:    user.GetUsername(),
		AccessToken: accessTkn.GetId(),
	}

	cookie := &http.Cookie{
		Name:     refreshToken,
		Value:    refreshTkn.GetId(),
		Expires:  refreshTkn.GetExpirationTime(),
		Path:     "/auth",
		HttpOnly: true,
	}

	c.SetCookie(cookie)
	return c.JSON(http.StatusOK, resp)
}

func (handler *handler) AccessMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		log.Info("check token")
		jwtFromHeader := c.Request().Header.Get(xAuthToken)
		tkn, err := jwt.ParseWithClaims(
			jwtFromHeader,
			&jwt.StandardClaims{},
			func(token *jwt.Token) (interface{}, error) {
				return []byte(config.Secret), nil
			},
		)
		if err != nil {
			log.Error(err)
			return c.JSON(http.StatusUnauthorized, ErrorResponse{"Failed to authenticate user"})
		}
		//claims, ok := tkn.Claims.(*token.CustomizedClaims)
		claims, ok := tkn.Claims.(*jwt.StandardClaims)

		if !ok {
			log.Error(errors.New("couldn't parse claims"))
			return c.JSON(http.StatusForbidden, ErrorResponse{Message: "Failed to parse token claims"})
		}
		if claims.ExpiresAt < time.Now().UTC().Unix() {
			log.Error(errors.New("jwt is expired"))
			return c.JSON(http.StatusForbidden, ErrorResponse{Message: "JWT is expired"})
		}
		c.Set("id_user", claims.Id)
		//c.Set("role", claims.Role)
		return next(c)
	}
}

// RefreshToken godoc
// @Summary Refresh access token
// @ID authRefreshToken
// @Tags Auth
// @Produce  json
// @Success 200 {object} LoginResponse
// @Failure 400,401 {object} ErrorResponse
// @Router /auth/refresh [post]
func (handler *handler) RefreshToken(c echo.Context) error {
	cookie, err := c.Cookie(refreshToken)
	if err != nil {
		log.Error("failed to auth: ", err)
		return c.JSON(http.StatusBadRequest, ErrorResponse{"Failed to refresh token"})
	}
	accessTkn, refreshTkn, err := handler.token.Refresh(cookie.Value)
	if err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{"Failed to refresh token"})
	}

	resp := LoginResponse{
		Username:    accessTkn.GetUsername(),
		IdUser:      accessTkn.GetUserId(),
		AccessToken: accessTkn.GetId(),
	}

	cookie = &http.Cookie{
		Name:     refreshToken,
		Value:    refreshTkn.GetId(),
		Expires:  refreshTkn.GetExpirationTime(),
		Path:     "/auth",
		HttpOnly: true,
	}

	c.SetCookie(cookie)

	return c.JSON(http.StatusOK, resp)
}

// GetUserProfileHandler godoc
// @Summary Get user profile
// @ID userGetProfile
// @Tags Profile
// @Produce  json
// @Success 200 {object} db.UserProfile
// @Failure 400,401 {object} ErrorResponse
// @Router /user/profile/{id} [get]
func (handler *handler) GetUserProfileHandler(c echo.Context) error {
	paramId := c.Param("id")
	id, err := strconv.ParseInt(paramId, 10, 64)
	if err != nil {
		log.Error(err)
		return c.JSON(http.StatusBadRequest,
			ErrorResponse{"Invalid user id"})
	}
	profile, err := handler.userManager.GetUserProfile(id)
	if err != nil {
		log.Error(err)
		return c.JSON(http.StatusBadRequest, ErrorResponse{Message: fmt.Sprintf("Failed to get profile: user with id %d does no exist", id)})
	}
	return c.JSON(http.StatusOK, profile)
}

// UpdateUserProfileHandler godoc
// @Summary Update user profile
// @ID userUpdateProfile
// @Tags Profile
// @Accept json
// @Produce  json
// @Param Body body db.UserProfile true "The body to update user profile"
// @Success 200 {string} string "user profile successfully updated"
// @Failure 400,401 {object} ErrorResponse
// @Router /user/profile [put]
// TODO: get user id from token
func (handler *handler) UpdateUserProfileHandler(c echo.Context) error {
	id, err := handler.getIdFromContext(c)
	if err != nil {
		return err
	}
	var userProfile db.UserProfile
	err = c.Bind(&userProfile)
	if err != nil {
		log.Error(err)
		return c.JSON(http.StatusBadRequest, ErrorResponse{"Invalid profile parameters"})
	}
	if id != userProfile.IdUser {
		if !handler.checkForAdminPrivileges(id) {
			return c.JSON(http.StatusForbidden, ErrorResponse{"No rights to configure users"})
		}
	}
	err = handler.userManager.UpdateUserProfile(&userProfile, userProfile.Sport...)
	if err != nil {
		log.Error(err)
		return c.JSON(http.StatusBadRequest, ErrorResponse{Message: "Failed to update user profile"})
	}
	return c.JSON(http.StatusOK, ErrorResponse{"User profile successfully updated"})
}

// DeleteUserHandler godoc
// @Summary Delete user
// @ID userDeleteProfile
// @Tags Profile
// @Success 200 {string} string "user successfully deleted"
// @Failure 400,401 {object} ErrorResponse
// @Router /user/{id} [delete]
func (handler *handler) DeleteUserHandler(c echo.Context) error {
	id, err := handler.getIdFromContext(c)
	if err != nil {
		return err
	}
	paramId := c.Param("id")
	idFromPath, err := strconv.ParseInt(paramId, 10, 64)
	if err != nil {
		log.Error(err)
		return c.JSON(http.StatusBadRequest,
			ErrorResponse{"Invalid user id"})
	}
	if id != idFromPath {
		if !handler.checkForAdminPrivileges(id) {
			return c.JSON(http.StatusForbidden, ErrorResponse{"No rights to configure users"})
		}
	}
	err = handler.userManager.DeleteUser(id)
	if err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{fmt.Sprintf("Failed to delete user: %s", err)})
	}
	return c.JSON(http.StatusOK, "User successfully deleted")
}

// GetProfilesHandler godoc
// @Summary Get users profiles filtered by provided parameters
// @ID userGetFilteredProfiles
// @Tags Training
// @Consumes json
// @Produce  json
// @Param Body body db.UserProfileFilterParams true "The body to filter users profiles"
// @Success 200 {object} []persistence.FilteredUserProfileImpl
// @Failure 400,401 {object} ErrorResponse
// @Router /training/profiles [get]
func (handler *handler) GetProfilesHandler(c echo.Context) error {
	var filterParams db.UserProfileFilterParams
	err := c.Bind(&filterParams)
	if err != nil {
		log.Error(err)
		return c.JSON(http.StatusBadRequest, ErrorResponse{"Invalid filter parameters"})
	}
	profiles := handler.userManager.GetProfiles(filterParams)
	return c.JSON(http.StatusOK, profiles)
}

// GetGroupTrainingsHandler godoc
// @Summary Get group trainings filtered by provided parameters
// @ID getFilteredTrainings
// @Tags Training
// @Consumes json
// @Produce  json
// @Param Body body training.GroupTrainingFilter false "The body to filter group trainings"
// @Success 200 {object} []training.GroupTraining
// @Failure 400,401 {object} ErrorResponse
// @Router /training [get]
func (handler *handler) GetGroupTrainingsHandler(c echo.Context) error {
	var filter training.GroupTrainingFilter
	if err := c.Bind(&filter); err != nil {
		return c.JSON(http.StatusBadRequest,
			ErrorResponse{"Invalid filter parameters"})
	}
	trainings := handler.trainingMgr.GetTrainings(filter)
	return c.JSON(http.StatusOK, trainings)
}

// GetTrainingHandler godoc
// @Summary Get training by id
// @ID getTrainingByID
// @Tags Training
// @Consumes json
// @Produce  json
// @Success 200 {object} training.GroupTraining
// @Failure 400,401 {object} ErrorResponse
// @Router /training/{id} [get]
func (handler *handler) GetTrainingHandler(c echo.Context) error {
	paramId := c.Param("id")
	idFromPath, err := strconv.ParseInt(paramId, 10, 64)
	if err != nil {
		log.Error(err)
		return c.JSON(http.StatusBadRequest,
			ErrorResponse{"Invalid training id"})
	}
	t := handler.trainingMgr.GetTraining(idFromPath)
	return c.JSON(http.StatusOK, t)
}

// AddGroupTrainingHandler godoc
// @Summary Add group training
// @ID addGroupTraining
// @Tags Training
// @Consumes json
// @Produce  json
// @Param Body body training.GroupTraining true "The body to add a group training"
// @Success 200 {object} object training.GroupTraining
// @Failure 400,401 {object} ErrorResponse
// @Router /training [post]
func (handler *handler) AddGroupTrainingHandler(c echo.Context) error {
	var gt training.GroupTraining
	if err := c.Bind(&gt); err != nil {
		log.Error(err)
		return c.JSON(http.StatusBadRequest,
			ErrorResponse{"Invalid training parameters"})
	}
	t, err := handler.trainingMgr.AddTraining(gt)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, ErrorResponse{fmt.Sprintf("Failed to add training %s", err)})
	}
	return c.JSON(http.StatusOK, t)
}

// UpdateGroupTrainingHandler godoc
// @Summary Update group training
// @ID updateGroupTraining
// @Tags Training
// @Consumes json
// @Produce  json
// @Param Body body training.GroupTraining true "The body to update a group training"
// @Success 200 {object} object training.GroupTraining
// @Failure 400,401 {object} ErrorResponse
// @Router /training/{id} [put]
func (handler *handler) UpdateGroupTrainingHandler(c echo.Context) error {
	code, resp := handler.checkForTrainingOwnership(c)
	if code != http.StatusOK {
		return c.JSON(code, resp)
	}
	trainingStruct := training.GroupTraining{}
	err := c.Bind(&trainingStruct)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, ErrorResponse{"Invalid training parameters"})
	}
	t, err := handler.trainingMgr.UpdateTraining(trainingStruct)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, ErrorResponse{fmt.Sprintf("Failed to update training: %s", err)})
	}
	return c.JSON(http.StatusOK, t)
}

// DeleteGroupTrainingHandler godoc
// @Summary Delete group training
// @ID deleteGroupTraining
// @Tags Training
// @Success 200 {string} string
// @Failure 400,401 {object} ErrorResponse
// @Router /training/{id} [delete]
func (handler *handler) DeleteGroupTrainingHandler(c echo.Context) error {
	code, resp := handler.checkForTrainingOwnership(c)
	if code != http.StatusOK {
		return c.JSON(code, resp)
	}
	id := c.Get("training_id").(int64)
	err := handler.trainingMgr.DeleteTraining(id)
	if err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{fmt.Sprintf("Failed to delete training: %s", err)})
	}
	return c.JSON(http.StatusOK, "Training successfully deleted")
}

// GetUserTrainingsHandler godoc
// @Summary Get user's trainings
// @ID getUserTrainings
// @Tags Calendar
// @Produce  json
// @Success 200 {object} training.GroupTraining
// @Failure 400,401 {object} ErrorResponse
// @Router /user/{id}/trainings [get]
func (handler *handler) GetUserTrainingsHandler(c echo.Context) error {
	idUser, err := handler.getIdFromContext(c)
	if err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{err.Error()})
	}
	idUserFromPath := c.Param("id")
	numId, err := strconv.ParseInt(idUserFromPath, 10, 64)
	if err != nil {
		log.Error(err)
		return c.JSON(http.StatusInternalServerError, ErrorResponse{"Failed to parse id from path"})
	}
	if idUser != numId {
		if !handler.checkForAdminPrivileges(idUser) {
			return c.JSON(http.StatusForbidden, ErrorResponse{"Access denied"})
		}
	}
	trainings, err := handler.userManager.GetUserTrainings(numId)
	if err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{err.Error()})
	}
	return c.JSON(http.StatusOK, trainings)
}

// RemoveFromContactsHandler godoc
// @Summary Remove user from contacts
// @ID removeFromContacts
// @Tags Messenger
// @Produce json
// @Success 200 {string} string
// @Failure 400,401 {object} ErrorResponse
// @Router /user/contacts/{id} [delete]
// TODO: get user id from token
func (handler *handler) RemoveFromContactsHandler(c echo.Context) error {
	return c.JSON(http.StatusNotImplemented, "Method is not implemented")
}

func (handler *handler) checkForAdminPrivileges(id int64) bool {
	role := handler.userManager.GetRole(id)
	if role == admin {
		return true
	}
	return false
}

func (handler *handler) getIdFromContext(c echo.Context) (int64, error) {
	id, ok := c.Get("id_user").(string)
	if !ok {
		log.Error("incorrect id ", c.Get("id_user"))
		return 0, c.JSON(http.StatusInternalServerError, "Failed to parse id from token")
	}
	numId, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		log.Error(err)
		return 0, c.JSON(http.StatusInternalServerError, "Failed to parse id from token")
	}
	return numId, nil
}

func (handler *handler) checkForTrainingOwnership(c echo.Context) (int, ErrorResponse) {
	paramId := c.Param("id")
	idFromPath, err := strconv.ParseInt(paramId, 10, 64)
	if err != nil {
		log.Error(err)
		return http.StatusBadRequest,
			ErrorResponse{"Invalid training id"}
	}
	t := handler.trainingMgr.GetTraining(idFromPath)
	userId, err := handler.getIdFromContext(c)
	trainingStruct := training.GroupTraining{}
	err = json.Unmarshal(t.Serialize(), &trainingStruct)
	if err != nil {
		return http.StatusInternalServerError, ErrorResponse{"Failed to get training owner"}
	}
	log.Info(trainingStruct)
	if userId != trainingStruct.Owner {
		if !handler.checkForAdminPrivileges(userId) {
			return http.StatusForbidden, ErrorResponse{"No rights to configure trainings"}
		}
	}
	c.Set("training_id", idFromPath)
	return http.StatusOK, ErrorResponse{}
}
