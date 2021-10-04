package services

import (
	"bytes"
	"crypto/rand"
	"encoding/json"
	"errors"
	"fmt"
	"image/png"
	"io"
	"log"
	"math/big"
	"net/http"
	"net/url"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/pquerna/otp/totp"
	"gitlab.com/Bananenpro05/hbank2-api/config"
	"gitlab.com/Bananenpro05/hbank2-api/models"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

var (
	bcryptCost          = 10
	hCaptchaVerifyUrl   = "https://hcaptcha.com/siteverify"
	emailCodeLifetime   = time.Minute.Milliseconds() * 5
	confirmEmailTimeout = time.Minute.Milliseconds() * 2
	loginTokenLifetime  = time.Minute.Milliseconds() * 5
)

func Register(ctx echo.Context, email, name, password string) (uuid.UUID, error) {
	db := dbFromCtx(ctx)

	if err := db.First(&models.User{}, "email = ?", email).Error; err != gorm.ErrRecordNotFound {
		return uuid.UUID{}, ErrEmailExists
	}

	user := models.User{
		Name:             name,
		Email:            email,
		ProfilePictureId: uuid.New(),
	}

	var err error
	user.PasswordHash, err = bcrypt.GenerateFromPassword([]byte(password), bcryptCost)
	if err != nil {
		return uuid.UUID{}, err
	}

	user.RecoveryCodes, err = generateRecoveryCodes()
	if err != nil {
		return uuid.UUID{}, err
	}

	db.Create(&user)

	return user.Id, nil
}

func VerifyCaptcha(token string) error {
	if config.Data.CaptchaEnabled {
		formValues := make(url.Values)
		formValues.Set("secret", config.Data.HCaptchaSecret)
		formValues.Set("response", token)
		formValues.Set("sitekey", config.Data.HCaptchaSiteKey)
		resp, err := http.PostForm(hCaptchaVerifyUrl, formValues)
		if err != nil {
			log.Printf("Failed to contact '%s': %s\n", hCaptchaVerifyUrl, err)
			return err
		}
		defer resp.Body.Close()

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			log.Println("Failed to read verify captcha response: ", err)
			return err
		}

		type Response struct {
			Success bool
		}
		var jsonResp Response
		json.Unmarshal(body, &jsonResp)

		if jsonResp.Success {
			return nil
		} else {
			return ErrInvalidCaptchaToken
		}
	}
	return nil
}

func SendConfirmEmail(ctx echo.Context, email string) error {
	db := dbFromCtx(ctx)

	var confirmEmailLastSent models.ConfirmEmailLastSent
	err := db.First(&confirmEmailLastSent, "email = ?", email).Error
	if err == gorm.ErrRecordNotFound {
		confirmEmailLastSent := models.ConfirmEmailLastSent{
			Email:    email,
			LastSent: time.Now().UnixMilli(),
		}
		db.Create(&confirmEmailLastSent)
	} else {
		if time.Now().UnixMilli()-confirmEmailLastSent.LastSent < confirmEmailTimeout {
			return ErrTimeout
		}
		confirmEmailLastSent.LastSent = time.Now().UnixMilli()
		db.Updates(&confirmEmailLastSent)
	}

	var user models.User
	err = db.Joins("EmailCode").First(&user, "email = ?", email).Error
	if err != nil {
		switch err {
		case gorm.ErrRecordNotFound:
			return ErrNotFound
		default:
			return err
		}
	}

	if !user.EmailConfirmed {
		code, err := generateRandomString(6)
		if err != nil {
			return err
		}

		db.Delete(&user.EmailCode)
		err = db.Model(&user).Association("EmailCode").Replace(&models.EmailCode{
			Code:           code,
			ExpirationTime: time.Now().UnixMilli() + emailCodeLifetime,
		})
		if err != nil {
			return err
		}

		if config.Data.EmailEnabled {
			type templateData struct {
				Name    string
				Content string
			}
			body, err := ParseEmailTemplate("template.html", templateData{
				Name:    user.Name,
				Content: "der Code lautet: " + user.EmailCode.Code,
			})
			if err != nil {
				return err
			}
			go SendEmail([]string{user.Email}, "H-Bank Bestätigungscode", body)
		}

		return nil
	} else {
		return ErrEmailAlreadyConfirmed
	}
}

