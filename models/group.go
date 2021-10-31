package models

import (
	"github.com/google/uuid"
)

type GroupStore interface {
	GetAllByUser(user *User, page, pageSize int, descending bool) ([]Group, error)
	GetById(id uuid.UUID) (*Group, error)
	Create(group *Group) error
	Update(group *Group) error
	Delete(group *Group) error
	DeleteById(id uuid.UUID) error

	GetGroupPicture(group *Group) ([]byte, error)

	GetMembers(except *User, group *Group, page, pageSize int, descending bool) ([]User, error)
	IsMember(group *Group, user *User) (bool, error)
	AddMember(group *Group, user *User) error
	RemoveMember(group *Group, user *User) error

	GetAdmins(except *User, group *Group, page, pageSize int, descending bool) ([]User, error)
	IsAdmin(group *Group, user *User) (bool, error)
	AddAdmin(group *Group, user *User) error
	RemoveAdmin(group *Group, user *User) error

	IsInGroup(group *Group, user *User) (bool, error)
	GetUserCount(group *Group) (int64, error)

	GetTransactionLog(group *Group, user *User, page, pageSize int, oldestFirst bool) ([]TransactionLogEntry, error)
	GetBankTransactionLog(group *Group, page, pageSize int, oldestFirst bool) ([]TransactionLogEntry, error)
	GetTransactionLogEntryById(group *Group, id uuid.UUID) (*TransactionLogEntry, error)
	GetLastTransactionLogEntry(group *Group, user *User) (*TransactionLogEntry, error)
	GetUserBalance(group *Group, user *User) (int, error)
	CreateTransaction(group *Group, senderIsBank, receiverIsBank bool, sender *User, receiver *User, title, description string, amount int) error
	CreateTransactionFromPaymentPlan(group *Group, senderIsBank, receiverIsBank bool, sender *User, receiver *User, title, description string, amount int, paymentPlanId uuid.UUID) error

	CreateInvitation(group *Group, user *User, message string) error
	GetInvitationById(id uuid.UUID) (*GroupInvitation, error)
	GetInvitationsByGroup(group *Group, page, pageSize int, oldestFirst bool) ([]GroupInvitation, error)
	GetInvitationsByUser(user *User, page, pageSize int, oldestFirst bool) ([]GroupInvitation, error)
	GetInvitationByGroupAndUser(group *Group, user *User) (*GroupInvitation, error)
	DeleteInvitation(invitation *GroupInvitation) error

	GetPaymentPlans(group *Group, user *User, page, pageSize int, descending bool) ([]PaymentPlan, error)
	GetBankPaymentPlans(group *Group, page, pageSize int, descending bool) ([]PaymentPlan, error)
	GetPaymentPlansThatNeedToBeExecuted() ([]PaymentPlan, error)
	GetPaymentPlanById(group *Group, id uuid.UUID) (*PaymentPlan, error)
	CreatePaymentPlan(group *Group, senderIsBank, receiverIsBank bool, sender *User, receiver *User, name, description string, amount, repeats, schedule int, scheduleUnit string, firstPayment int64) error
	UpdatePaymentPlan(paymentPlan *PaymentPlan) error
	DeletePaymentPlan(paymentPlan *PaymentPlan) error
}

type Group struct {
	Base
	Name           string
	Description    string
	GroupPicture   []byte
	GroupPictureId uuid.UUID `gorm:"type:uuid"`

	Memberships []GroupMembership
	Invitations []GroupInvitation
}

type GroupMembership struct {
	Base
	GroupName string
	GroupId   uuid.UUID `gorm:"type:uuid"`
	UserName  string
	UserId    uuid.UUID `gorm:"type:uuid"`
	IsMember  bool
	IsAdmin   bool
}

type GroupInvitation struct {
	Base
	Message string
	GroupId uuid.UUID `gorm:"type:uuid"`
	UserId  uuid.UUID `gorm:"type:uuid"`
}

type TransactionLogEntry struct {
	Base
	Title       string
	Description string
	Amount      int

	GroupId uuid.UUID `gorm:"type:uuid"`

	SenderIsBank            bool
	SenderId                uuid.UUID `gorm:"type:uuid"`
	NewBalanceSender        int
	BalanceDifferenceSender int

	ReceiverIsBank            bool
	ReceiverId                uuid.UUID `gorm:"type:uuid"`
	NewBalanceReceiver        int
	BalanceDifferenceReceiver int

	PaymentPlanId uuid.UUID `gorm:"type:uuid"`
}

const (
	ScheduleUnitDay   = "day"
	ScheduleUnitWeek  = "week"
	ScheduleUnitMonth = "month"
	ScheduleUnitYear  = "year"
)

type PaymentPlan struct {
	Base
	Name        string
	Description string

	Amount int

	// negative payment count for unlimited payments
	PaymentCount int

	NextExecute  int64
	Schedule     int
	ScheduleUnit string

	SenderIsBank bool
	SenderId     uuid.UUID `gorm:"type:uuid"`

	ReceiverIsBank bool
	ReceiverId     uuid.UUID `gorm:"type:uuid"`

	GroupId uuid.UUID `gorm:"type:uuid"`
}
