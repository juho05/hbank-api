package handlers

import (
	"bytes"
	"crypto/subtle"
	"fmt"
	"image/png"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/Bananenpro/hbank-api/bindings"
	"github.com/Bananenpro/hbank-api/config"
	"github.com/Bananenpro/hbank-api/models"
	"github.com/Bananenpro/hbank-api/responses"
	"github.com/Bananenpro/hbank-api/services"
	"github.com/google/uuid"
	"github.com/h2non/bimg"
	"github.com/labstack/echo/v4"
	"golang.org/x/crypto/bcrypt"
)

// /v1/user?includeSelf=bool (GET)
func (h *Handler) GetUsers(c echo.Context) error {
	lang := c.Get("lang").(string)
	authUserId := c.Get("userId").(uuid.UUID)
	authUser, err := h.userStore.GetById(authUserId)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, responses.NewUnexpectedError(err, lang))
	}
	if authUser == nil {
		return c.JSON(http.StatusUnauthorized, responses.NewUserNoLongerExists(lang))
	}

	var users []models.User
	if services.StrToBool(c.QueryParams().Get("includeSelf")) {
		users, err = h.userStore.GetAll(nil)
	} else {
		users, err = h.userStore.GetAll(authUser)
	}
	if err != nil {
		return c.JSON(http.StatusInternalServerError, responses.NewUnexpectedError(err, lang))
	}

	return c.JSON(http.StatusOK, responses.NewUsers(users))
}

// /v1/user/:id (GET)
func (h *Handler) GetUser(c echo.Context) error {
	lang := c.Get("lang").(string)
	authUserId := c.Get("userId").(uuid.UUID)
	authUser, err := h.userStore.GetById(authUserId)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, responses.NewUnexpectedError(err, lang))
	}
	if authUser == nil {
		return c.JSON(http.StatusUnauthorized, responses.NewUserNoLongerExists(lang))
	}

	userId, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, responses.New(false, "Invalid or missing id parameter", lang))
	}

	user, err := h.userStore.GetById(userId)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, responses.NewUnexpectedError(err, lang))
	}
	if user == nil {
		return c.JSON(http.StatusNotFound, responses.NewNotFound(lang))
	}

	if bytes.Equal(authUserId[:], userId[:]) {
		return c.JSON(http.StatusOK, responses.NewAuthUser(authUser))
	}
	return c.JSON(http.StatusOK, responses.NewUser(user))
}

// /v1/user (DELETE)
func (h *Handler) DeleteUser(c echo.Context) error {
	lang := c.Get("lang").(string)
	userId := c.Get("userId").(uuid.UUID)
	user, err := h.userStore.GetById(userId)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, responses.NewUnexpectedError(err, lang))
	}
	if user == nil {
		return c.JSON(http.StatusUnauthorized, responses.NewUserNoLongerExists(lang))
	}

	var body bindings.DeleteUser
	err = c.Bind(&body)
	if err != nil {
		return c.JSON(http.StatusBadRequest, responses.NewInvalidRequestBody(lang))
	}

	if bcrypt.CompareHashAndPassword(user.PasswordHash, []byte(body.Password)) != nil {
		return c.JSON(http.StatusForbidden, responses.NewInvalidCredentials(lang))
	}

	twoFAToken, err := h.userStore.GetTwoFATokenByCode(user, body.TwoFAToken)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, responses.NewUnexpectedError(err, lang))
	}
	if twoFAToken == nil {
		return c.JSON(http.StatusForbidden, responses.NewInvalidCredentials(lang))
	}
	h.userStore.DeleteTwoFAToken(twoFAToken)
	if twoFAToken.ExpirationTime < time.Now().Unix() {
		return c.JSON(http.StatusForbidden, responses.NewInvalidCredentials(lang))
	}

	h.userStore.Delete(user)
	return c.JSON(http.StatusOK, responses.New(true, "Successfully deleted account", lang))
}