func VerifyConfirmEmailCode(ctx echo.Context, email string, code string) bool {
	db := dbFromCtx(ctx)

	var user models.User
	err := db.Joins("EmailCode").First(&user, "email = ?", email).Error
	if err != nil {
		return false
	}

	success := false

	if user.EmailCode.Code == code {
		if user.EmailCode.ExpirationTime > time.Now().UnixMilli() {
			user.EmailConfirmed = true
			db.Model(&user).Select("email_confirmed").Updates(&user)

			success = true
		}

		db.Delete(&user.EmailCode)
	}

	return success
}

func Activate2FAOTP(ctx echo.Context, email, password string) ([]byte, error) {
	db := dbFromCtx(ctx)

	var user models.User
	err := db.First(&user, "email = ?", email).Error
	if err != nil {
		return []byte{}, ErrInvalidCredentials
	}

	if bcrypt.CompareHashAndPassword(user.PasswordHash, []byte(password)) != nil {
		return []byte{}, ErrInvalidCredentials
	}

	if !user.TwoFaOTPEnabled {
		key, err := totp.Generate(totp.GenerateOpts{
			Issuer:      config.Data.DomainName,
			AccountName: user.Email,
		})
		if err != nil {
			return []byte{}, errors.New("Failed to generate OTP")
		}

		img, err := key.Image(200, 200)
		if err != nil {
			return []byte{}, errors.New("Failed to generate OTP QR code")
		}

		var qr bytes.Buffer

		png.Encode(&qr, img)

		secret := key.Secret()

		user.OtpSecret = secret
		user.OtpQrCode = qr.Bytes()

		return qr.Bytes(), db.Model(&user).Select([]string{"two_fa_otp_enabled", "otp_secret", "otp_qr_code"}).Updates(&user).Error
	}

	return user.OtpQrCode, nil
}

func VerifyOTPCode(ctx echo.Context, email, code string) bool {
	db := dbFromCtx(ctx)
	var user models.User
	err := db.First(&user, "email = ?", email).Error
	if err != nil {
		return false
	}

	if totp.Validate(code, user.OtpSecret) {
		user.TwoFaOTPEnabled = true
		db.Model(&user).Select("two_fa_otp_enabled").Updates(&user)
		return true
	}

	return false
}

func Login(ctx echo.Context, email, password string) (string, error) {
	db := dbFromCtx(ctx)
	var user models.User
	err := db.First(&user, "email = ?", email).Error
	if err != nil {
		fmt.Println(err)
		return "", ErrInvalidCredentials
	}

	if bcrypt.CompareHashAndPassword(user.PasswordHash, []byte(password)) != nil {
		return "", ErrInvalidCredentials
	}

	var code string
	existingLoginToken := models.LoginToken{}
	for code == "" || db.First(&existingLoginToken, "code = ?", code).Error != gorm.ErrRecordNotFound {
		code, err = generateRandomString(64)
		if err != nil {
			return "", err
		}
	}

	user.LoginTokens = append(user.LoginTokens, models.LoginToken{
		Code:           code,
		ExpirationTime: time.Now().UnixMilli() + loginTokenLifetime,
	})
	db.Updates(&user)

	return code, nil
}

// ========== Helper functions ==========

func generateRecoveryCodes() ([]models.RecoveryCode, error) {
	codes := make([]models.RecoveryCode, 10)

	var err error
	for i := range codes {
		codes[i].Code, err = generateRandomString(64)
		if err != nil {
			return []models.RecoveryCode{}, errors.New("Couldn't generate recovery codes")
		}
	}

	return codes, nil
}

func dbFromCtx(ctx echo.Context) *gorm.DB {
	return ctx.Get(models.DBContextKey).(*gorm.DB)
}

func init() {
	assertAvailablePRNG()
}

func assertAvailablePRNG() {
	buf := make([]byte, 1)

	_, err := io.ReadFull(rand.Reader, buf)
	if err != nil {
		panic(fmt.Sprintf("crypto/rand is unavailable: Read() failed with %#v", err))
	}
}

func generateRandomBytes(length int) ([]byte, error) {
	b := make([]byte, length)
	_, err := rand.Read(b)
	if err != nil {
		return nil, err
	}

	return b, nil
}

func generateRandomString(length int) (string, error) {
	const letters = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"
	ret := make([]byte, length)
	for i := 0; i < length; i++ {
		num, err := rand.Int(rand.Reader, big.NewInt(int64(len(letters))))
		if err != nil {
			return "", err
		}
		ret[i] = letters[num.Int64()]
	}

	return string(ret), nil
}
