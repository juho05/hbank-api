package handlers

import (
	"github.com/Bananenpro/hbank-api/router/middlewares"
	"github.com/labstack/echo/v4"
)

func (h *Handler) RegisterApi(api *echo.Group) {
	api.GET("/status", h.Status)

	auth := api.Group("/auth")
	auth.POST("/register", h.Register)
	auth.GET("/confirmEmail/:email", h.SendConfirmEmail)
	auth.POST("/confirmEmail", h.VerifyConfirmEmailCode)
	auth.POST("/passwordAuth", h.PasswordAuth)
	auth.POST("/login", h.Login)
	auth.POST("/refresh", h.Refresh)
	auth.POST("/logout", h.Logout, middlewares.JWT)
	auth.POST("/changePassword", h.ChangePassword, middlewares.JWT)
	auth.POST("/forgotPassword", h.ForgotPassword)
	auth.POST("/resetPassword", h.ResetPassword)
	auth.POST("/requestChangeEmail", h.RequestChangeEmail, middlewares.JWT)
	auth.POST("/changeEmail", h.ChangeEmail, middlewares.JWT)

	twoFactor := auth.Group("/twoFactor")
	twoFactor.POST("/otp/activate", h.Activate2FAOTP)
	twoFactor.POST("/otp/qr", h.GetOTPQRCode)
	twoFactor.POST("/otp/key", h.GetOTPKey)
	twoFactor.POST("/otp/verify", h.VerifyOTPCode)
	twoFactor.POST("/otp/new", h.NewOTP, middlewares.JWT)

	twoFactor.POST("/recovery/verify", h.VerifyRecoveryCode)
	twoFactor.POST("/recovery/new", h.NewRecoveryCodes, middlewares.JWT)

	api.GET("/user", h.GetUsers, middlewares.JWT)
	api.GET("/user/:id", h.GetUser, middlewares.JWT)
	api.PUT("/user", h.UpdateUser, middlewares.JWT)
	api.DELETE("/user/:id", h.DeleteUserByDeleteToken)
	api.POST("/user/delete", h.DeleteUser, middlewares.JWT)
	api.POST("/user/picture", h.SetProfilePicture, middlewares.JWT)
	api.DELETE("/user/picture", h.RemoveProfilePicture, middlewares.JWT)
	api.GET("/user/:id/picture", h.GetProfilePicture, middlewares.JWT)

	user := api.Group("/user")

	user.GET("/cash/current", h.GetCurrentCash, middlewares.JWT)
	user.GET("/cash/:id", h.GetCashLogEntryById, middlewares.JWT)
	user.GET("/cash", h.GetCashLog, middlewares.JWT)
	user.POST("/cash", h.AddCashLogEntry, middlewares.JWT)

	api.GET("/group", h.GetGroups, middlewares.JWT)
	api.GET("/group/:id", h.GetGroupById, middlewares.JWT)
	api.POST("/group", h.CreateGroup, middlewares.JWT)
	api.PUT("/group/:id", h.UpdateGroup, middlewares.JWT)

	group := api.Group("/group")
	group.GET("/:id/member", h.GetGroupMembers, middlewares.JWT)
	group.DELETE("/:id/member", h.LeaveGroup, middlewares.JWT)
	group.GET("/:id/admin", h.GetGroupAdmins, middlewares.JWT)
	group.POST("/:id/admin", h.AddGroupAdmin, middlewares.JWT)
	group.DELETE("/:id/admin", h.RemoveAdminRights, middlewares.JWT)
	group.GET("/:id/user", h.GetGroupUsers, middlewares.JWT)
	group.GET("/:id/picture", h.GetGroupPicture, middlewares.JWT)
	group.POST("/:id/picture", h.SetGroupPicture, middlewares.JWT)
	group.DELETE("/:id/picture", h.RemoveGroupPicture, middlewares.JWT)

	group.GET("/:id/transaction/balance", h.GetBalance, middlewares.JWT)
	group.GET("/:id/transaction/:transactionId", h.GetTransactionById, middlewares.JWT)
	group.GET("/:id/transaction", h.GetTransactionLog, middlewares.JWT)
	group.POST("/:id/transaction", h.CreateTransaction, middlewares.JWT)

	group.GET("/:id/invitation", h.GetInvitationsByGroup, middlewares.JWT)
	group.GET("/invitation", h.GetInvitationsByUser, middlewares.JWT)
	group.GET("/invitation/:id", h.GetInvitationById, middlewares.JWT)
	group.POST("/:id/invitation", h.CreateInvitation, middlewares.JWT)
	group.POST("/invitation/:id", h.AcceptInvitation, middlewares.JWT)
	group.DELETE("/invitation/:id", h.DenyInvitation, middlewares.JWT)

	group.GET("/:id/paymentPlan/:paymentPlanId", h.GetPaymentPlanById, middlewares.JWT)
	group.GET("/:id/paymentPlan", h.GetPaymentPlans, middlewares.JWT)
	group.GET("/:id/paymentPlan/nextPayment", h.GetPaymentPlanNextPayments, middlewares.JWT)
	group.POST("/:id/paymentPlan", h.CreatePaymentPlan, middlewares.JWT)
	group.PUT("/:id/paymentPlan/:paymentPlanId", h.UpdatePaymentPlan, middlewares.JWT)
	group.DELETE("/:id/paymentPlan/:paymentPlanId", h.DeletePaymentPlan, middlewares.JWT)
}
