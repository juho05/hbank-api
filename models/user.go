package models

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type User struct {
	Base
	Name             string
	Email            string `gorm:"unique"`
	PasswordHash     []byte
	ProfilePicture   []byte
	ProfilePictureId uuid.UUID `gorm:"type:uuid"`
	EmailConfirmed   bool
	TwoFaOTPEnabled  bool
	OtpSecret        string
	OtpQrCode        []byte
	EmailCode        EmailCode `gorm:"constraint:OnDelete:CASCADE"`
	TokenKey         []byte
	RefreshTokens    []RefreshToken `gorm:"constraint:OnDelete:CASCADE"`
	LoginTokens      []LoginToken   `gorm:"constraint:OnDelete:CASCADE"`
	RecoveryCodes    []RecoveryCode `gorm:"constraint:OnDelete:CASCADE"`
}

type ConfirmEmailLastSent struct {
	Base
	Email    string `gorm:"unique"`
	LastSent int64
}

func userAutoMigrate(db *gorm.DB) (errs []error) {
	err := db.AutoMigrate(User{})
	if err != nil {
		errs = append(errs, err)
	}

	err = db.AutoMigrate(ConfirmEmailLastSent{})
	if err != nil {
		errs = append(errs, err)
	}

	return
}