// /v1/user/:id (DELETE)
func (h *Handler) DeleteUserByConfirmEmailCode(c echo.Context) error {
	lang := c.Get("lang").(string)
	var body bindings.DeleteUserByConfirmEmailCode
	err := c.Bind(&body)
	if err != nil {
		return c.JSON(http.StatusBadRequest, responses.NewInvalidRequestBody(lang))
	}

	userId, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, responses.New(false, "Invalid or missing id parameter", lang))
	}

	user, err := h.userStore.GetById(userId)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, responses.NewUnexpectedError(err, lang))
	}
	if user == nil {
		return c.JSON(http.StatusUnauthorized, responses.NewInvalidCredentials(lang))
	}

	code, err := h.userStore.GetConfirmEmailCode(user)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, responses.NewUnexpectedError(err, lang))
	}

	if subtle.ConstantTimeCompare(code.CodeHash, services.HashToken(body.ConfirmEmailCode)) == 1 {
		if !user.EmailConfirmed {
			h.userStore.Delete(user)
			return c.JSON(http.StatusOK, responses.New(true, "Successfully deleted account", lang))
		}
		h.userStore.DeleteConfirmEmailCode(code)
	}
	return c.JSON(http.StatusUnauthorized, responses.NewInvalidCredentials(lang))
}

// /v1/user/profilePicture (POST)
func (h *Handler) SetProfilePicture(c echo.Context) error {
	lang := c.Get("lang").(string)

	userId := c.Get("userId").(uuid.UUID)
	user, err := h.userStore.GetById(userId)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, responses.NewUnexpectedError(err, lang))
	}
	if user == nil {
		return c.JSON(http.StatusUnauthorized, responses.NewUserNoLongerExists(lang))
	}

	file, err := c.FormFile("profile_picture")
	if err != nil {
		return c.JSON(http.StatusBadRequest, responses.New(false, "Invalid or missing profile picture file", lang))
	}

	if file.Size > config.Data.UserMaxProfilePictureSize {
		return c.JSON(http.StatusBadRequest, responses.New(false, fmt.Sprintf(services.Tr("File too big (max %d bytes)", lang), config.Data.UserMaxProfilePictureSize), ""))
	}

	src, err := file.Open()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, responses.NewUnexpectedError(err, lang))
	}
	defer src.Close()

	var buffer bytes.Buffer
	_, err = io.Copy(&buffer, src)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, responses.NewUnexpectedError(err, lang))
	}
	_, err = png.Decode(src)
	data := buffer.Bytes()

	img := bimg.NewImage(data)
	if img.Type() == "unknown" {
		return c.JSON(http.StatusBadRequest, responses.New(false, "File is not an image", lang))
	}

	data, err = img.Convert(bimg.JPEG)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, responses.NewUnexpectedError(err, lang))
	}

	data, err = bimg.NewImage(data).AutoRotate()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, responses.NewUnexpectedError(err, lang))
	}

	data, err = bimg.NewImage(data).Thumbnail(config.Data.UserProfilePictureSize)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, responses.NewUnexpectedError(err, lang))
	}

	user.ProfilePicture = data
	user.ProfilePictureId = uuid.New()
	err = h.userStore.Update(user)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, responses.NewUnexpectedError(err, lang))
	}

	return c.JSON(http.StatusOK, responses.ProfilePictureId{
		Base: responses.Base{
			Success: true,
			Message: services.Tr("Successfully updated profile picture", lang),
		},
		ProfilePictureId: user.ProfilePictureId.String(),
	})
}

// /v1/user/:id/profilePicture?id=string&size=int (GET)
func (h *Handler) GetProfilePicture(c echo.Context) error {
	lang := c.Get("lang").(string)

	userId, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, responses.New(false, "Invalid or missing id parameter", lang))
	}

	user, err := h.userStore.GetById(userId)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, responses.NewUnexpectedError(err, lang))
	}
	if user == nil {
		return c.JSON(http.StatusNotFound, responses.NewNotFound(lang))
	}

	if c.QueryParam("id") != "" && c.QueryParam("id") != user.ProfilePictureId.String() {
		return c.JSON(http.StatusNotFound, responses.New(false, "Wrong profile picture id", lang))
	}

	size := config.Data.UserProfilePictureSize
	if c.QueryParam("size") != "" {
		size, err = strconv.Atoi(c.QueryParam("size"))
		if err != nil {
			return c.JSON(http.StatusBadRequest, responses.New(false, "The 'size' query parameter is not a number", lang))
		}

		if size > config.Data.UserProfilePictureSize {
			return c.JSON(http.StatusBadRequest, responses.New(false, fmt.Sprintf(services.Tr("Max allowed size is %d", lang), config.Data.UserProfilePictureSize), ""))
		}
	}

	profilePicture, err := h.userStore.GetProfilePicture(user)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, responses.NewUnexpectedError(err, lang))
	}
	if len(profilePicture) == 0 {
		return c.JSON(http.StatusNotFound, responses.New(false, "No profile picture set", lang))
	}

	data, err := bimg.NewImage(profilePicture).Thumbnail(size)

	return c.Blob(http.StatusOK, "image/jpeg", data)
}
