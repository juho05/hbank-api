package models

import (
	"github.com/google/uuid"
)

type UserStore interface {
	GetAll(except *User, page, pageSize int, descending bool) ([]User, error)
	GetById(id uuid.UUID) (*User, error)
	GetByEmail(email string) (*User, error)
	Create(user *User) error
	Update(user *User) error
	Delete(user *User) error
	DeleteById(id uuid.UUID) error
	DeleteByEmail(email string) error

	GetProfilePicture(user *User) ([]byte, error)

	GetCashLog(user *User, page, pageSize int, oldestFirst bool) ([]CashLogEntry, error)
	GetLastCashLogEntry(user *User) (*CashLogEntry, error)
	GetCashLogEntryById(user *User, id uuid.UUID) (*CashLogEntry, error)
	AddCashLogEntry(user *User, entry *CashLogEntry) error

	GetConfirmEmailCode(user *User) (*ConfirmEmailCode, error)
	DeleteConfirmEmailCode(code *ConfirmEmailCode) error

	GetResetPasswordCode(user *User) (*ResetPasswordCode, error)
	DeleteResetPasswordCode(code *ResetPasswordCode) error

	GetChangeEmailCode(user *User) (*ChangeEmailCode, error)
	DeleteChangeEmailCode(code *ChangeEmailCode) error

	GetRefreshToken(user *User, id uuid.UUID) (*RefreshToken, error)
	AddRefreshToken(user *User, refreshToken *RefreshToken) error
	RotateRefreshToken(user *User, oldRefreshToken *RefreshToken) (*RefreshToken, string, error)
	DeleteRefreshToken(refreshToken *RefreshToken) error
	DeleteRefreshTokens(user *User) error

	GetPasswordTokenByCode(user *User, code string) (*PasswordToken, error)
	GetPasswordTokens(user *User) ([]PasswordToken, error)
	DeletePasswordToken(token *PasswordToken) error

	GetTwoFATokenByCode(user *User, code string) (*TwoFAToken, error)
	GetTwoFATokens(user *User) ([]TwoFAToken, error)
	DeleteTwoFAToken(token *TwoFAToken) error

	GetRecoveryCodeByCode(user *User, code string) (*RecoveryCode, error)
	GetRecoveryCodes(user *User) ([]RecoveryCode, error)
	NewRecoveryCodes(user *User) ([]string, error)
	DeleteRecoveryCode(code *RecoveryCode) error

	GetConfirmEmailLastSent(email string) (int64, error)
	SetConfirmEmailLastSent(email string, time int64) error

	GetForgotPasswordEmailLastSent(email string) (int64, error)
	SetForgotPasswordEmailLastSent(email string, time int64) error
}

const (
	ProfilePictureEverybody = "everybody"
	ProfilePictureNobody    = "nobody"
)

type User struct {
	Base
	Name                  string
	Email                 string `gorm:"unique"`
	PasswordHash          []byte
	ProfilePicture        []byte
	ProfilePictureId      uuid.UUID `gorm:"type:uuid"`
	ProfilePicturePrivacy string    `gorm:"default:everybody"`
	CashLog               []CashLogEntry
	EmailConfirmed        bool
	TwoFaOTPEnabled       bool
	OtpSecret             string
	OtpQrCode             []byte
	ConfirmEmailCode      ConfirmEmailCode  `gorm:"constraint:OnDelete:CASCADE"`
	ResetPasswordCode     ResetPasswordCode `gorm:"constraint:OnDelete:CASCADE"`
	ChangeEmailCode       ChangeEmailCode   `gorm:"constraint:OnDelete:CASCADE"`
	RefreshTokens         []RefreshToken
	PasswordTokens        []PasswordToken
	TwoFATokens           []TwoFAToken
	RecoveryCodes         []RecoveryCode
	MemberGroups          []Group `gorm:"many2many:group_members"`
	AdminGroups           []Group `gorm:"many2many:group_admins"`
}

type CashLogEntry struct {
	Base
	ChangeTitle       string
	ChangeDescription string
	TotalAmount       int
	ChangeDifference  int

	Ct1  int
	Ct2  int
	Ct5  int
	Ct10 int
	Ct20 int
	Ct50 int

	Eur1   int
	Eur2   int
	Eur5   int
	Eur10  int
	Eur20  int
	Eur50  int
	Eur100 int
	Eur200 int
	Eur500 int

	UserId uuid.UUID `gorm:"type:uuid"`
}

type ConfirmEmailLastSent struct {
	Base
	Email    string `gorm:"unique"`
	LastSent int64
}

type ConfirmEmailCode struct {
	Base
	CodeHash []byte
	UserId   uuid.UUID `gorm:"type:uuid"`
}

type ResetPasswordCode struct {
	Base
	CodeHash       []byte
	ExpirationTime int64
	UserId         uuid.UUID `gorm:"type:uuid"`
}

type ChangeEmailCode struct {
	Base
	CodeHash       []byte
	ExpirationTime int64
	UserId         uuid.UUID `gorm:"type:uuid"`
	NewEmail       string
}

type RefreshToken struct {
	Base
	CodeHash       []byte
	ExpirationTime int64
	Used           bool
	UserId         uuid.UUID `gorm:"type:uuid"`
}

type PasswordToken struct {
	Base
	CodeHash       []byte
	ExpirationTime int64
	UserId         uuid.UUID `gorm:"type:uuid"`
}

type TwoFAToken struct {
	Base
	CodeHash       []byte
	ExpirationTime int64
	UserId         uuid.UUID `gorm:"type:uuid"`
}

type RecoveryCode struct {
	Base
	CodeHash []byte
	UserId   uuid.UUID `gorm:"type:uuid"`
}

type ForgotPasswordEmailLastSent struct {
	Base
	Email    string `gorm:"unique"`
	LastSent int64
}
